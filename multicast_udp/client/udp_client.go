package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/net/ipv4"
	"golang.org/x/sys/unix"
)

func main() {
	group := "239.0.0.222" // IPv4 multicast group to join
	port := 9999           // Multicast UDP port
	ifaceName := "eth0"    // Interface to join on (e.g., eth0, en0). If empty, auto-select
	buffSize := 1024       //  Recieve buffer size (bytes)

	ip := net.ParseIP(group)
	if ip == nil || ip.To4() == nil || !ip.IsMulticast() {
		log.Fatal("invalid IPv4 multicast group:", group)
	}

	// Choose interface
	ifi, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("invalid %q not found: %v", ifaceName, err)
	}

	addr := fmt.Sprintf("0.0.0.0:%d", port)
	lc := net.ListenConfig{Control: reuseControl}
	pc, err := lc.ListenPacket(context.Background(), "udp4", addr)
	if err != nil {
		log.Fatalf("ListenPacket(%s): %v", addr, err)
	}
	defer pc.Close()

	p := ipv4.NewPacketConn(pc)
	defer p.Close()

	// join multicast group on the choosen interface
	if err := p.JoinGroup(ifi, &net.UDPAddr{IP: ip}); err != nil {
		log.Fatalf("JoinGroup(iface=%s, group=%s): %v", ifi.Name, ip, err)
	}

	log.Printf("Listening on %s, joined group %s on iface=%s", addr, ip.String(), ifi.Name)
	log.Printf("Press Ctrl+C to stop.")

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Receiving from the channel
	buffer := make([]byte, buffSize)
	for {
		select {
		case <-stop:
			log.Println("client: leaving group and shutting down.")
			_ = p.LeaveGroup(ifi, &net.UDPAddr{IP: ip})
			return
		default:
			n, cm, src, err := p.ReadFrom(buffer)
			if err != nil {
				log.Printf("ReadFrom error: %v", err)
				continue
			}
			dst := ""
			if cm != nil && cm.Dst != nil {
				dst = cm.Dst.String()
			}
			log.Printf("recv %dB from %s -> group=%s: %s\\n", n, src.String(), dst, string(buffer[:n]))
		}
	}
}

func reuseControl(network, address string, c syscall.RawConn) error {
	var err error
	c.Control(func(fd uintptr) {
		// Allow multiple sockets to bind same addr:port
		if e := unix.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); e != nil && err == nil {
			err = e
		}
	})
	return err
}
