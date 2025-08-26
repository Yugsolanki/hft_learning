package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"golang.org/x/net/ipv4"
)

func main() {
	group := "239.0.0.222"         // IPv4 multicast group (224.0.0.0 - 239.255.255.255)
	port := 9999                   // UDP port
	ifaceName := "Wi-Fi"           // Interface name to send on (e.g., eth0, en0). If empty, auto-select.
	ttl := 1                       // Multicast TTL (hops)
	interval := time.Second        // Send interval
	loop := true                   // Loop back multicast to local host
	payload := "hello from server" // Message payload

	// Parse IP address
	ip := net.ParseIP(group)
	if ip == nil || ip.To4() == nil || !ip.IsMulticast() {
		log.Printf("invalid IPv4 multicast group: %v", group)
		os.Exit(1)
	}

	// Set interface
	ifi, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Printf("interface %v not found: %v", group, err)
		os.Exit(1)
	}

	// destination (multicast) address
	dst := &net.UDPAddr{
		IP:   ip,
		Port: port,
	}

	pc, err := net.ListenPacket("udp4", "0.0.0.0:0") // udp4 = strictly setting ipv4
	if err != nil {
		log.Fatal("ListenPacket:", err)
	}
	defer pc.Close()

	// converting to ipv4 connection
	p := ipv4.NewPacketConn(pc)
	defer p.Close()

	// Configure multicast options
	if err := p.SetMulticastInterface(ifi); err != nil {
		log.Fatalf("SetMulticastInterface(%s): %v", ifi.Name, err)
	}
	if err := p.SetMulticastTTL(ttl); err != nil {
		log.Fatalf("SetMulticastTTL(%d): %v", ttl, err)
	}
	if err := p.SetMulticastLoopback(loop); err != nil {
		log.Fatalf("SetMulticastLoopback(%v): %v", loop, err)
	}

	log.Printf("Sending to %s:%d via iface=%s ttl=%d loopback=%v every %v", ip.String(), port, ifi.Name, ttl, loop, interval)

	// Grateful shutfown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Ticker that send a tick every second to it's channel
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Sending data
	i := 0
	for {
		select {
		case <-ticker.C:
			i++
			msg := fmt.Sprintf("%s | #%d", payload, i)
			n, err := p.WriteTo([]byte(msg), nil, dst)
			if err != nil {
				log.Printf("WriteTo error: %v", err)
				continue
			}
			log.Printf("sent %d bytes -> %s", n, dst.String())
		case <-stop:
			log.Println("server: shutting down")
			return
		}
	}
}
