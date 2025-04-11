package main

import (
	"encoding/hex"
	"fmt"
	"syscall"
	"testing"
)

func TestMakePacket(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	src := [4]byte{192, 168, 0, 169}
	host := [4]byte{142, 250, 113, 139}

	_, addr := establish_proxy(host)
	baseHeader = initBaseHeader(src, addr)

	packets := makePackets(data)
	if len(packets) != 1 {
		t.Error("packets != 1")
	}
	if len(packets[0]) != 28+9 {
		t.Error("packet != 37")
	}
	packet := packets[0]
	cs := uint16(packet[10])<<8 | uint16(packet[11])
	packet[10] = 0
	packet[11] = 0
	if checkSum(packet[:20]) != cs {
		t.Error("checksums not eq")
	}
	packets = makePackets([]byte{1, 2, 3, 4})
	packet = packets[0]
	fmt.Println(hex.Dump(packet))
}
func TestSendPacket(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	src := [4]byte{192, 168, 0, 169}
	dest := [4]byte{142, 250, 113, 139}

	fd, addr := establish_proxy(dest)
	baseHeader = initBaseHeader(src, addr)

	packets := makePackets(data)
	if len(packets) != 1 {
		t.Error("packets != 1")
	}
	if len(packets[0]) != 28+9 {
		t.Error("packet != 37")
	}
	packet := packets[0]
	cs := uint16(packet[10])<<8 | uint16(packet[11])
	packet[10] = 0
	packet[11] = 0
	if checkSum(packet[:20]) != cs {
		t.Error("checksums not eq")
	}
	packets = makePackets(data)
	packet = packets[0]
	fmt.Println("sending:", hex.Dump(packet))

	syscall.Sendto(fd, packet[:], 0, addr)
}
