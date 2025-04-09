package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chase-horton/go-icmp/protocol"
)

type IPAddr [4]byte

func saveIpsToFile(ipMap map[IPAddr]protocol.PingResult) {
	// Open a file for writing
	file, err := os.Create("ip_results.txt")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	// Write the IP addresses and their ping results to the file
	for ip, res := range ipMap {
		line := fmt.Sprintf("%d.%d.%d.%d> %s\n", ip[0], ip[1], ip[2], ip[3], res.FileString())
		file.WriteString(line)
	}
	fmt.Println("IP results saved to ip_results.txt")
}
func loadIpsFromFile() map[IPAddr]protocol.PingResult {
	i := 0
	file, err := os.Open("ip_results.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	ipMap := make(map[IPAddr]protocol.PingResult)
	for scanner.Scan() {
		line := scanner.Text()
		var ip IPAddr
		var res protocol.PingResult
		ipStr := strings.Split(line, ">")[0]
		ipParts := strings.Split(ipStr, ".")
		if len(ipParts) != 4 {
			panic(fmt.Sprintf("Invalid IP format: %s", ipStr))
		}
		for i, part := range ipParts {
			partVal, err := strconv.Atoi(part)
			if err != nil {
				panic(fmt.Sprintf("Invalid IP part: %s", part))
			}
			if partVal < 0 || partVal > 255 {
				panic(fmt.Sprintf("IP part out of range: %s", part))
			}
			ip[i] = byte(partVal)
		}
		resStr := strings.TrimSpace(strings.Split(line, ">")[1])
		resultParts := strings.Split(resStr, ",")
		if len(resultParts) != 3 {
			panic(fmt.Sprintf("Invalid result format: %s", resStr))
		}
		res.Success = resultParts[0] == "true"
		res.Duration, err = time.ParseDuration(resultParts[1])
		res.Error = errors.New(resultParts[2])

		if err := scanner.Err(); err != nil {
			panic(fmt.Sprintf("Error reading file: %s", err))
		}
		if err != nil {
			panic(fmt.Sprintf("Error parsing duration: %s", resultParts[1]))
		}
		//fmt.Println(fmt.Sprintf("Loading IP %d: %s", i, ipStr))
		i++
		ipMap[ip] = res
	}
	return ipMap
}
func pingEveryIp() map[IPAddr]protocol.PingResult {
	const (
		numWorkers   = 500
		saveInterval = 10000
	)
	timeStart := time.Now()
	ipMap := loadIpsFromFile()
	//could change this to just be based on length
	initialMap := loadIpsFromFile()
	lastSaved := 0
	mutex := &sync.Mutex{}
	ipsToPing := make(chan IPAddr, 1000)
	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipsToPing {
				res := protocol.TryPing(ip, time.Second*2)
				mutex.Lock()
				ipMap[ip] = res
				lastSaved++
				if lastSaved%saveInterval == 0 {
					saveIpsToFile(ipMap)
					fmt.Println(fmt.Sprintf("Saved %d IPs, took %s", lastSaved, time.Since(timeStart)))
				}
				mutex.Unlock()
			}
		}()
	}

	go func() {
		for a := 0; a < 256; a++ {
			for b := 0; b < 256; b++ {
				for c := 0; c < 256; c++ {
					for d := 0; d < 256; d++ {
						// Skip already pinged IPs
						if _, exists := initialMap[IPAddr{byte(a), byte(b), byte(c), byte(d)}]; exists {
							continue
						}
						ip := IPAddr{byte(a), byte(b), byte(c), byte(d)}
						ipsToPing <- ip
					}
				}
			}
		}
		close(ipsToPing)
	}()

	wg.Wait()
	// Save any remaining IPs to file
	saveIpsToFile(ipMap)

	return ipMap
}
func main() {
	//results := loadIpsFromFile()
	//fmt.Println(len(results))
	results := pingEveryIp()
	saveIpsToFile(results)
	//test
	//ip := [4]byte{142, 250, 114, 133}
	//res := protocol.TryPing(ip, 2*time.Second)
	//fmt.Println(fmt.Sprintf("Ping result for %d.%d.%d.%d -> %s", ip[0], ip[1], ip[2], ip[3], res.String()))
}
