package main

import (
	"log"

	"github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thomasjinlo/gochatter-client/internal/api"
	"github.com/thomasjinlo/gochatter-client/internal/tui"
	"github.com/thomasjinlo/gochatter-client/internal/ws"
)

func main() {
	zone.NewGlobal()
	wc := ws.NewClient()
	api := api.NewClient()
	cm := tui.NewChatModel(wc, api)
	lm := tui.NewLoginModel(wc, api, cm)
	p := tea.NewProgram(lm, tea.WithMouseCellMotion())

	// Schedule messages from websocket server
	go func() {
		for {
			select {
			case msg, ok := <-wc.MsgCh:
				if !ok {
					p.Kill()
				}
				p.Send(msg)
			}
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
