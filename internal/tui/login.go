package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/thomasjinlo/gochatter-client/internal/api"
	"github.com/thomasjinlo/gochatter-client/internal/ws"
)

type LoginModel struct {
	ws       *ws.Client
	api      *api.Client
	wm       tea.Msg
	cm       tea.Model
	loginErr bool
	ti       textinput.Model
}

func NewLoginModel(ws *ws.Client, api *api.Client, cm tea.Model) *LoginModel {
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 250
	return &LoginModel{
		ws:  ws,
		api: api,
		cm:  cm,
		ti:  ti,
	}
}

func (lm *LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (lm *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var tiCmd tea.Cmd
	lm.ti, tiCmd = lm.ti.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return lm, tea.Quit
		case tea.KeyEnter:
			username := lm.ti.Value()
			if ok := lm.api.Login(username); !ok {
				// display error message and clear input
				lm.ti.Reset()
				lm.loginErr = true
				return lm, tiCmd
			}
			go lm.ws.Connect(username)
			getUsersCmd := func() tea.Msg { return lm.api.GetUsers() }
			usernameCmd := func() tea.Msg { return UsernameMsg(username) }
			windowCmd := func() tea.Msg { return lm.wm }
			return lm.cm, tea.Batch(windowCmd, usernameCmd, getUsersCmd)
		}
	case tea.WindowSizeMsg:
		lm.wm = msg
	}
	return lm, tiCmd
}

func (lm *LoginModel) View() string {
	if lm.loginErr {
		return fmt.Sprintf(
			"Login failed. Re-enter your desired username: \n\n%s\n\n%s",
			lm.ti.View(),
			"(esc to quit)",
		) + "\n"
	}

	return fmt.Sprintf(
		"Enter your desired username: \n\n%s\n\n%s",
		lm.ti.View(),
		"(esc to quit)",
	) + "\n"
}
