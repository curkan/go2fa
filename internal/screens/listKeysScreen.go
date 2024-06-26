package screens

import (
	"encoding/json"
	"fmt"
	"go2fa/internal/crypto"
	"go2fa/internal/structure"
	"go2fa/internal/twofactor"
	"io"
	"os"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

var globalCopied = false

type itemKey struct {
	title string
	desc string
	secret string
}

type ItemDelegate struct{}
type tickMsg struct{}


func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc, code, secret string
		exp int64
		s = list.NewDefaultItemStyles()
	)

	s.SelectedTitle = s.SelectedTitle.Foreground(lipgloss.Color("#99dd99")).BorderLeftForeground(lipgloss.Color("#7aa37a"))
	s.SelectedDesc = s.SelectedDesc.Foreground(lipgloss.Color("#7aa37a")).BorderLeftForeground(lipgloss.Color("#7aa37a")).MarginBottom(1)
	s.NormalDesc = s.NormalDesc.MarginBottom(1)

	if i, ok := listItem.(itemKey); ok {
		title = i.title
		desc = i.desc
		secret = i.secret
		code, exp = twofactor.GenerateTOTP(secret)
	} else {
		return
	}

	// Conditions
	var (
		isSelected  = index == m.Index()
	)

	if !isSelected {
		code = "******"
	}

	if globalCopied {
		s.SelectedTitle = s.SelectedTitle.BorderLeftBackground(lipgloss.Color("#7aa37a"))
		s.SelectedDesc = s.SelectedDesc.BorderLeftBackground(lipgloss.Color("#7aa37a"))
	}

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


func (i itemKey) Title() string       { return i.title }
func (i itemKey) Description() string { return i.desc }
func (i itemKey) FilterValue() string { return i.title }

type listKeysModel struct {
	list list.Model
	itemsKeysList []structure.TwoFactorItem
}

func (m listKeysModel) Init() tea.Cmd {
	return tick()
}

func (m listKeysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
	}
	switch msg := msg.(type) {

	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
				case "q", "ctrl+c":
					output.ClearScreen()

					return m, tea.Quit
				case "d": 
					item, ok := m.list.SelectedItem().(itemKey)

					if !ok {
						return m, tick()
					}

					twoFactorItem := structure.TwoFactorItem{
						Title: item.title,
						Desc: item.desc,
						Secret: item.secret,
					}

					screen := ScreenDeleteKey(m.itemsKeysList, twoFactorItem)
					return RootScreen().SwitchScreen(&screen)
				}
		}

		switch msg.Type {
			case tea.KeyCtrlC:
				output.ClearScreen()
				return m, tea.Quit

			case tea.KeyEsc:
				screen := ListMethodsScreen()
				return RootScreen().SwitchScreen(&screen)

			case tea.KeyEnter:
				if m.list.FilterState() != list.Filtering {
					item, ok := m.list.SelectedItem().(itemKey)

					if !ok {
						return m, tick()
					}

					code, _ := twofactor.GenerateTOTP(item.secret)
					clipboard.WriteAll(code)
					globalCopied = true
				}
			default:
				globalCopied = false
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
	var itemKeysList []structure.TwoFactorItem
	vault := crypto.GetDataVault()

	if len(vault.Db) > 0 {
		err := json.Unmarshal([]byte(vault.Db), &itemKeysList)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	var itemKeys []list.Item

	for _, item := range itemKeysList {
		itemKeys = append(itemKeys, itemKey{
			title: item.Title,
			desc:  item.Desc,
			secret: item.Secret,
		})
	}

	m := listKeysModel{
		list: list.New(itemKeys, ItemDelegate{}, 30, 20),
		itemsKeysList: itemKeysList,
	}

	m.list.Title = "Active Keys"

	return m
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

