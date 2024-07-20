package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	MsgCh chan DirectMessage
}

func NewClient() *Client {
	return &Client{MsgCh: make(chan DirectMessage)}
}

type DirectMessage struct {
	Author  string
	Content string
}

func (c *Client) Connect(username string) {
	h := http.Header{}
	h.Add("AccountId", username)
	conn, _, err := websocket.DefaultDialer.Dial("wss://websockets.gochatter.app:2053/connect", h)
	if err != nil {
		log.Fatalf("[gochatter-client] received error while dialing %v", err)
	}

	go func() { setKeepAlive(conn) }()

	defer conn.Close()

	for {
		var pm DirectMessage
		err := conn.ReadJSON(&pm)
		if err != nil {
			log.Printf("[gochatter-client] received error while reading message %v", err)
			conn, _, _ = websocket.DefaultDialer.Dial("wss://websockets.gochatter.app:2053/connect", h)
		} else {
			c.MsgCh <- pm
		}
	}
}

func setKeepAlive(conn *websocket.Conn) {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			err := conn.WriteMessage(websocket.PingMessage, []byte{})
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
