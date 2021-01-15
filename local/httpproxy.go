package local

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/guange2015/lightsocks/cmd"
	"github.com/guange2015/lightsocks/core"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var localSocksAddr string

func StartProxy(config *cmd.Config) {
	addr, _ := net.ResolveTCPAddr("tcp", config.Httpproxy)
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	localSocksAddr = config.ListenAddr

	log.Println("start http proxy: ", addr)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatal("accept error: ", err)
		}
		go handleConn(conn)
	}


}

func handleConn(conn *net.TCPConn) {
	buf := make([]byte, 255)

	var tcpConn *net.TCPConn

	close_chan := make(chan int)

	for {
		i, err := core.TcpRead(conn, buf)
		if err != nil {
			log.Println(err)
			return
		}

		if strings.HasPrefix(string(buf[:i]), "CONNECT") {
			//通过sock5连接服务器
			//CONNECT www.baidu.com:443 HTTP/1.1
			ss := strings.Split(string(buf[:i]), " ")
			tcpConn, err = connectRemote(ss[1])
			if err != nil {
				log.Printf("connect remove error: %v, %v\n",
					ss[1], err)
				return
			}

			//连接成功

			core.TcpWrite(conn,[]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

			defer func() {
				log.Println("tcpConn close")
				tcpConn.Close()
			}()


			go func() {
				copySocket(conn, tcpConn)
				close_chan <- 1
			}()

			go func() {
				copySocket(tcpConn, conn)
				close_chan <- 1
			}()

			break
		}
	}

	<- close_chan

	log.Printf("close socket %v\n", tcpConn.RemoteAddr())
}

func copySocket(src *net.TCPConn, dst *net.TCPConn) {
	buf := make([]byte, 255)
	for {
		i, err :=  core.TcpRead(src, buf)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println(string(buf[:i]))

		if dst != nil {
			_, err := core.TcpWrite(dst,buf[:i])
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func connectRemote(address string) (*net.TCPConn, error)  {
	addr, _ := net.ResolveTCPAddr("tcp", address)

	matched, err := regexp.MatchString(`\d+\.\d+\.\d+\.\d+:\+`, address)
	atype := 1
	if matched {
		atype = 1 //ip
	} else {
		atype = 3 //domain
	}

	log.Println("start connect: ", address)

	localSocks, _ := net.ResolveTCPAddr("tcp", localSocksAddr)
	tcpConn, err := net.DialTCP("tcp", nil, localSocks)
	if err!=nil{
		log.Fatal("connect local socks error: ",  err)
		return nil, err
	}

	buf := make([]byte, 255)

	core.TcpWrite(tcpConn, []byte{0x5,0,0})

	n, err := core.TcpRead(tcpConn,buf)
	if err != nil {
		return nil, err
	}

	if n==2 && buf[0]==0x5 && buf[1]==0 {
		//验证通过
	} else {
		return nil, errors.New("local socks verify error")
	}



	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, uint8(0x5))
	binary.Write(buffer, binary.BigEndian, uint8(0x1))
	binary.Write(buffer, binary.BigEndian, uint8(0x0))
	//atyp 1 ip, 3 domain
	binary.Write(buffer, binary.BigEndian, uint8(atype))
	if atype==1 {
		//地址
		binary.Write(buffer, binary.BigEndian, addr.IP.To4())
		//PORT
		binary.Write(buffer, binary.BigEndian, uint16(addr.Port))
	} else {
		ss := strings.Split(address, ":")
		binary.Write(buffer, binary.BigEndian, uint8(len(ss[0])))
		log.Println("write domain len:", len(ss[0]))
		buffer.Write([]byte(ss[0]))
		port, _ := strconv.Atoi(ss[1])
		binary.Write(buffer, binary.BigEndian, uint16(port))
		log.Println("write domain port:", uint16(port))
	}


	bufs := buffer.Bytes()
	_, err = core.TcpWrite(tcpConn, bufs)
	if err != nil {
		log.Println("write head:", err)
		return nil, err
	}

	n, err = core.TcpRead(tcpConn,buf)
	if err != nil {
		return nil, err
	}

	if n==10 && buf[0]==0x5 && buf[1]==0 {
		//验证通过
	} else {
		return nil, errors.New("socks connect remote error")
	}

	return tcpConn, nil
}