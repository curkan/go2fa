package screens

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)


type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc, code  string
		s = list.NewDefaultItemStyles()
	)

	s.SelectedTitle = s.SelectedTitle.Foreground(lipgloss.Color("#99dd99")).BorderLeftForeground(lipgloss.Color("#7aa37a"))
	s.SelectedDesc = s.SelectedDesc.Foreground(lipgloss.Color("#7aa37a")).BorderLeftForeground(lipgloss.Color("#7aa37a"))

	if i, ok := listItem.(list.DefaultItem); ok {
		title = i.Title()
		desc = i.Description()
		code = "123 123"
	} else {
		return
	}

	// Conditions
	var (
		isSelected  = index == m.Index()
	)

	width, _, _ := term.GetSize(0)
	width = 50
	padding := 0
	if isSelected && m.FilterState() != list.Filtering {
		padding = width - len(title) - len(code) - 5
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		padding = width - len(title) - len(code) - 5
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}


	fmt.Fprintf(w, "%s%*s%s\n%s", title, padding, " ", code, desc)
	// fmt.Printf("%s%*s%s\n%s\n", title, padding, "", code, desc)
}

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

	// delegate := list.NewDefaultDelegate()
	// delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.Foreground(lipgloss.Color("#99dd99")).BorderLeftForeground(lipgloss.Color("#7aa37a"))
	// delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(lipgloss.Color("#7aa37a")).BorderLeftForeground(lipgloss.Color("#7aa37a"))

	m := listKeysModel{list: list.New(itemKeys, ItemDelegate{}, 30, 20)}
	// m := listKeysModel{list: list.New(itemKeys, list.NewDefaultDelegate(), 30, 20)}
	m.list.Title = "Доступные ключи"

	return m
}
