package main

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/homepunks/p2punch/server"
)

const (
	HOST = "0.0.0.0:6969"
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
	var mu sync.Mutex
	roomKeeper := make(map[server.RoomName]*server.Room)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, cAddr, err := ln.ReadFromUDP(buf)
			if err != nil {
				log.Printf("read error: %v", err)
				continue
			}

			roomName := string(buf[:n])
			roomName = strings.TrimSuffix(roomName, "\n")

			peer := server.NewPeer(cAddr)

			mu.Lock()
			room, exists := roomKeeper[roomName]
			if !exists {
				room = new(server.Room)
				roomKeeper[roomName] = room
			}
			err = room.Add(peer)
			n = room.Len()
			mu.Unlock()

			if err != nil {
				log.Printf("Client <%s> tried to join room %s (already closed): %v\n", peer.IP(), roomName, err)
				server.DisconnectClient(ln, peer)
				continue
			}
			log.Printf("Peer <%s> joined. Room %s: [%d/2]\n", peer.IP(), roomName, n)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		mu.Lock()
		for name, r := range roomKeeper {
			if !r.Done() {
				if r.Len() == 2 {
					err := r.ExchangePeers(ln)
					if err == nil {
						log.Printf("%s: peers <%s> and <%s> have exchanged their addresses!\n", name, r.Peers()[0].IP(), r.Peers()[1].IP())
					}
				}
			} else {
				continue
			}
		}
		mu.Unlock()
	}
}
