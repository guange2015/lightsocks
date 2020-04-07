package core

import "net"

const NetDebug = true

func TcpRead(conn *net.TCPConn,b []byte) (int, error) {
	read, err := conn.Read(b)
	if NetDebug {
		if err ==nil && read > 0{
			DebugNet(conn, b[:read], Read)
		}
	}
	return read, err
}


func TcpWrite(conn *net.TCPConn, b []byte) (int, error){
	if NetDebug {
		DebugNet(conn, b, Write)
	}

	return conn.Write(b)
}