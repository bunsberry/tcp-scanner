package main

import (
	"flag"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"net"
	"sort"
	"time"
)

func worker(ports, results chan int, host string, timeout time.Duration) {
	for p := range ports {
		address := fmt.Sprintf("%s:%d", host, p)
		conn, err := net.DialTimeout("tcp", address, timeout)
		if err != nil {
			results <- 0
			continue
		}
		conn.Close()
		results <- p
	}
}

func main() {

	// CMD Arguments
	host := flag.String("host", "", "Host to scan (required)")
	max_port := flag.Int("max-port", 1024, "Max port to scan")
	channels := flag.Int("channels", 100, "Number of channels to scan concurrently")
	timeout := flag.Duration("timeout", 5_000_000_000, "Timeout for one port connection")
	flag.Parse()

	if len(*host) == 0 {
		fmt.Println("[!] Host was not specified - aborting...")
		return
	}

	// actual logic
	ports := make(chan int, *channels)
	results := make(chan int)
	var openports []int

	for i := 0; i < cap(ports); i++ {
		go worker(ports, results, *host, *timeout)
	}

	go func() {
		for i := 1; i <= *max_port; i++ {
			ports <- i
		}
	}()

	bar := progressbar.Default(int64(*max_port))

	for i := 0; i < *max_port; i++ {
		port := <-results
		bar.Add(1)
		if port != 0 {
			openports = append(openports, port)
		}
	}

	close(ports)
	close(results)
	sort.Ints(openports)
	fmt.Printf("---------\nScan finished with:\n")
	for _, port := range openports {
		fmt.Printf("Open port on %d\n", port)
	}
}
