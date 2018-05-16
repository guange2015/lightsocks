package cmd

import (
	"flag"
	"os"
)

//解析参数，获取配置文件
func ParseCmd() string {
	configPath := flag.String("c", "config.json", "config path")
	flag.Parse()

	if configPath == nil || len(*configPath)<=0{
		flag.PrintDefaults()
		os.Exit(-1)
	}

	return *configPath
}