package local

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"strings"
)

func StartProxy()  {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:1097")
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

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
		i, err := conn.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println(string(buf[:i]))


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

			conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

			defer tcpConn.Close()

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
		i, err := src.Read(buf)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println(string(buf[:i]))

		if dst != nil {
			_, err := dst.Write(buf[:i])
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
}

func connectRemote(address string) (*net.TCPConn, error)  {
	addr, _ := net.ResolveTCPAddr("tcp", address)

	localSocks, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:1099")
	tcpConn, err := net.DialTCP("tcp", nil, localSocks)
	if err!=nil{
		return nil, err
	}

	buf := make([]byte, 255)

	tcpConn.Write([]byte{0x5,0,0})

	_, err = tcpConn.Read(buf)
	if err != nil {
		return nil, err
	}

	/**
	  +----+-----+-------+------+----------+----------+
	  |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
	  +----+-----+-------+------+----------+----------+
	  | 1  |  1  | X'00' |  1   | Variable |    2     |
	  +----+-----+-------+------+----------+----------+
	*/
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, 0x5)
	binary.Write(buffer, binary.BigEndian, 0x1)
	binary.Write(buffer, binary.BigEndian, 0x0)
	//atyp 1 ip, 3 domain
	binary.Write(buffer, binary.BigEndian, 0x1)
	//地址
	binary.Write(buffer, binary.BigEndian, addr.IP)
	//PORT
	binary.Write(buffer, binary.BigEndian, uint16(addr.Port))

	_, err = tcpConn.Read(buf)
	if err != nil {
		return nil, err
	}

	log.Println(buf[0])

	return tcpConn, nil
}