package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	// first resolve the address for the protocol
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
	}

	// then we start listening
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	fmt.Println("UDP server started on", addr.String())

	// Getting client address
	clientAddr, err := net.ResolveUDPAddr("udp", "localhost:9090")
	if err != nil {
		fmt.Println("Error resolving client addr:", err)
		return
	}

	done := make(chan bool)

	go func(addr *net.UDPAddr) {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		count := 0
		for range ticker.C {
			msg := fmt.Sprintf("Server message #%d", count)
			_, err := conn.WriteToUDP([]byte(msg), addr)
			if err != nil {
				fmt.Println("Error sending:", err)
				return
			}
			fmt.Printf("Send to %v: %s\n", addr, msg)
			count++
		}
	}(clientAddr)

	<-done
}
