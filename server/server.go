package server

import (
	"errors"
	"net"
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

func (r *Room) Done() bool {
	return r.done
}

func (r *Room) Peers() [2]*Peer {
	return r.peers
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

func DisconnectClient(ln *net.UDPConn, p *Peer) {
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
