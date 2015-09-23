package agent

import (
	"fmt"
	"strings"
	"sync"
)

type Logger interface {
	Error(string, ...interface{})
	Fatal(string, ...interface{})
	Warn(string, ...interface{})
	Info(string, ...interface{})
	Debug(string, ...interface{})
}

type LogLevel int32

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	FATAL
)

// LogDev
type LogDev struct {
	logger Logger
	level  LogLevel
	mut    sync.RWMutex
}

type LoggerConf struct {
	LogFile  string 
	LogLevel string
}

var LogConfig LoggerConf

func LogInit(logFile ,myLevel string) {

	LogConfig.LogFile = logFile
	LogConfig.LogLevel = myLevel
	
	fmt.Println("INFO: Logger(%s,%s)", LogConfig.LogFile, LogConfig.LogLevel)

	var logger Logger
	logger = NewFileLogger(LogConfig.LogFile)
	SetLogger(logger)
	SetLevel(GetLevelStr(LogConfig.LogLevel))
}

func (t *LogDev) GetLevelStr(level string) LogLevel {
	switch strings.ToLower(level) {
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn":
		return WARN
	case "warning":
		return WARN
	case "error":
		return WARN
	case "fatal":
		return WARN
	default:
		fmt.Printf("ERROR: logger level unknown: %v\n", level)
		return INFO
	}
}

// SetLogger defines a new logger.
func (t *LogDev) SetLogger(l Logger) {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.logger = l
}

// SetLogger defines a new logger.
func (t *LogDev) SetLevel(l LogLevel) {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.level = l
}

// Error
func (t *LogDev) Error(format string, v ...interface{}) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.logger != nil {
		t.logger.Error(format, v...)
	}
}

// Fatal
func (t *LogDev) Fatal(format string, v ...interface{}) {
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.logger != nil {
		t.logger.Fatal(format, v...)
	}
}

// Warn
func (t *LogDev) Warn(format string, v ...interface{}) {
	if t.level > WARN {
		return
	}
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.logger != nil {
		t.logger.Debug(format, v...)
	}
}

// Debug
func (t *LogDev) Debug(format string, v ...interface{}) {
	if t.level > DEBUG {
		return
	}
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.logger != nil {
		t.logger.Debug(format, v...)
	}
}

// Info
func (t *LogDev) Info(format string, v ...interface{}) {
	if t.level > INFO {
		return
	}
	t.mut.RLock()
	defer t.mut.RUnlock()
	if t.logger != nil {
		t.logger.Debug(format, v...)
	}
}

// GetStdLogger returns a standard Logger instance
//func (t *LogDev) GetStdLogger() *log.Logger {
//	t.mut.RLock()
//	defer t.mut.RUnlock()
//	if t.logger != nil {
//		return t.logger.GetStdLogger()
//	}
//	return nil
//}

var DefaultLogDev = new(LogDev)

// Error
func Error(format string, v ...interface{}) {
	DefaultLogDev.Error(format, v...)
}

// Fatal
func Fatal(format string, v ...interface{}) {
	DefaultLogDev.Fatal(format, v...)
}

// Warn
func Warn(format string, v ...interface{}) {
	DefaultLogDev.Warn(format, v...)
}

// Debug
func Debug(format string, v ...interface{}) {
	DefaultLogDev.Debug(format, v...)
}

//Info
func Info(format string, v ...interface{}) {
	DefaultLogDev.Info(format, v...)
}

// GetStdLogger is a wrapper.
//func GetStdLogger() *log.Logger {
//	return DefaultLogDev.GetStdLogger()
//}

// SetLogger.
func SetLogger(logger Logger) {
	DefaultLogDev.SetLogger(logger)
}

// SetLevel.
func SetLevel(lv LogLevel) {
	DefaultLogDev.SetLevel(lv)
}

// GetLevel.
func GetLevelStr(lvs string) LogLevel {
	return DefaultLogDev.GetLevelStr(lvs)
}
