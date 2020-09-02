package local

import (
	"bytes"
	"github.com/guange2015/lightsocks/core"
	"io"
	"log"
	"net"
)

type LsLocal struct {
	*core.SecureSocket
}

// 新建一个本地端
// 本地端的职责是:
// 1. 监听来自本机浏览器的代理请求
// 2. 转发前加密数据
// 3. 转发socket数据到墙外代理服务端
// 4. 把服务端返回的数据转发给用户的浏览器
func New(password *core.Password, listenAddr, remoteAddr *net.TCPAddr) *LsLocal {
	return &LsLocal{
		SecureSocket: &core.SecureSocket{
			Cipher:     core.NewCipher(password),
			ListenAddr: listenAddr,
			RemoteAddr: remoteAddr,
		},
	}
}

// 本地端启动监听，接收来自本机浏览器的连接
func (local *LsLocal) Listen(didListen func(listenAddr net.Addr)) error {
	listener, err := net.ListenTCP("tcp", local.ListenAddr)
	if err != nil {
		return err
	}

	defer listener.Close()

	if didListen != nil {
		didListen(listener.Addr())
	}

	for {
		userConn, err := listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		// userConn被关闭时直接清除所有数据 不管没有发送的数据
		userConn.SetLinger(0)
		go local.handleConn(userConn)
	}
	return nil
}

func (local *LsLocal) handleConn(userConn *net.TCPConn) {
	defer userConn.Close()

	//socks5协议处理
	buf := make([]byte, 256)
	/**
	   The localConn connects to the dstServer, and sends a ver
	   identifier/method selection message:
		          +----+----------+----------+
		          |VER | NMETHODS | METHODS  |
		          +----+----------+----------+
		          | 1  |    1     | 1 to 255 |
		          +----+----------+----------+
	   The VER field is set to X'05' for this ver of the protocol.  The
	   NMETHODS field contains the number of method identifier octets that
	   appear in the METHODS field.
	*/
	// 第一个字段VER代表Socks的版本，Socks5默认为0x05，其固定长度为1个字节

	versionBuf := make([]byte, 3)
	_, err := io.ReadFull(userConn, versionBuf)
	// 只支持版本5
	if err != nil || versionBuf[0] != 0x05 {
		log.Println("VER only support 0x5")
		return
	}

	/**
	   The dstServer selects from one of the methods given in METHODS, and
	   sends a METHOD selection message:

		          +----+--------+
		          |VER | METHOD |
		          +----+--------+
		          | 1  |   1    |
		          +----+--------+
	*/
	// 不需要验证，直接验证通过
	userConn.Write([]byte{0x05, 0x00})

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/

	// 获取真正的远程服务的地址
	atypeBuf := make([]byte, 4)
	_, err = io.ReadFull(userConn, atypeBuf)
	if err != nil {
		log.Println("read address head error: ", err)
		return
	}

	// CMD代表客户端请求的类型，值长度也是1个字节，有三种类型
	// CONNECT X'01'
	if atypeBuf[1] != 0x01 {
		// 目前只支持 CONNECT
		log.Println("only support CONNECT", buf[1])
		return
	}

	sendBuf := &bytes.Buffer{}
	sendBuf.Write(atypeBuf[3:])

	switch atypeBuf[3] {
	case 1: //ipv4
		ipv4Buf := make([]byte, net.IPv4len+2)
		_, err = io.ReadFull(userConn, ipv4Buf)
		if err != nil {
			log.Println("read ipv4 error: ", err)
			return
		}
		sendBuf.Write(ipv4Buf)
	case 3: //domain len+domain
		domainLenBuf := make([]byte, 1)
		_, err = io.ReadFull(userConn, domainLenBuf)
		if err != nil {
			log.Println("read domain len error: ", err)
			return
		}
		sendBuf.Write(domainLenBuf)

		if domainLenBuf[0] <= 0 || domainLenBuf[0] > 255 {
			log.Println("domain len error: ", domainLenBuf[0])
			return
		}

		domainBuf := make([]byte, int(domainLenBuf[0])+2)
		_, err = io.ReadFull(userConn, domainBuf)
		if err != nil {
			log.Println("read domain error: ", err)
			return
		}

		log.Printf("start connect %v\n", string(domainBuf[:domainLenBuf[0]]))
		sendBuf.Write(domainBuf)

	case 4: //ipv6
		ipv6Buf := make([]byte, net.IPv6len+2)
		_, err = io.ReadFull(userConn, ipv6Buf)
		if err != nil {
			log.Println("read ipv6 error: ", err)
			return
		}
		sendBuf.Write(ipv6Buf)

	default: //不支持
		log.Println("unkown atype ", atypeBuf[3])
		return
	}

	proxyServer, err := local.DialRemote()
	if err != nil {
		log.Println(err)
		return
	}

	defer proxyServer.Close()
	// Conn被关闭时直接清除所有数据 不管没有发送的数据
	proxyServer.SetLinger(0)

	local.EncodeWrite(proxyServer, sendBuf.Bytes())

	close_chan := make(chan int)

	// 进行转发
	// 从 proxyServer 读取数据发送到 localUser
	go func() {
		_ = local.DecodeCopy(userConn, proxyServer)
		close_chan <- 1
	}()

	// 从 localUser 发送数据发送到 proxyServer，这里因为处在翻墙阶段出现网络错误的概率更大
	go func() {
		local.EncodeCopy(proxyServer, userConn)
		close_chan <- 1
	}()

	<-close_chan
}
