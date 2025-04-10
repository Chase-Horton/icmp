package main

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"time"

	"github.com/chase-horton/go-icmp/protocol_unix"
)

func recievePackets(packets chan string) {
	fdRecv, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		panic("Error opening recv socket")
	}
	defer syscall.Close(fdRecv)
	buf := make([]byte, 65535)
	for {
		n, _, err := syscall.Recvfrom(fdRecv, buf, 0)
		if err != nil {
			panic(err)
		}
		dump := hex.Dump(buf[:n])
		packets <- fmt.Sprintf("Recieved Packed:\n%s", dump)
	}
}
func main() {
	start := time.Now()
	src, dest := [4]byte{}, [4]byte{}
	numberOfIps := int(4.2 * 10e9)
	fmt.Printf("number of ips %d", numberOfIps)
	for range numberOfIps {
		protocol_unix.MakePingPacket(src, dest)
	}
	dur := time.Since(start)
	fmt.Printf("Time Taken: %s\n", dur)
	fmt.Printf("Time Taken per: %s\n", dur/1000000)
}

// func main() {
// 	src, dest := [4]byte{192, 168, 0, 169}, [4]byte{142, 250, 113, 139}
// 	packets := make(chan string)
// 	go recievePackets(packets)
// 	packet := protocol_unix.MakePingPacket(src, dest)
//
// 	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
// 	if err != nil {
// 		panic("Error opening send socket")
// 	}
// 	defer syscall.Close(fd)
//
// 	addr := &syscall.SockaddrInet4{Port: 0}
// 	copy(addr.Addr[:], []byte{142, 250, 113, 139})
// 	syscall.Sendto(fd, packet, 0, addr)
// 	for packetRecv := range packets {
// 		fmt.Println(packetRecv)
// 	}
// }
