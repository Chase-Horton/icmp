package protocol_unix

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"testing"
)

func TestPing(t *testing.T) {
	src, dest := [4]byte{192, 168, 0, 169}, [4]byte{142, 250, 113, 139}

	packet := MakePingPacketFast(src, dest)
	dump := hex.Dump(packet[:])
	fmt.Println("Packet length:", len(packet), "bytes")
	fmt.Println(dump)
}

func TestRealPing(t *testing.T) {
	src, dest := [4]byte{192, 168, 0, 169}, [4]byte{142, 250, 113, 139}

	packet := MakePingPacketFast(src, dest)

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		t.Errorf("error opening socket")
	}
	addr := &syscall.SockaddrInet4{Port: 0}
	copy(addr.Addr[:], []byte{142, 250, 113, 139})
	syscall.Sendto(fd, packet[:], 0, addr)
}
