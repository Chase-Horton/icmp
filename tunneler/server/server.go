package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

//	func openTunForTest(name string) *os.File {
//		file, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
//		if err != nil {
//			panic("failed to open tun")
//		}
//	}
type Interface struct {
	io.ReadWriteCloser
	name string
}

func openDev() (ifce *Interface, err error) {
	var fdInt int
	if fdInt, err = syscall.Open(
		"/dev/net/tun", os.O_RDWR|syscall.O_NONBLOCK, 0); err != nil {
		return nil, err
	}

	name, err := setupFd(uintptr(fdInt))
	if err != nil {
		return nil, err
	}

	return &Interface{
		ReadWriteCloser: os.NewFile(uintptr(fdInt), "tun"),
		name:            name,
	}, nil
}
func setupFd(fd uintptr) (name string, err error) {
	var flags uint16 = 0x1000 //cIFFNOPI
	flags |= 0x0002           // cIFFTUN

	if name, err = createInterface(fd, "tun_server", flags); err != nil {
		return "", err
	}
	return
}
func createInterface(fd uintptr, ifName string, flags uint16) (createdIFName string, err error) {
	var req ifReq
	req.Flags = flags
	copy(req.Name[:], ifName)

	err = ioctl(fd, syscall.TUNSETIFF, uintptr(unsafe.Pointer(&req)))
	if err != nil {
		return
	}

	createdIFName = strings.Trim(string(req.Name[:]), "\x00")
	return
}

type ifReq struct {
	Name  [0x10]byte
	Flags uint16
	pad   [0x28 - 0x10 - 2]byte
}

func ioctl(fd uintptr, request uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(request), argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

const TEST = true

func checkSum(data []byte) uint16 {
	sum := 0
	for i := 0; i < len(data)-1; i += 2 {
		sum += int(data[i])<<8 | int(data[i+1])
	}
	if len(data)%2 == 1 {
		sum += int(data[len(data)-1]) << 8
	}
	for (sum >> 16) > 0 {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	return uint16(^sum)
}

func establish_proxy(server [4]byte) (int, *syscall.SockaddrInet4) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		panic("error opening send socket")
	}
	addr := &syscall.SockaddrInet4{Port: 0}
	copy(addr.Addr[:], server[:])
	return fd, addr
}
func recievePackets(server *syscall.SockaddrInet4, packets chan []byte) {
	fdRecv, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_ICMP)
	if err != nil {
		panic("Error opening recv socket")
	}
	defer syscall.Close(fdRecv)
	buf := make([]byte, 65535)
	for {
		n, from, err := syscall.Recvfrom(fdRecv, buf, 0)
		fmt.Printf("bytes recieved:\n%d", n)
		if err != nil {
			panic(err)
		}
		switch sa := from.(type) {
		case *syscall.SockaddrInet4:
			//if testing locally all addresses will resolve to 127.0.0.1 so need to parse packet
			if TEST {
				packets <- buf[:n]
			} else {
				if sa.Addr == server.Addr {
					packets <- buf[:n]
				}
			}
			fmt.Println("addresses")
			fmt.Println(sa.Addr)
			fmt.Println(server.Addr)
		}
	}
}
func initBaseHeader(src [4]byte, dest *syscall.SockaddrInet4) [28]byte {
	return [28]byte{
		4<<4 | 5,     // Vers
		0,            // TOS
		0,            //totlength 1/2
		29,           //totlength 2/2
		0xCA,         // ID 1/2
		0xFE,         // ID 2/2
		0b01000000,   //flags 1/2
		0,            //flags 2/2
		64,           //ttl
		1,            //protocol icmp
		0,            //ipv4 checksum 1/2
		0,            //ipv4 checksum 2/2
		src[0],       //src 1/4
		src[1],       //src 2/4
		src[2],       //src 3/4
		src[3],       //src 4/4
		dest.Addr[0], //dest 1/4
		dest.Addr[1], //dest 2/4
		dest.Addr[2], //dest 3/4
		dest.Addr[3], //dest 4/4
		8,            //icmp type ping
		0,            //code of icmp message
		0,            //checksum 1/2
		0,            //checksum 2/2
		0,            //rest of header 1/4
		0,            //rest of header 2/4
		0,            //rest of header 3/4
		0,            //rest of header 4/4
	}
}

func makePacket(data []byte) []byte {
	numDataByte := len(data)
	baseCpy := baseHeader
	baseCpy[2] = byte(numDataByte >> 8)
	baseCpy[3] = byte(numDataByte)

	v4Checksum := checkSum(baseCpy[:20])
	baseCpy[10] = byte(v4Checksum >> 8)
	baseCpy[11] = byte(v4Checksum)

	packet := append(baseCpy[:], data...)

	icmpChecksum := checkSum(packet[20:])
	packet[22] = byte(icmpChecksum >> 8)
	packet[23] = byte(icmpChecksum)
	return packet
}

const maxPayloadSize = 65500

func makePackets(data []byte) [][]byte {
	packets := [][]byte{}
	//maximum length of 1 ipv4 packet is 65,535 bytes
	//header is 28 so max 65,507 of data
	for offset := 0; offset < len(data); offset += maxPayloadSize {
		end := min(offset+maxPayloadSize, len(data))

		chunk := data[offset:end]
		packet := makePacket(chunk)
		packets = append(packets, packet)
	}
	return packets
}

var baseHeader [28]byte

func sendPackets(fd int, src [4]byte, dest *syscall.SockaddrInet4, packets chan []byte) {
	baseHeader = initBaseHeader(src, dest)
	for packet := range packets {
		fmt.Printf("Sending:\n%s", hex.Dump(packet))
		syscall.Sendto(fd, packet[:], 0, dest)
	}
}
func parsePacket(packet []byte) []byte {
	var totalLen uint16 = uint16(packet[2])<<8 | uint16(packet[3])
	//check checksums and req data if it was wrong
	return packet[28:totalLen]
}
func main() {
	if !TEST {
		src := [4]byte{127, 0, 0, 3}
		client := [4]byte{127, 0, 0, 2}

		fd, clientAddr := establish_proxy(client)

		packetsRecv := make(chan []byte)
		packetsSend := make(chan []byte)

		go recievePackets(clientAddr, packetsRecv)
		go sendPackets(fd, src, clientAddr, packetsSend)
		// add something here to intercept traffic on client
		// when traffic is recieved

		fmt.Printf("Listening on port %d.%d.%d.%d\n", client[0], client[1], client[2], client[3])
		for packetRecv := range packetsRecv {
			data := parsePacket(packetRecv)
			fmt.Println("Parsed:")
			fmt.Println(hex.Dump(data))
			//packetsSend <- makePackets([]byte("Message Recieved :)"))[0]
		}
	} else {
		tun, err := openDev()
		if err != nil {
			panic(err)
		}
		fmt.Println("Listening on tun_server")
		frame := []byte{}
		for {
			n, err := tun.Read([]byte(frame))
			if err != nil {
				panic(err)
			}
			dat := frame[:n]
			if n != 0 {
				fmt.Println(hex.Dump(dat))
			}
		}
	}
}
