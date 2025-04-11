package protocol

import (
	"fmt"
	"net"
	"time"
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

type PingResult struct {
	Success  bool // Indicates if the ping was successful
	Duration time.Duration
	Error    error
}

func (pr *PingResult) String() string {
	return fmt.Sprintf("Success: %v, Duration: %v, Error: %v", pr.Success, pr.Duration, pr.Error)
}
func (pr *PingResult) FileString() string {
	return fmt.Sprintf("%v,%v,%v", pr.Success, pr.Duration, pr.Error)
}

func NewPingResult(success bool, duration time.Duration, err error) PingResult {
	return PingResult{
		Success:  success,
		Duration: duration,
		Error:    err,
	}
}
func Ping(address string) PingResult {
	conn, err := net.Dial("ip4:icmp", address)
	if err != nil {
		return NewPingResult(false, 0, fmt.Errorf("error dialing: %w", err))
	}
	defer conn.Close()

	msg := icmpMessage{
		Type:         8,
		Code:         0,
		Checksum:     0, // Checksum will be calculated later
		RestOfHeader: 0,
		Data:         []byte("!"),
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
		return NewPingResult(false, 0, fmt.Errorf("error writing to connection: %w", err))
	}
	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	if err != nil {
		return NewPingResult(false, 0, fmt.Errorf("error reading from connection: %w", err))
	}
	timeElapsed := time.Since(start)
	return NewPingResult(true, timeElapsed, nil)
}
func TryPing(ip [4]byte, maxDuration time.Duration) PingResult {
	// Attempt to ping the IP address
	ipStr := fmt.Sprintf("%d.%d.%d.%d", ip[0], ip[1], ip[2], ip[3])
	c1 := make(chan PingResult)
	go func() {
		result := Ping(ipStr)
		c1 <- result
	}()
	select {
	case result := <-c1:
		//fmt.Println("Ping result:", result)
		return result
	case <-time.After(maxDuration):
		//fmt.Println("Ping timed out")
		return NewPingResult(false, 0, fmt.Errorf("ping timed out"))
	}
}
