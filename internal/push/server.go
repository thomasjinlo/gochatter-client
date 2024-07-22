package push

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Message interface{}
type MessageHandler func(Message, bool)
type Server interface {
	Connect(username string) error
	Subscribe(handler MessageHandler)
}

type DirectMessage struct {
	Author string
	Content string
}

type defaultServer struct {
	endpoint string
	msgCh chan Message
}

func NewServer(endpoint string) *defaultServer{
	return &defaultServer{endpoint: endpoint, msgCh: make(chan Message)}
}

func (s *defaultServer) Connect(username string) error {
	h := http.Header{}
	h.Add("AccountId", username)
	conn, _, err := websocket.DefaultDialer.Dial(s.endpoint + "/connect", h)
	if err != nil {
		log.Fatalf("[gochatter-client] received error while dialing %v", err)
		return err
	}

	go func() {
		defer conn.Close()
		for {
			var pm DirectMessage
			err := conn.ReadJSON(&pm)
			if err != nil {
				log.Printf("[gochatter-client] received error while reading message %v", err)
				conn, _, _ = websocket.DefaultDialer.Dial(s.endpoint, h)
			} else {
				s.msgCh <- pm
			}
		}
	}()

	go func() {
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
	}()

	return nil
}

func (s *defaultServer) Subscribe(handler MessageHandler) {
	go func() {
		for {
			select {
			case msg, ok := <-s.msgCh:
				handler(msg, ok)
			}
		}
	}()
}
