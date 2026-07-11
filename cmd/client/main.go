package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"
)

func punch(conn *net.UDPConn, peer *net.UDPAddr) error {
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

func session(conn *net.UDPConn, peer *net.UDPAddr) {
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

func main() {
	roomPtr := flag.String("room", "", "room name for you and your peer")
	flag.Parse()

	if *roomPtr == "" {
		fmt.Println("Please specify your room name!")
		return
	}

	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:6969")
	if err != nil {
		log.Fatalf("could not resolve server address: %v\n", err)
	}

	conn, err := net.ListenUDP("udp", &net.UDPAddr{})
	if err != nil {
		log.Fatalf("could not listen udp: %v\n", err)
	}
	defer conn.Close()

	payload := []byte(*roomPtr)
	maxRetries := 3
	var peerAddr *net.UDPAddr

	for i := 1; i < maxRetries+1; i++ {
		_, err = conn.WriteToUDP(payload, serverAddr)
		if err == nil {
			log.Println("successfully sent packets")
			buf := make([]byte, 1024)

			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("could not receive peer address: %v\n", err)
				continue
			}

			if string(buf[:n]) == "KICKED" {
				fmt.Println("ERROR: You cannot join this room, it's already taken")
				return
			}

			peerAddr, err = net.ResolveUDPAddr("udp", string(buf[:n]))
			if err != nil {
				log.Printf("could not resolve peer address; %v\n", err)
				continue
			}

			if err = punch(conn, peerAddr); err != nil {
				log.Printf("punch failed: %v", err)
				return
			}

			log.Println("hole punched - connected to peer!!!")
			break
		}

		log.Printf("attempt %v failed. trying again...\n", i)
		if i < maxRetries {
			time.Sleep(time.Second)
		} else {
			log.Printf("could not send packets after %v attempts: %v\n", i, err)
			log.Println("aborting...")
		}
	}

	if peerAddr != nil {
		session(conn, peerAddr)
	}
}
