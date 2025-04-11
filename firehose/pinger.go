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

const workers = 64

func pingEveryAddressParallel() {
	//test
	start := time.Now()
	src := [4]byte{192, 168, 0, 169}
	ipChan := make(chan [4]byte, 1000)

	// Launch workers
	for range workers {
		go func() {
			fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
			if err != nil {
				panic("Error opening send socket")
			}
			defer syscall.Close(fd)

			for dest := range ipChan {
				packet := protocol_unix.MakePingPacketFast(src, dest)
				addr := &syscall.SockaddrInet4{}
				copy(addr.Addr[:], dest[:])
				syscall.Sendto(fd, packet[:], 0, addr)
			}
		}()
	}

	// Feed IPs to workers
	var sent int
	for a := byte(1); a <= 255; a++ {
		for b := byte(0); b <= 255; b++ {
			for c := byte(0); c <= 255; c++ {
				for d := byte(1); d <= 254; d++ {
					// skip reserved IPs
					if a == 10 || (a == 172 && b >= 16 && b <= 31) || (a == 192 && b == 168) || a == 127 ||
						(a == 169 && b == 254) || (a == 100 && b >= 64 && b <= 127) || a == 0 ||
						(a >= 224 && a <= 239) || a >= 240 || (a == 255 && b == 255 && c == 255 && d == 255) {
						continue
					}
					ipChan <- [4]byte{a, b, c, d}
					sent++
					if sent == 10_000_000 {
						//test
						fmt.Printf("10m packets sent in %s", time.Since(start))
						close(ipChan)
						return
					}
					if sent%1_000_000 == 0 {
						fmt.Printf("%d packets sent\n", sent)
					}
				}
			}
		}
	}
	close(ipChan)
}

func main() {
	packets := make(chan string)
	go recievePackets(packets)

	go pingEveryAddressParallel()

	for range packets {
		//fmt.Println(packetRecv)
	}
}

func test() {
	src, dest := [4]byte{192, 168, 0, 169}, [4]byte{142, 250, 113, 139}
	packets := make(chan string)
	go recievePackets(packets)
	packet := protocol_unix.MakePingPacketFast(src, dest)

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		panic("Error opening send socket")
	}
	defer syscall.Close(fd)

	addr := &syscall.SockaddrInet4{Port: 0}
	//test
	dump := hex.Dump(packet[:])
	fmt.Println(dump)
	copy(addr.Addr[:], []byte{142, 250, 113, 139})
	syscall.Sendto(fd, packet[:], 0, addr)

	for packetRecv := range packets {
		fmt.Println(packetRecv)
	}
}
