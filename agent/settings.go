package agent

import (
	"os"
	"io/ioutil"
	"fmt"
	"reflect"
	
	"github.com/BurntSushi/toml"
)

type MonSettings struct {
	Verbose             bool
	SentinelIp          string `toml:"sentinel_ip"`
	SentinelPort        string `toml:"sentinel_port"`
	HaproxyAddr      	string `toml:"haproxy_url"`
	LogFile             string `toml:"log_file"`
}

var Settings MonSettings = MonSettings{}

func ValidateSettings() {
	if  Settings.SentinelIp     == "" || 
		Settings.SentinelPort   == "" || 
		Settings.HaproxyAddr    == "" {
		fmt.Printf("parameters error:begin")
		PrintObject(Settings)
		fmt.Printf("parameters error:end")
		os.Exit(1)
	}
	
	PrintObject(Settings)
}


func ReadConfig(path string, obj interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("Read config|%s|%v", path, err)
		return err
	}
	
	if _, err := toml.Decode(string(data), obj); err != nil {
		fmt.Printf("toml.Decode|%v|%v", data, err)
		return err
	}
	return nil
}

func PrintObject(o interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("ObjectToString|Panic|%v", err)
		}
	}()
	v := reflect.ValueOf(o)
	for i := 0; i < v.NumField(); i++ {
		fmt.Println("Object ", v.Type().Field(i).Name, v.Field(i).Interface())
	}
}

func SettingsLoad(cfg string, verbose bool) {
	
	
	
	Settings.Verbose = verbose
	ReadConfig(cfg, &Settings)

	if len(Settings.LogFile) == 0 {
		Settings.LogFile = "monredis"
	}
	
	logStr := "info"
	if verbose {
		logStr = "debug"
	}
	
	LogInit(Settings.LogFile, logStr)
	ValidateSettings()
	Info("Loading configure done.")
}
