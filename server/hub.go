package server

import (
	"net"
	"sync"
)

type Hub struct {
	mu    sync.Mutex
	rooms map[string]*room
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string]*room)}
}

func (h *Hub) Join(name string, addr *net.UDPAddr) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	r, ok := h.rooms[name]
	if !ok {
		r = &room{}
		h.rooms[name] = r
	}
	if err := r.add(addr); err != nil {
		return 0, err
	}
	return r.count, nil
}

type Exchange struct {
	Room string
	A    string
	B    string
}

func (h *Hub) Ready(conn *net.UDPConn) []Exchange {
	h.mu.Lock()
	defer h.mu.Unlock()

	var exchanged []Exchange
	for name, r := range h.rooms {
		if r.done || r.count != 2 {
			continue
		}
		if err := r.exchange(conn); err != nil {
			continue
		}
		exchanged = append(exchanged, Exchange{
			Room: name,
			A:    r.peers[0].String(),
			B:    r.peers[1].String(),
		})
	}
	return exchanged
}

func Kick(conn *net.UDPConn, addr *net.UDPAddr) {
	conn.WriteToUDP([]byte("KICKED"), addr)
}
