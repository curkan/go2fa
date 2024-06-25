package screens

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	helpStyleInput      = blurredStyle
	focusedButton = focusedStyle.Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

type (
	errMsg error
)

type screenInputSecret struct {
	textInput     textinput.Model
	err       error
}

func ScreenInputSecret() screenInputSecret {
	ti := textinput.New()
	ti.Placeholder = "SecretKey"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return screenInputSecret{
		textInput: ti,
		err:       nil,
	}
}

func (m screenInputSecret) Init() tea.Cmd {
	return textinput.Blink
}

func (m screenInputSecret) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
			case tea.KeyEsc:
				screen := ListMethodsScreen()
				screen.list.Select(1)
				return RootScreen().SwitchScreen(&screen)

			case tea.KeyCtrlC:
				output := termenv.NewOutput(os.Stdout)
				output.ClearScreen()
				return m, tea.Quit

			case tea.KeyEnter:
				secretKey := m.textInput.Value()
				fmt.Print(secretKey)
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m screenInputSecret) View() string {
	return quitTextStyle.Render(
		fmt.Sprintf("Введите Secret Key\n\n%s\n\n%s", m.textInput.View(), "(esc для возврата назад)",) + "\n",
	)
}
