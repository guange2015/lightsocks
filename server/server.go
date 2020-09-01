package server

import (
	"encoding/binary"
	"github.com/guange2015/lightsocks/core"
	"log"
	"net"
)

type LsServer struct {
	*core.SecureSocket
}

// 新建一个服务端
// 服务端的职责是:
// 1. 监听来自本地代理客户端的请求
// 2. 解密本地代理客户端请求的数据，解析 SOCKS5 协议，连接用户浏览器真正想要连接的远程服务器
// 3. 转发用户浏览器真正想要连接的远程服务器返回的数据的加密后的内容到本地代理客户端
func New(password *core.Password, listenAddr *net.TCPAddr) *LsServer {
	return &LsServer{
		SecureSocket: &core.SecureSocket{
			Cipher:     core.NewCipher(password),
			ListenAddr: listenAddr,
		},
	}
}

// 运行服务端并且监听来自本地代理客户端的请求
func (lsServer *LsServer) Listen(didListen func(listenAddr net.Addr)) error {
	listener, err := net.ListenTCP("tcp", lsServer.ListenAddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	if didListen != nil {
		didListen(listener.Addr())
	}

	for {
		localConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		// localConn被关闭时直接清除所有数据 不管没有发送的数据
		localConn.SetLinger(0)
		go lsServer.handleConn(localConn)
	}
	return nil
}

// 解 SOCKS5 协议
// https://www.ietf.org/rfc/rfc1928.txt
func (lsServer *LsServer) handleConn(localConn *net.TCPConn) {
	defer localConn.Close()

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	atypeBuf := make([]byte, 1)
	_, err := lsServer.DecodeReadFull(localConn, atypeBuf)
	// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
	if err != nil {
		log.Println("read atype error: ", err)
		return
	}

	var dIP []byte
	// aType 代表请求的远程服务器地址类型，值长度1个字节，有三种类型
	switch atypeBuf[0] {
	case 0x01:
		//	IP V4 address: X'01'
		ipv4Buf := make([]byte, net.IPv4len)
		_, err := lsServer.DecodeReadFull(localConn, ipv4Buf)
		// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
		if err != nil {
			log.Println("read ipv4 error: ", err)
			return
		}
		dIP = ipv4Buf
		log.Println("connect ipv4: ", ipv4Buf)
	case 0x03:
		//	DOMAINNAME: X'03'
		domainLenBuf := make([]byte, 1)
		_, err := lsServer.DecodeReadFull(localConn, domainLenBuf)
		if err != nil {
			log.Println("read domain len error: ", err)
			return
		}
		if domainLenBuf[0] <= 0 || domainLenBuf[0] > 255 {
			log.Println("domain len error: ", domainLenBuf[0])
			return
		}

		domainBuf := make([]byte, int(domainLenBuf[0]))
		_, err = lsServer.DecodeReadFull(localConn, domainBuf)
		if err != nil {
			log.Println("read domain error: ", err)
			return
		}

		ipAddr, err := net.ResolveIPAddr("ip", string(domainBuf))
		if err != nil {
			return
		}
		dIP = ipAddr.IP

		log.Println("connect domain: ", string(domainBuf))
	case 0x04:
		//	IP V6 address: X'04'
		ipv6Buf := make([]byte, net.IPv6len)
		_, err := lsServer.DecodeReadFull(localConn, ipv6Buf)
		// n 最短的长度为7 情况为 ATYP=3 DST.ADDR占用1字节 值为0x0
		if err != nil {
			log.Println("read ipv6 error: ", err)
			return
		}
		dIP = ipv6Buf
		log.Println("connect ipv6: ", ipv6Buf)
	default:
		log.Println("unkown atype ", atypeBuf[0])
		return
	}

	portBuf := make([]byte, 2)
	_, err = lsServer.DecodeReadFull(localConn, portBuf)
	if err != nil {
		log.Println("read port error: ", err)
		return
	}

	port := int(binary.BigEndian.Uint16(portBuf))
	dstAddr := &net.TCPAddr{
		IP:   dIP,
		Port: port,
	}

	log.Println("start connect ip: ", dIP, " port: ", port)

	// 连接真正的远程服务
	dstServer, err := net.DialTCP("tcp", nil, dstAddr)
	if err != nil {
		log.Println(err)
		return
	} else {
		defer dstServer.Close()
		// Conn被关闭时直接清除所有数据 不管没有发送的数据
		dstServer.SetLinger(0)

		// 响应客户端连接成功
		/**
		  +----+-----+-------+------+----------+----------+
		  |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		  +----+-----+-------+------+----------+----------+
		  | 1  |  1  | X'00' |  1   | Variable |    2     |
		  +----+-----+-------+------+----------+----------+
		*/
		// 响应客户端连接成功
		lsServer.EncodeWrite(localConn, []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	}

	close_chan := make(chan int)

	// 进行转发
	// 从 localUser 读取数据发送到 dstServer
	go func() {
		_ = lsServer.DecodeCopy(dstServer, localConn)
		close_chan <- 1

	}()

	// 从 dstServer 读取数据发送到 localUser，这里因为处在翻墙阶段出现网络错误的概率更大
	go func() {
		_ = lsServer.EncodeCopy(localConn, dstServer)
		close_chan <- 1
	}()

	<-close_chan

	log.Println("close ip: ", dIP, " port: ", port)
}
