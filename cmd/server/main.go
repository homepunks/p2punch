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
	done  bool
}

type RoomName = string

func (r *Room) Add(p *Peer) error {
	for i := 0; i < r.count; i++ {
		if r.peers[i].IP() == p.IP() {
			return nil /* already registered */
		}
	}

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

	r.done = true
	return nil
}

func disconnectClient(ln *net.UDPConn, p *Peer) {
	msg := []byte("KICKED")
	ln.WriteToUDP(msg, p.ip)
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

			peer := NewPeer(cAddr)

			mu.Lock()
			room, exists := roomKeeper[roomName]
			if !exists {
				room = new(Room)
				roomKeeper[roomName] = room
			}
			err = room.Add(peer)
			n = room.Len()
			mu.Unlock()

			if err != nil {
				log.Printf("Client <%s> tried to join room %s (already closed): %v\n", peer.IP(), roomName, err)
				disconnectClient(ln, peer)
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
			if !r.done {
				if r.Len() == 2 {
					err := r.ExchangePeers(ln)
					if err == nil {
						log.Printf("%s: peers <%s> and <%s> have exchanged their addresses!\n", name, r.peers[0].IP(), r.peers[1].IP())
					}
				}
			} else {
				continue
			}
		}
		mu.Unlock()
	}
}
