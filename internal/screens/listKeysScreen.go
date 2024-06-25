package screens

import (
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type itemKey struct {
	title, desc string
}

func (i itemKey) Title() string       { return i.title }
func (i itemKey) Description() string { return i.desc }
func (i itemKey) FilterValue() string { return i.title }

type listKeysModel struct {
	list list.Model
}

func (m listKeysModel) Init() tea.Cmd {
	return nil
}

func (m listKeysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.Type {
			case tea.KeyEsc:
				screen := ListMethodsScreen()
				return RootScreen().SwitchScreen(&screen)
		}

		switch msg.String() {
		case "q", "ctrl+c":
			output := termenv.NewOutput(os.Stdout)
			output.ClearScreen()
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	}


	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listKeysModel) View() string {
	return docStyle.Render(m.list.View())
}

func ListKeysScreen() listKeysModel {
	itemKeys := []list.Item{
		itemKey{title: "Slack", desc: "https://slack.com"},
		itemKey{title: "Redmine", desc: "Tracker egamings"},
		itemKey{title: "Gitlab vcsx", desc: "http://vcsxa.egamings.com"},
	}

	m := listKeysModel{list: list.New(itemKeys, list.NewDefaultDelegate(), 30, 20)}
	m.list.Title = "Доступные ключи"

	return m
}
