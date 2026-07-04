package main

import (
	"net"
	"log"
)

const (
	HOST = "127.0.0.1:6969"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", HOST)
	if err != nil {
		log.Fatalf("could not resolve addr: %v", err)
	}

	ln, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("could not listen: %v", err)
	}
	defer ln.Close()

	log.Printf("Listening UDP connections on %v\n", HOST)
	buf := make([]byte, 1024)
	for {
		n, cAddr, err := ln.ReadFromUDP(buf)
		if err != nil {
			log.Printf("read error: %v", err)
			continue
		}

		log.Printf("Received: %s from %s", string(buf[:n]), cAddr)
	}
}
