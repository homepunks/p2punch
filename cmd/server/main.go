package main

import (
	"errors"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	HOST = "127.0.0.1:6969"
)

var ErrFull = errors.New("a room cannot hold more than 2 peers")

type Peer struct {
	ip *net.UDPAddr
}

type Room struct {
	peers [2]*Peer
	count int
}

type RoomName = string

func (r *Room) Add(p *Peer) error {
	if r.count >= 2 {
		return ErrFull
	}

	r.peers[r.count] = p
	r.count++
	return nil
}

func (r *Room) Len() int {
	return r.count
}

func (r *Room) ExchangePeers(ln *net.UDPConn) error {
	peerA := r.peers[0]
	peerB := r.peers[1]
	addrPeerA := []byte(peerA.IP())
	addrPeerB := []byte(peerB.IP())

	_, err := ln.WriteToUDP(addrPeerA, peerB.ip)
	if err != nil {
		return err
	}
	_, err = ln.WriteToUDP(addrPeerB, peerA.ip)
	if err != nil {
		return err
	}

	return nil
}

func NewPeer(ip *net.UDPAddr) *Peer {
	return &Peer{
		ip: ip,
	}
}

func (p *Peer) IP() string {
	return p.ip.String()
}

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
	roomKeeper := make(map[RoomName]*Room)

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
			mu.Lock()
			room, exists := roomKeeper[roomName]
			mu.Unlock()
			if !exists {
				room := new(Room)
				peer := NewPeer(cAddr)
				room.Add(peer)
				mu.Lock()
				roomKeeper[roomName] = room
				mu.Unlock()
				log.Printf("Peer 1 <%s> joined. Room %s: %d/2\n", peer.IP(), roomName, room.Len())

				continue
			}

			peer := NewPeer(cAddr)
			err = room.Add(peer)
			if err != nil {
				log.Printf("Client <%s> tried to join room %s: %v\n", peer.IP(), roomName, err)
				continue
			}
			log.Printf("Peer 2 <%s> joined. Room %s: %d/2\n", peer.IP(), roomName, room.Len())
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for range ticker.C {
		mu.Lock()
		for _, r := range roomKeeper {
			if r.Len() == 2 {
				r.ExchangePeers(ln)
			}
		}
		mu.Unlock()
	}
}
