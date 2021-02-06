package bridge

import (
	"github.com/guange2015/lightsocks/cmd"
	"io"
	"log"
	"net"
)


// 流程说明
// 1. 监听tcp
// 2. 客户端连上来
// 3. 连上服务器
// 4. 交换两边的消息

func Run( config *cmd.Config ) error {
	listenAddr, err := net.ResolveTCPAddr("tcp", config.ListenAddr)
	if err != nil {
		return err
	}

	log.Printf("成功监听于 %v \n", config.ListenAddr)

	remoteAddr, err := net.ResolveTCPAddr("tcp", config.RemoteAddr)
	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()


	for {
		localConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}

		log.Printf("客户端连接 %v \n", localConn.LocalAddr())

		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		localConn.SetLinger(0)
		go handleConn(localConn, remoteAddr)
	}
	return nil


}

func handleConn(localConn *net.TCPConn, remoteAddr *net.TCPAddr) error {
	defer localConn.Close()

	//连接远程服务器

	remoteConn, err := net.DialTCP("tcp", nil, remoteAddr)
	if err!=nil {
		log.Printf("连接远程服务器出错: %v", err)
		return err
	}

	log.Printf("连接远程服务器成功")

	defer remoteConn.Close()

	return copySocket(remoteConn, localConn)
}

func copySocket( conn1 *net.TCPConn, conn2 *net.TCPConn) error  {
	log.Println("开始转发数据")
	close_chan := make(chan int)
	go func() {
		_, err := io.Copy(conn1, conn2)
		if err!=nil{
			log.Println(err)
		}
		close_chan <- 1
	}()

	go func() {
		_, err := io.Copy(conn2, conn1)
		if err!=nil{
			log.Println(err)
		}
		close_chan <- 1
	}()

	<-close_chan

	log.Println("退出转发数据")

	return nil
}