package protocol_unix

import (
	"fmt"
	"os"
	"time"
)

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

var baseIpPacket = [29]byte{
	4<<4 | 5,   // Vers
	0,          // TOS
	0,          //totlength 1/2
	29,         //totlength 2/2
	0xCA,       // ID 1/2
	0xFE,       // ID 2/2
	0b01000000, //flags 1/2
	0,          //flags 2/2
	64,         //ttl
	1,          //protocol icmp
	0,          //ipv4 checksum 1/2
	0,          //ipv4 checksum 2/2
	0,          //src 1/4
	0,          //src 2/4
	0,          //src 3/4
	0,          //src 4/4
	0,          //dest 1/4
	0,          //dest 2/4
	0,          //dest 3/4
	0,          //dest 4/4
	8,          //icmp type ping
	0,          //code of icmp message
	0xd6,       //checksum 1/2
	0xff,       //checksum 2/2
	0,          //rest of header 1/4
	0,          //rest of header 2/4
	0,          //rest of header 3/4
	0,          //rest of header 4/4
	'!',        //data: !
}

func MakePingPacketFast(src, dest [4]byte) [29]byte {
	packet := baseIpPacket
	packet[12] = src[0]
	packet[13] = src[1]
	packet[14] = src[2]
	packet[15] = src[3]
	packet[16] = dest[0]
	packet[17] = dest[1]
	packet[18] = dest[2]
	packet[19] = dest[3]

	v4Checksum := checkSum(packet[:20])
	packet[10] = byte(v4Checksum >> 8)
	packet[11] = byte(v4Checksum)

	return packet
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
