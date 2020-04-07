package core

import (
	"net"
	"testing"
)

func TestFormatCurrentTime(t *testing.T)  {
	time := formatCurrentTime()
	t.Log(time)
	time = formatCurrentTime()
	t.Log(time)
}

func TestFormatIp(t *testing.T)  {
	addr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:80")
	t.Log(formatIp(addr))
}

func TestDebugNet(t *testing.T)  {
	localAddr , _ := net.ResolveTCPAddr("tcp", "192.168.146.6:62625")
	remoteAddr, _ := net.ResolveTCPAddr("tcp", "192.168.146.131:22")

	buf := []byte{
		0x45,00,00, 0x6c, 0x4c, 0xc9, 0x40,0x00, 0x80,0x06, 0x07, 0xe8, 0xc0, 0xa8, 0x92, 0x06,
		0x6a,0xec, 0x95,0xb2, 0x65,0x6f, 0x2d,0x67, 0x03,0x44, 0xf6,0xde,
	}

	debugNet(localAddr, remoteAddr, buf, Read)
}