package screens

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
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
	focusedButton = focusedStyle.Padding(0,2).Render("[ Добавить ]")
	blurredButton = blurredStyle.Padding(0,2).Render("[Добавить]")

)

type (
	errMsg error
)

type screenInputSecret struct {
	focusIndex int
	textInputs     []textinput.Model
	cursorMode cursor.Mode
	err       error
}

func ScreenInputSecret() screenInputSecret {
	m := screenInputSecret{
		textInputs: make([]textinput.Model, 3),
	}

	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Название"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Описание"
			t.CharLimit = 128
		case 2:
			t.Placeholder = "SecretKey"
			t.CharLimit = 64
		}

		m.textInputs[i] = t
	}

	return m
}

func (m screenInputSecret) Init() tea.Cmd {
	return textinput.Blink
}

func (m screenInputSecret) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	output := termenv.NewOutput(os.Stdout)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
			case tea.KeyEsc:
				screen := ListMethodsScreen()
				screen.list.Select(1)
				return RootScreen().SwitchScreen(&screen)

			case tea.KeyCtrlC:
				output.ClearScreen()
				return m, tea.Quit

		}

		switch msg.String() {
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				if s == "enter" && m.focusIndex == len(m.textInputs) {
					output.ClearScreen()
					return m, tea.Quit
				}

				// Cycle indexes
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > len(m.textInputs) {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = len(m.textInputs)
				}

				cmds := make([]tea.Cmd, len(m.textInputs))
				for i := 0; i <= len(m.textInputs)-1; i++ {
					if i == m.focusIndex {
						// Set focused state
						cmds[i] = m.textInputs[i].Focus()
						m.textInputs[i].PromptStyle = focusedStyle
						m.textInputs[i].TextStyle = focusedStyle
						continue
					}
					// Remove focused state
					m.textInputs[i].Blur()
					m.textInputs[i].PromptStyle = noStyle
					m.textInputs[i].TextStyle = noStyle
				}

				return m, tea.Batch(cmds...)

				// secretKey := m.textInput.Value()
				// fmt.Print(secretKey)
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	cmd := m.updateInputs(msg)

	return m, cmd
}
func (m *screenInputSecret) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))

	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func (m screenInputSecret) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().MarginTop(1).MarginLeft(2).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Padding(0, 5, 0, 5).Render(fmt.Sprintf("Добавление ключа")))

	fmt.Fprintf(&b, "\n\n")

	for i := range m.textInputs {
		b.WriteString(lipgloss.NewStyle().Padding(0,2).Render(m.textInputs[i].View()))
		// b.WriteString(m.textInputs[i].View())
		if i < len(m.textInputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.textInputs) {
		button = &focusedButton
	}

	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	return b.String()
}
