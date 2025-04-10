package protocol_unix

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type icmpMsg struct {
	Type         uint8  // Type of the ICMP message
	Code         uint8  // Code of the ICMP message
	Checksum     uint16 // Checksum for error-checking
	RestOfHeader uint32 // Rest of the header (depends on the type and code)
	Data         []byte // Payload data
}

func makeIcmpMsg() *icmpMsg {
	msg := &icmpMsg{
		Type:         8,
		Code:         0,
		Checksum:     0, // Checksum will be calculated later
		RestOfHeader: 0,
		Data:         []byte("!"),
	}
	return msg
}
func (msg *icmpMsg) Bytes() []byte {
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
	return raw
}

type ipv4Header struct {
	VersionAndIHL      uint8
	TOS                uint8
	TotalLength        uint16
	Identification     uint16
	FlagsAndFrag       uint16
	TTL                uint8
	Protocol           uint8
	HeaderChecksum     uint16
	SourceAddress      uint32
	DestinationAddress uint32
}

func (ip *ipv4Header) Bytes() []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, ip.VersionAndIHL)
	_ = binary.Write(buf, binary.BigEndian, ip.TOS)
	_ = binary.Write(buf, binary.BigEndian, ip.TotalLength)
	_ = binary.Write(buf, binary.BigEndian, ip.Identification)
	_ = binary.Write(buf, binary.BigEndian, ip.FlagsAndFrag)
	_ = binary.Write(buf, binary.BigEndian, ip.TTL)
	_ = binary.Write(buf, binary.BigEndian, ip.Protocol)
	_ = binary.Write(buf, binary.BigEndian, ip.HeaderChecksum)
	_ = binary.Write(buf, binary.BigEndian, ip.SourceAddress)
	_ = binary.Write(buf, binary.BigEndian, ip.DestinationAddress)
	return buf.Bytes()
}

func makeIpv4Header(source, dest [4]byte) *ipv4Header {
	header := &ipv4Header{}
	header.VersionAndIHL = 4<<4 | 5
	header.TOS = 0          //dscp | ecn 0<<6 | 0
	header.TotalLength = 29 //20 header 8 icmp header 1 data
	header.Identification = 0xCAFE
	header.FlagsAndFrag = 0b010 << 13 //0<<15 | 1<<14 | 0 <<13 | 0 // flags reserved, dont fragment, more fragments, frag offset
	header.TTL = 64                   //max 64 hops
	header.Protocol = 1               //icmp
	header.HeaderChecksum = 0
	header.SourceAddress = binary.BigEndian.Uint32(source[:])
	header.DestinationAddress = binary.BigEndian.Uint32(dest[:])

	bytes := header.Bytes()
	header.HeaderChecksum = checkSum(bytes)
	return header
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

func MakePingPacket(src, dest [4]byte) []byte {
	v4Header := makeIpv4Header(src, dest)
	icmpMsg := makeIcmpMsg()
	totalBytes := []byte{}
	totalBytes = append(totalBytes, v4Header.Bytes()...)
	totalBytes = append(totalBytes, icmpMsg.Bytes()...)
	return totalBytes
}

type PingResult struct {
	Success  bool // Indicates if the ping was successful
	Duration time.Duration
	Error    error
}

func (pr *PingResult) String() string {
	return fmt.Sprintf("Success: %v, Duration: %v, Error: %v", pr.Success, pr.Duration, pr.Error)
}
func (pr *PingResult) Bytes() [2]byte {
	seconds := byte(int(pr.Duration.Abs().Seconds()))
	var success byte = 0
	if pr.Success {
		success = 1
	}
	return [2]byte{success, seconds}
}

type BitWriter struct {
	file     *os.File
	currByte byte
	bitCount uint8
}

func NewBitWriter(filename string) *BitWriter {
	f, err := os.Create(filename)
	if err != nil {
		panic(fmt.Sprintf("error creating file: %v", err))
	}
	return &BitWriter{
		file: f,
	}
}
func (bw *BitWriter) writeBit(bit bool) {
	if bit {
		bw.currByte = bw.currByte<<1 | 1
	} else {
		bw.currByte = bw.currByte<<1 | 0
	}

	if bw.bitCount == 8 {
		_, err := bw.file.Write([]byte{bw.currByte})
		if err != nil {
			panic(fmt.Sprintf("error writing bit %v", err))
		}
	}
}
func (bw *BitWriter) writeFinalByte() {
	if bw.bitCount > 0 {
		for bw.bitCount < 8 {
			bw.currByte = bw.currByte<<1 | 0
			bw.bitCount++
		}
		_, err := bw.file.Write([]byte{bw.currByte})
		if err != nil {
			panic(fmt.Sprintf("error writing bit %v", err))
		}
		bw.file.Close()
	}
}
