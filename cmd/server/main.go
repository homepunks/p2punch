package main

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/homepunks/p2punch/server"
)

const host = "0.0.0.0:6969"

func main() {
	addr, err := net.ResolveUDPAddr("udp", host)
	if err != nil {
		log.Fatalf("could not resolve addr: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
	defer conn.Close()

	log.Printf("listening for UDP connections on %s", host)

	hub := server.NewHub()

	go func() {
		buf := make([]byte, 1024)
		for {
			n, peerAddr, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}

			name := strings.TrimSuffix(string(buf[:n]), "\n")

			count, err := hub.Join(name, peerAddr)
			if err != nil {
				log.Printf("peer <%s> rejected from room %s: %v", peerAddr, name, err)
				server.Kick(conn, peerAddr)
				continue
			}
			log.Printf("peer <%s> joined room %s [%d/2]", peerAddr, name, count)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		for _, ex := range hub.Ready(conn) {
			log.Printf("%s: peers <%s> and <%s> have exchanged their addresses", ex.Room, ex.A, ex.B)
		}
	}
}
