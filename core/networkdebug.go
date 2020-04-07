package core

import (
	"fmt"
	"net"
	"time"
)


type NetworkDirection int
const (
	Read NetworkDirection = 1 + iota
	Write
)

//网络字节流输出日志格式
//14:31:53.579814 192.168.146.6.62625 > 192.168.146.131.22 length 68
//0x0000:  4500 006c 4cc9 4000 8006 07e8 c0a8 9206  E..lL.@.........
//0x0010:  c0a8 9283 f4a1 0016 249f 73ff e59b 54e0  ........$.s...T.
//0x0020:  5018 1005 b7e2 0000 0000 0020 8b78 77e3  P............xw.
//0x0030:  d830 0129 84fe 2b2e b27e cebb 8b81 ca8d  .0.)..+..~......
//0x0040:  1dde fb10 3e63 0694 e48e 773d 9fcd 072d  ....>c....w=...-
//0x0050:  6fd9 30a1 73be c6bb a8f3 64a2 f406 4acf  o.0.s.....d...J.
//0x0060:  6aec 95b2 656f 2d67 0344 f6de            j...eo-g.D..

//direction 方向
func DebugNet(conn *net.TCPConn, buf []byte, direction NetworkDirection)  {
	debugNet(conn.LocalAddr(), conn.RemoteAddr(), buf, direction)
}

func debugNet(localAddr net.Addr, remoteAddr net.Addr, buf []byte, direction NetworkDirection)  {
	//打印头部
	fmt.Println()
	direction_s := "<"
	if direction == Write{
		direction_s = ">"
	}

	fmt.Printf("%v %v %v %v length %v\n",
		formatCurrentTime(),
		formatIp(localAddr),
		direction_s,
		formatIp(remoteAddr),
		len(buf))

	//打印内容体
	//16字节一行
	num := 1
	for i := 0; i < len(buf); i+=16 {
		fmt.Printf("0x%04d:  ", num)

		for j := 0; j < 16; j++ {
			if j+i < len(buf) {
				fmt.Printf("%02x", buf[i+j])
			} else {//内容不够打空格
				fmt.Printf("  ")
			}

			if j%2==1 { //两字节一空格
				fmt.Printf(" ")
			}
		}

		fmt.Printf(" ")

		//打ASCII码
		for j := 0; j < 16; j++ {
			if j+i < len(buf) {
				fmt.Printf(formatAscii(buf[i+j]))
			}
		}

		fmt.Println("")

		num++
	}
	fmt.Println()
}

func formatCurrentTime() string {
	now := time.Now()
	return fmt.Sprintf("%d:%d:%d.%v", now.Hour(), now.Second(), now.Minute(), now.Nanosecond()/1000)
}

func formatIp(addr net.Addr) string {
	return addr.String();
}

func formatAscii(b byte) string {
	if b >= 33 && b <= 126 {
		return fmt.Sprintf("%c", b)
	}
	return "."
}