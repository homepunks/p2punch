package main

import (
	"net"
	"log"
	"flag"
	"time"
)

func main() {
	roomPtr := flag.String("room", "", "room name for you and your peer")
	flag.Parse()

	if *roomPtr == "" {
		log.Println("Please specify your room name!")
		return
	}

	conn, err := net.Dial("udp", "127.0.0.1:6969")
	if err != nil {
		log.Fatalf("could not connect: %v\n", err)
	}
	defer conn.Close()

	payload := []byte(*roomPtr)
	maxRetries := 3

	for i := 1; i < maxRetries + 1; i++ {
		_, err = conn.Write(payload)
		if err == nil {
			log.Println("successfully sent packets")
			break
		}

		log.Printf("attempt %v failed. trying again...\n", i)
		if i < maxRetries - 1 {
			time.Sleep(time.Second)
		} else {
			log.Printf("could not send packets after %v attempts: %v\n", i, err)
			log.Println("aborting...")
		}
	}
}
