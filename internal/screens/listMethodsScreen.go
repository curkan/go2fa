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
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#54B575"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#A1FCC0"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item struct{
	alias string
	title string
}

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

	str := fmt.Sprintf("%d. %s", index+1, i.title)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("→ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func ListMethodsScreen() ListMethodsModel {
	items := []list.Item{
		item{ title: "Show keys", alias: "show_keys" },
		item{ title: "Add key", alias: "add_key" },
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "GO2FA"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	output := termenv.NewOutput(os.Stdout)
	return ListMethodsModel{list: l, output: output}
}

type ListMethodsModel struct {
	list     list.Model
	quitting bool
	output   *termenv.Output
}

func (m ListMethodsModel) Init() tea.Cmd {
	return tea.SetWindowTitle("Go2FA")
}

func (m ListMethodsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				item, ok := m.list.SelectedItem().(item)
				if ok {
					if item.alias == "add_key" {
						screen_y := ScreenInputSecret()
						return RootScreen().SwitchScreen(&screen_y)
					}

					if item.alias == "show_keys" {
						screen_y := ListKeysScreen()
						return RootScreen().SwitchScreen(&screen_y)
					}
				}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ListMethodsModel) View() string {
	return "\n" + m.list.View()
}

