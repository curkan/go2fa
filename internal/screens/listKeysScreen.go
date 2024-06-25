package screens

import (
	"fmt"
	"go2fa/internal/twofactor"
	"io"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)


type ItemDelegate struct{}
type tickMsg struct{}


func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc, code  string
		exp int64
		s = list.NewDefaultItemStyles()
	)

	s.SelectedTitle = s.SelectedTitle.Foreground(lipgloss.Color("#99dd99")).BorderLeftForeground(lipgloss.Color("#7aa37a"))
	s.SelectedDesc = s.SelectedDesc.Foreground(lipgloss.Color("#7aa37a")).BorderLeftForeground(lipgloss.Color("#7aa37a"))

	if i, ok := listItem.(list.DefaultItem); ok {
		title = i.Title()
		desc = i.Description()
		// code = twofactor.GeneratePassCode(title)
		code, exp = twofactor.GenerateTOTP(title)
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

	until := exp - time.Now().Unix()

	codeStyle := lipgloss.NewStyle().Bold(true).Blink(true)

	if until <= 15 && until > 5 {
		codeStyle = codeStyle.Foreground(lipgloss.Color("#FFF0A1"))
	}

	if until <= 5 {
		codeStyle = codeStyle.Foreground(lipgloss.Color("#FF7575"))
	}

	if isSelected && m.FilterState() != list.Filtering {
		padding = width - len(title) - len(code) - 5
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		padding = width - len(title) - len(code) - 5
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}


	code = codeStyle.Render(code)

	fmt.Fprintf(w, "%s%*s%s %ds\n%s", title, padding, " ", code, until, desc)
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
	return tick()
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
	case tickMsg:
		return m, tick()
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

	m := listKeysModel{list: list.New(itemKeys, ItemDelegate{}, 30, 20)}
	m.list.Title = "Доступные ключи"

	return m
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

