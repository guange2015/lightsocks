package main

import (
	"flag"
	"fmt"
	"github.com/guange2015/lightsocks/cmd"
	"github.com/guange2015/lightsocks/core"
	"github.com/guange2015/lightsocks/local"
	"github.com/guange2015/lightsocks/server"
	"github.com/phayes/freeport"
	"log"
	"net"
	"strings"
)

var version = "master"

var (
	runMode    string
	configPath string
)

func main() {
	var h bool
	flag.StringVar(&runMode, "m", "client", "run mode: server or client")
	flag.StringVar(&configPath, "c", "data/config.json", "config file path")
	flag.BoolVar(&h, "h", false, "help for this")

	flag.Parse()

	if h {
		flag.Usage()
		return
	}

	log.Println(runMode)

	if strings.Compare(runMode, "server") == 0 {
		log.SetFlags(log.Lshortfile)

		err, config := cmd.ReadConfig(configPath)
		if err != nil {

			log.Println("开始重新生成配置")
			//重新生成密码，保存配置
			// 服务端监听端口随机生成
			port, err := freeport.GetFreePort()
			if err != nil {
				// 随机端口失败就采用 7448
				port = 7448
			}
			// 默认配置
			config = &cmd.Config{
				ListenAddr: fmt.Sprintf(":%d", port),
				// 密码随机生成
				Password: core.RandPassword().String(),
			}
			err = cmd.SaveConfig(configPath, config)
			if err != nil {
				return
			}
		}

		// 解析配置
		password, err := core.ParsePassword(config.Password)
		if err != nil {
			log.Fatalln(err)
		}
		listenAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
		if err != nil {
			log.Fatalln(err)
		}

		// 启动 server 端并监听
		lsServer := server.New(password, listenAddr)
		log.Fatalln(lsServer.Listen(func(listenAddr net.Addr) {
			log.Println("使用配置：", fmt.Sprintf(`
本地监听地址 listen：
%s
密码 password：
%s
	`, listenAddr, password))
			log.Printf("lightsocks-server:%s 启动成功 监听在 %s\n", version, listenAddr.String())
		}))
	} else if strings.Compare(runMode, "client") == 0 {
		log.SetFlags(log.Lshortfile)

		err, config := cmd.ReadConfig(configPath)
		if err != nil {
			log.Fatalf("读取配置文件失败: %v", err)
		}

		// 解析配置
		password, err := core.ParsePassword(config.Password)
		if err != nil {
			log.Fatalln(err)
		}
		listenAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
		if err != nil {
			log.Fatalln(err)
		}
		remoteAddr, err := net.ResolveTCPAddr("tcp", config.RemoteAddr)
		if err != nil {
			log.Fatalln(err)
		}

		go local.StartProxy(config)
		go local.StartWebServer(config)

		// 启动 local 端并监听
		lsLocal := local.New(password, listenAddr, remoteAddr)
		log.Fatalln(lsLocal.Listen(func(listenAddr net.Addr) {
			log.Println("使用配置：", fmt.Sprintf(`
本地监听地址 listen：
%s
远程服务地址 remote：
%s
密码 password：
%s
	`, listenAddr, remoteAddr, password))
			log.Printf("lightsocks-local:%s 启动成功 监听在 %s\n", version, listenAddr.String())
		}))
	} else {
		flag.Usage()
	}

}
