package main

import (
	"log"
	"os"

	"github.com/charmbracelet/bubbletea"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thomasjinlo/gochatter-client/internal/api"
	"github.com/thomasjinlo/gochatter-client/internal/push"
	"github.com/thomasjinlo/gochatter-client/internal/tui"
)

func main() {
	var localhost bool
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-l", "--localhost":
			localhost = true
		}
	}
	pushServerEndpoint := "wss://ws.gochatter.app:2053"
	apiEndpoint := "https://api.gochatter.app:8443"
	if localhost {
		pushServerEndpoint = "ws://localhost:8444"
		apiEndpoint = "http://localhost:8443"
	}

	ps := push.NewServer(pushServerEndpoint)

	zone.NewGlobal()
	api := api.NewClient(apiEndpoint)
	cm := tui.NewChatModel(api)
	lm := tui.NewLoginModel(ps, api, cm)
	p := tea.NewProgram(lm, tea.WithMouseCellMotion())

	ps.Subscribe(func(msg push.Message, ok bool) {
		if !ok {
			p.Quit()
		}

		p.Send(msg)
	})

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
