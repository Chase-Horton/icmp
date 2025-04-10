package protocol_unix

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"testing"
)

func TestProtocol(t *testing.T) {
	src, dest := [4]byte{192, 168, 0, 1}, [4]byte{170, 250, 15, 2}
	header := makeIpv4Header(src, dest)

	if header.VersionAndIHL != 0x45 {
		t.Errorf("invalid version or ihl")
	}
	if header.TOS != 0x0 {
		t.Errorf("invalid TOS")
	}
	if header.TotalLength != 29 {
		t.Errorf("invalid Length")
	}
	if header.FlagsAndFrag != 0x2<<13 {
		t.Errorf("invalid flags")
	}
	if header.TTL != 64 {
		t.Errorf("invalid TTL")
	}
	if header.Protocol != 1 {
		t.Errorf("invalid protocol")
	}
	bytes := header.Bytes()
	if len(bytes) != 20 {
		t.Errorf("invalid header length %d", len(bytes))
	}
	// err := os.WriteFile("ping_header.bin", bytes, 0644)
	// if err != nil {
	// 	t.Fatalf("failed to write binary file: %v", err)
	// }
	dump := hex.Dump(bytes)
	fmt.Println(dump)
}
func TestPing(t *testing.T) {
	src, dest := [4]byte{192, 168, 0, 169}, [4]byte{142, 250, 113, 139}

	packet := MakePingPacket(src, dest)
	dump := hex.Dump(packet)
	fmt.Println("Packet length:", len(packet), "bytes")
	fmt.Println(dump)
}

func TestRealPing(t *testing.T) {
	src, dest := [4]byte{192, 168, 0, 169}, [4]byte{142, 250, 113, 139}

	packet := MakePingPacket(src, dest)

	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_RAW, syscall.IPPROTO_RAW)
	if err != nil {
		t.Errorf("error opening socket")
	}
	addr := &syscall.SockaddrInet4{Port: 0}
	copy(addr.Addr[:], []byte{142, 250, 113, 139})
	syscall.Sendto(fd, packet, 0, addr)
}
