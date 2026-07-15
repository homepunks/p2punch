package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/homepunks/p2punch/client"
)

func main() {
	roomPtr := flag.String("room", "", "room name for you and your peer")
	addrPtr := flag.String("addr", "", "ip of the rendezvous server")
	flag.Parse()

	if *roomPtr == "" {
		fmt.Println("Please specify your room name!")
		return
	}

	if *addrPtr == "" {
		fmt.Println("Please specify your rendezvous server's IP!")
		return
	}

	host := *addrPtr + ":6969"
	serverAddr, err := net.ResolveUDPAddr("udp", host)
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

			if err = client.Punch(conn, peerAddr); err != nil {
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
		client.Session(conn, peerAddr)
	}
}
