package main

import (
	"fmt"
	"net"
	"time"
)

const (
	address = "142.250.114.113" // Replace with the actual IPv4 address
)

type icmpMessage struct {
	Type         uint8  // Type of the ICMP message
	Code         uint8  // Code of the ICMP message
	Checksum     uint16 // Checksum for error-checking
	RestOfHeader uint32 // Rest of the header (depends on the type and code)
	Data         []byte // Payload data
}

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

func main() {
	conn, err := net.Dial("ip4:icmp", address)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Successfully connected to", address)

	msg := icmpMessage{
		Type:         8,
		Code:         0,
		Checksum:     0, // Checksum will be calculated later
		RestOfHeader: 0,
		Data:         []byte("Hello, ICMP!"),
	}

	raw := []byte{
		msg.Type,
		msg.Code,
		0, 0, // Placeholder for checksum
		byte(msg.RestOfHeader >> 24),
		byte(msg.RestOfHeader >> 16),
		byte(msg.RestOfHeader >> 8),
		byte(msg.RestOfHeader & 0xff),
	}
	raw = append(raw, msg.Data...)

	msg.Checksum = checkSum(raw)
	raw[2] = byte(msg.Checksum >> 8)
	raw[3] = byte(msg.Checksum & 0xff)

	start := time.Now()
	_, err = conn.Write(raw)
	if err != nil {
		fmt.Println("Failed to send packet:", err)
		return
	}
	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	internetHeaderLength := (reply[1] & 0x0f) * 4 // Get the Internet Header Length (IHL) from the IP header
	if internetHeaderLength == 0 {
		internetHeaderLength = 20
	}
	icmpData := reply[internetHeaderLength:] // Skip the IP header

	replyData := icmpData[8:] //Skip the ICMP header (first 8 bytes)
	if err != nil {
		fmt.Println("Failed to read response:", err)
		return
	}
	elapsed := time.Since(start)

	fmt.Println("Received reply in", elapsed)
	fmt.Println("Reply data:", string(replyData))
}
