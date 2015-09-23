package agent

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	errorPrefix = "E "
	fatalPrefix = "F "
	infoPrefix  = "I "
	warnPrefix  = "W "
	debugPrefix = "D "
)

type fileLogger struct {
	filename        string
	logfd           *os.File
	openday         int
	bufferList      *buffer
	bufferListMutex sync.Mutex
}

func NewFileLogger(fileName string) Logger {
	newfile := fmt.Sprintf("%s.%s.log", fileName, time.Now().Format("20060102"))
	newfd, err := os.OpenFile(newfile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: FileOpen(%s): %v\n", newfile, err)
		panic("log")
	}
	nowday := time.Now().Day()
	fmt.Fprintf(os.Stderr, "INFO: FileOpen(%s)\n", newfile)
	return &fileLogger{logfd: newfd, filename: fileName, openday: nowday}
}

func (l *fileLogger) Error(format string, o ...interface{}) {
	l.WriterMsgfmt(ERROR, fmt.Sprintf(format, o...))
}

func (l *fileLogger) Fatal(format string, o ...interface{}) {
	l.WriterMsgfmt(FATAL, fmt.Sprintf(format, o...))
	os.Exit(1)
}

func (l *fileLogger) Info(format string, o ...interface{}) {
	l.WriterMsgfmt(INFO, fmt.Sprintf(format, o...))
}

func (l *fileLogger) Warn(format string, o ...interface{}) {
	l.WriterMsgfmt(WARN, fmt.Sprintf(format, o...))
}

func (l *fileLogger) Debug(format string, o ...interface{}) {
	l.WriterMsgfmt(DEBUG, fmt.Sprintf(format, o...))
}

func (l *fileLogger) DoCheck() {
	if time.Now().Day() != l.openday {
		if err := l.DoRotate(); err != nil {
			fmt.Fprintf(os.Stderr, "FileLogChange: %v\n", err)
			return
		}
	}
}

func (l *fileLogger) DoRotate() error {
	fname := fmt.Sprintf("%s.%s.log", l.filename, time.Now().Format("20060102"))
	l.logfd.Close()
	filefd, err := os.OpenFile(fname, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FileOpen(%s): %s\n", l.filename, err)
		return err
	}
	nowday := time.Now().Day()
	flogger := fileLogger{logfd: filefd, filename: l.filename, openday: nowday}
	DefaultLogDev.logger = &flogger
	return nil
}

// Creating a bytes.Buffer is expensive so we will re-use existing ones.
type buffer struct {
	bytes.Buffer
	next *buffer
}

// getBuffer returns a new, ready-to-use buffer.
func (l *fileLogger) getBuffer() *buffer {
	l.bufferListMutex.Lock()
	b := l.bufferList
	if b != nil {
		l.bufferList = b.next
	}
	l.bufferListMutex.Unlock()
	if b == nil {
		b = new(buffer)
	} else {
		b.next = nil
		b.Reset()
	}
	return b
}

// putBuffer returns a buffer to the list.
func (l *fileLogger) putBuffer(b *buffer) {
	if b.Len() >= 1024 {
		// Let big buffers die a natural death.
		return
	}
	l.bufferListMutex.Lock()
	b.next = l.bufferList
	l.bufferList = b
	l.bufferListMutex.Unlock()
}

func (l *fileLogger) addTimeFmt(buffer *bytes.Buffer) {
	now := time.Now()
	year, month, day := now.Date()
	hour, minute, second := now.Clock()
	buffer.WriteString(fmt.Sprintf("%4d/%02d/%02d %02d:%02d:%02d.%03d ", year, month, day, hour, minute, second, now.Nanosecond()/1000000))
}

func (l *fileLogger) Formatter(buffer *bytes.Buffer, level LogLevel, args ...interface{}) {
	if level == DEBUG {
		buffer.WriteString(debugPrefix)
	} else if level == INFO {
		buffer.WriteString(infoPrefix)
	} else if level == WARN {
		buffer.WriteString(warnPrefix)
	} else if level == ERROR {
		buffer.WriteString(errorPrefix)
	} else if level == FATAL {
		buffer.WriteString(fatalPrefix)
	} else {
		buffer.WriteString(infoPrefix)
	}
	fmt.Fprintln(buffer, args...)
}

// destroy file logger.
func (l *fileLogger) Destroy() {
	l.logfd.Close()
}

func (l *fileLogger) WriterMsgfmt(level LogLevel, args ...interface{}) {
	l.DoCheck()
	buffer := l.getBuffer()
	l.addTimeFmt(&buffer.Buffer)
	pc, file, line, ok := runtime.Caller(4)
	if ok {
		// Get caller function name.
		fn := runtime.FuncForPC(pc)
		var fnName string
		if fn == nil {
			fnName = "?()"
		} else {
			fnName = strings.TrimLeft(filepath.Ext(fn.Name()), ".") + "()"
		}
		buffer.Buffer.WriteString(fmt.Sprintf("%s:%05d %s ", filepath.Base(file), line, fnName))
	}
	l.Formatter(&buffer.Buffer, level, args...)
	//l.logger.Print(string(buffer.Bytes()))
	l.logfd.Write(buffer.Bytes())
	l.putBuffer(buffer)
}
