package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type PushMessage struct {
	Author  string
	Content string
}

func main() {
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8444/connect", nil)
	if err != nil {
		log.Fatalf("[gochatter-client] received error while dialing %v", err)
	}

	defer c.Close()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Fatalf("[gochatter-client] received error while reading message %v", err)
		}
		log.Printf("[gochatter-client] received message from %s\n", msg)
	}
}
