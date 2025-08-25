package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	// Server address
	serverAddr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatal("Error resolving server addr:", err)
	}

	// local client addr
	clientAddr, err := net.ResolveUDPAddr("udp", "localhost:9090")
	if err != nil {
		log.Fatal("Error resolving client addr:", err)
	}

	// local udp socket
	conn, err := net.DialUDP("udp", clientAddr, serverAddr)
	if err != nil {
		log.Fatal("Error resolving client addr:", err)
	}
	defer conn.Close()
	log.Printf("Connected to UDP server at %s:%s", serverAddr, clientAddr)

	// Step 2: listen for messages
	buffer := make([]byte, 1024)
	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}
		fmt.Println("Received from server:", string(buffer[:n]))
	}
}
