package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func Punch(conn *net.UDPConn, peer *net.UDPAddr) error {
	stop := make(chan struct{})
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				conn.WriteToUDP([]byte("SYN"), peer)
			}
		}
	}()

	buf := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			close(stop)
			return err
		}

		if from.IP.Equal(peer.IP) && from.Port == peer.Port {
			close(stop)
			log.Printf("punched, got %q from <%s>", buf[:n], from)
			return nil
		}
	}
}

func Session(conn *net.UDPConn, peer *net.UDPAddr) {
	conn.SetReadDeadline(time.Time{})

	done := make(chan struct{})
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		for {
			n, from, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Printf("Could not read: %v\n", err)
				return
			}

			if !from.IP.Equal(peer.IP) || from.Port != peer.Port {
				continue
			}

			msg := string(buf[:n])
			switch msg {
			case "SYN":
				conn.WriteToUDP([]byte("ACK"), peer)
				continue
			case "ACK", "PING":
				continue
			}

			fmt.Printf("\rpeer> %s\nyou> ", msg)
		}
	}()

	go func() {
		t := time.NewTicker(20 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-done:
				return
			case <-t.C:
				conn.WriteToUDP([]byte("PING"), peer)
			}
		}
	}()

	fmt.Println("connected! type messages, ctrl+D to quit")
	fmt.Print("you> ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		select {
		case <-done:
			return
		default:
		}
		text := scanner.Text()
		if text == "" {
			fmt.Print("you> ")
			continue
		}
		if _, err := conn.WriteToUDP([]byte(text), peer); err != nil {
			log.Printf("send error: %v", err)
			return
		}
		fmt.Print("you> ")
	}
}
