package tui

import (
	"fmt"
	// "math"
	"regexp"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
	"github.com/thomasjinlo/gochatter-client/internal/api"
	"github.com/thomasjinlo/gochatter-client/internal/ws"
)

type ChatModel struct {
	username  string
	liFocused bool
	msgCh     chan ws.DirectMessage
	dms       map[string]string
	ws        *ws.Client
	api       *api.Client
	li        list.Model
	ta        textarea.Model
	vp        viewport.Model
	bs        lipgloss.Style
	ms        lipgloss.Style
}

type UsernameMsg string
type displayDMMsg string

func NewChatModel(wsClient *ws.Client, apiClient *api.Client) tea.Model {
	ta := textarea.New()
	ta.Prompt = "> "
	ta.CharLimit = 250
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	ta.KeyMap.InsertNewline.Enabled()
	ta.Focus()

	var items []list.Item
	li := list.New(items, list.NewDefaultDelegate(), 15, 0)
	li.SetFilteringEnabled(false)
	li.FilterInput.Blur()

	// set width and height according to screensize in #Update
	vp := viewport.New(0, 0)

	return &ChatModel{
		ta:  ta,
		vp:  vp,
		li:  li,
		ws:  wsClient,
		api: apiClient,
		ms:  lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		bs:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()),
		dms: make(map[string]string),
	}
}

func (m *ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		taCmd tea.Cmd
		vpCmd tea.Cmd
		liCmd tea.Cmd
	)

	m.ta, taCmd = m.ta.Update(msg)
	m.vp, vpCmd = m.vp.Update(msg)
	cmds := []tea.Cmd{taCmd, vpCmd}
	if m.liFocused {
		m.li, liCmd = m.li.Update(msg)
		cmds = append(cmds, liCmd)
	}

	switch msg := msg.(type) {
	case UsernameMsg:
		m.username = string(msg)
		m.vp.SetContent(m.username)
	case []api.User:
		var items []list.Item
		for _, u := range msg {
			if u.Name == m.username {
				continue
			}
			items = append(items, &userItem{id: u.Name, title: u.Name})
		}
		m.li.SetItems(items)
		cmds = append(cmds, m.pollUsers())
	case ws.DirectMessage:
		if messages, ok := m.dms[msg.Author]; ok {
			m.dms[msg.Author] = messages + "\n" + fmt.Sprintf("%s: %s", msg.Author, msg.Content)
		} else {
			m.dms[msg.Author] = fmt.Sprintf("%s: %s", msg.Author, msg.Content)
		}
		if m.li.SelectedItem().FilterValue() == msg.Author {
			cmds = append(cmds, m.displayDMCmd(msg.Author))
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.ta.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			if selectedUser := m.li.SelectedItem(); selectedUser != nil {
				targetUserId := selectedUser.FilterValue()
				m.api.SendDirectMessage(m.username, targetUserId, m.ta.Value())
				m.dms[targetUserId] = m.dms[targetUserId] + "\n" + fmt.Sprintf("You: %s", m.ta.Value())
				cmds = append(cmds, m.displayDMCmd(targetUserId))
			}
			m.ta.Reset()
			m.vp.GotoBottom()
		case tea.KeyRunes:
			re := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
			if !re.MatchString(string(msg.Runes)) {
				return m, nil
			}
		}
	case tea.WindowSizeMsg:
		m.vp.Height = msg.Height - m.ta.Height() - 4
		m.vp.Width = msg.Width - m.li.Width()
		m.ta.SetWidth(msg.Width - m.li.Width())
		m.li.SetHeight(m.vp.Height + m.ta.Height() + 2)
	case tea.MouseMsg:
		m.liFocused = msg.X <= m.li.Width()
		if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
			m.vp.SetContent(fmt.Sprintf("INSIDE release %v", len(m.li.VisibleItems())))
			for i, item := range m.li.VisibleItems() {
				v, ok := item.(*userItem)
				m.vp.SetContent(fmt.Sprintf("Is OKay? %v %v", ok, v))
				// Check each item to see if it's in bounds.
				if zone.Get(v.id).InBounds(msg) {
					// If so, select it in the list.
					m.vp.SetContent("Found zone")
					m.li.Select(i)
					break
				}
			}
			return m, nil
		}
		switch tea.MouseEvent(msg).Action {
		case tea.MouseActionPress:
			if m.liFocused {
				// if msg.Y < 5 {
				// 	return m, nil
				// }
				if msg.Action == tea.MouseActionRelease && msg.Button == tea.MouseButtonLeft {
					for i, item := range m.li.VisibleItems() {
						v, _ := item.(*userItem)
						// Check each item to see if it's in bounds.
						if zone.Get(v.id).InBounds(msg) {
							// If so, select it in the list.
							m.li.Select(i)
							break
						}
					}
				}
				return m, nil

				// targetCursor := (msg.Y - 5) / 3
				// currCursor := m.li.Cursor()
				// diff := math.Abs(float64(targetCursor - currCursor))
				// for i := 0; i < int(diff); i++ {
				// 	if currCursor < targetCursor {
				// 		m.li.CursorDown()
				// 	} else {
				// 		m.li.CursorUp()
				// 	}
				// }
				// cmds = append(cmds, m.displayDMCmd(m.li.SelectedItem().FilterValue()))
			}
		}
	case displayDMMsg:
		m.vp.SetContent(m.dms[string(msg)])
	}
	return m, tea.Batch(cmds...)
}

func (m *ChatModel) View() string {
	left := zone.Scan(m.bs.Render(m.li.View()))
	right := lipgloss.JoinVertical(lipgloss.Top,
		m.bs.Render(m.vp.View()),
		m.bs.Render(m.ta.View()))
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m *ChatModel) pollUsers() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(5 * time.Second)
		return m.api.GetUsers()
	}
}

func (m *ChatModel) displayDMCmd(targetUser string) tea.Cmd {
	return func() tea.Msg {
		return displayDMMsg(targetUser)
	}
}

type userItem struct {
	id    string
	title string
}

func (i *userItem) Title() string       { return zone.Mark(i.id, i.title) }
func (i *userItem) Description() string { return "" }
func (i *userItem) FilterValue() string { return zone.Mark(i.id, i.title) }
