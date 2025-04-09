package main

import (
	"fmt"
	"net"
)

const (
	ipv4Address = "127.0.0.1" // Replace with the actual IPv4 address
	port        = "8080"      // Replace with the desired port
)

func main() {
	address := fmt.Sprintf("%s:%s", ipv4Address, port)

	conn, err := net.Dial("tcp4", address)
	if err != nil {
		fmt.Println("Error dialing:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Successfully connected to", address)

	// You can now use the 'conn' to send and receive data
}
