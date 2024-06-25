package screens

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Padding(0, 5, 0, 5)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func ScreenOne() model {
	items := []list.Item{
		item("Показать ключи"),
		item("Добавить ключ"),
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "GO2FA"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	output := termenv.NewOutput(os.Stdout)
	m := model{list: l, output: output}

	return m
}

type model struct {
	list     list.Model
	choice   string
	quitting bool
	output   *termenv.Output
}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("Go2FA")
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			m.quitting = true
			m.output.ClearScreen()

			return m, tea.Quit
		}

		switch msg.Type {
			case tea.KeyCtrlC:
				m.quitting = true
				m.output.ClearScreen()

				return m, tea.Quit

			case tea.KeyEnter:
				screen_y := ScreenInputSecret()
				return RootScreen().SwitchScreen(&screen_y)
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != "" {
		return quitTextStyle.Render(fmt.Sprintf("%s? Sounds good to me.", m.choice))
	}

	return "\n" + m.list.View()
}

