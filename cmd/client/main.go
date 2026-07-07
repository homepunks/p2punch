package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	roomPtr := flag.String("room", "", "room name for you and your peer")
	flag.Parse()

	if *roomPtr == "" {
		fmt.Println("Please specify your room name!")
		return
	}

	conn, err := net.Dial("udp", "127.0.0.1:6969")
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}
	defer conn.Close()

	payload := []byte(*roomPtr)
	maxRetries := 3

	for i := 1; i < maxRetries+1; i++ {
		_, err = conn.Write(payload)
		if err == nil {
			log.Println("successfully sent packets")
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("could not receive peer address: %v\n", err)
				continue
			}

			log.Printf("PEER: %s, SELF: %s\n", string(buf[:n]), conn.LocalAddr())

			break
		}

		log.Printf("attempt %v failed. trying again...\n", i)
		if i < maxRetries-1 {
			time.Sleep(time.Second)
		} else {
			log.Printf("could not send packets after %v attempts: %v\n", i, err)
			log.Println("aborting...")
		}
	}

	time.Sleep(5 * time.Second)
}
