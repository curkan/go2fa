package screens

import (
	"fmt"
	"go2fa/internal/deletekey"
	"go2fa/internal/structure"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type deleteKeyScreen struct {
	TwoFactorItem structure.TwoFactorItem
	itemKeysList []structure.TwoFactorItem
}

var (
	warning = lipgloss.NewStyle().Bold(true).MarginTop(1).Padding(0,2).Foreground(lipgloss.Color("#FFB775"))
	help = lipgloss.NewStyle().Bold(false).Padding(0,2).Foreground(lipgloss.Color("#D2D2D2"))
)

func ScreenDeleteKey(itemsKeysList []structure.TwoFactorItem, TwoFactorItem structure.TwoFactorItem) deleteKeyScreen {
	return deleteKeyScreen{
		TwoFactorItem:TwoFactorItem,
		itemKeysList: itemsKeysList,
	}
}

func (m deleteKeyScreen) Init() tea.Cmd {
	return nil
}

func (m deleteKeyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		switch msg.String() {
			case "q", "ctrl+c":
				output.ClearScreen()

				return m, tea.Quit

			case "enter":

				result := deletekey.DeleteKey(&m.itemKeysList, m.TwoFactorItem)

				if result {
					screen := ListKeysScreen()
					return RootScreen().SwitchScreen(&screen)
				}
			case "esc":
				screen := ListKeysScreen()
				return RootScreen().SwitchScreen(&screen)
			}
		}
	var cmd tea.Cmd
	return m, cmd
}

func (m deleteKeyScreen) View() string {
	var b strings.Builder
	b.WriteString(warning.Render(fmt.Sprintf("Are you sure you want to delete %s?", m.TwoFactorItem.Title)))

	fmt.Fprintf(&b, "\n\n")
	fmt.Fprintf(&b, "%s\n", help.Render("To confirm - [Enter]\nTo cancel - [Esc]"))

	return b.String()
}

