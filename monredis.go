package main

import (
	"flag"
	"fmt"
	
	"sandswind/monredis/agent"
)

const (
	VERSION = "0.0.1"
)


var (
	cfgFile = flag.String("c", "config.cfg", "configuration file")
	verbose  = flag.Bool("v", true, "Log generic info")
)

func init() {
	
	flag.Parse()
	
	agent.SettingsLoad(*cfgFile, *verbose)
}



func main() {
	fmt.Printf("now running... ...")
	agent.Run()
}
