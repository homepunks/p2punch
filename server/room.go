package server

import (
	"errors"
	"net"
)

var ErrFull = errors.New("a room cannot hold more than 2 peers")

type room struct {
	peers [2]*net.UDPAddr
	count int
	done  bool
}

func (r *room) add(addr *net.UDPAddr) error {
	for i := 0; i < r.count; i++ {
		if r.peers[i].String() == addr.String() {
			return nil
		}
	}
	if r.count == 2 {
		return ErrFull
	}
	r.peers[r.count] = addr
	r.count++
	return nil
}

func (r *room) exchange(conn *net.UDPConn) error {
	a, b := r.peers[0], r.peers[1]
	if _, err := conn.WriteToUDP([]byte(a.String()), b); err != nil {
		return err
	}
	if _, err := conn.WriteToUDP([]byte(b.String()), a); err != nil {
		return err
	}
	r.done = true
	return nil
}
