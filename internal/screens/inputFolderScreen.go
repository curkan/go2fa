package screens

import (
	"fmt"
	"go2fa/internal/storage"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

type inputFolderMode int

const (
	inputFolderCreate inputFolderMode = iota
	inputFolderRename
)

type inputFolderScreen struct {
	input    textinput.Model
	mode     inputFolderMode
	targetID string
	err      string
}

// ScreenCreateFolder opens the create-folder form.
func ScreenCreateFolder() inputFolderScreen {
	ti := textinput.New()
	ti.Placeholder = "Folder name"
	ti.CharLimit = 32
	ti.Focus()
	ti.PromptStyle = focusedStyle
	ti.TextStyle = focusedStyle
	return inputFolderScreen{input: ti, mode: inputFolderCreate}
}

// ScreenRenameFolder opens the rename form pre-populated with currentName.
func ScreenRenameFolder(folderID, currentName string) inputFolderScreen {
	ti := textinput.New()
	ti.Placeholder = "Folder name"
	ti.CharLimit = 32
	ti.SetValue(currentName)
	ti.Focus()
	ti.PromptStyle = focusedStyle
	ti.TextStyle = focusedStyle
	return inputFolderScreen{input: ti, mode: inputFolderRename, targetID: folderID}
}

func (m inputFolderScreen) Init() tea.Cmd { return textinput.Blink }

func (m inputFolderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		switch msg.Type {
		case tea.KeyCtrlC:
			output.ClearScreen()
			return m, tea.Quit
		case tea.KeyEsc:
			screen := ListFoldersScreen()
			return RootScreen().SwitchScreen(&screen)
		case tea.KeyEnter:
			name := strings.TrimSpace(m.input.Value())
			if name == "" {
				m.err = "folder name must not be empty"
				return m, nil
			}

			store, err := storage.LoadStore()
			if err != nil {
				m.err = err.Error()
				return m, nil
			}

			switch m.mode {
			case inputFolderCreate:
				if _, err := storage.NewFolder(&store, name); err != nil {
					m.err = err.Error()
					return m, nil
				}
			case inputFolderRename:
				if err := storage.RenameFolder(&store, m.targetID, name); err != nil {
					m.err = err.Error()
					return m, nil
				}
			}

			if err := storage.SaveStore(store); err != nil {
				m.err = err.Error()
				return m, nil
			}

			screen := ListFoldersScreen()
			return RootScreen().SwitchScreen(&screen)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m inputFolderScreen) View() string {
	var b strings.Builder
	heading := "New folder"
	if m.mode == inputFolderRename {
		heading = "Rename folder"
	}
	b.WriteString(lipgloss.NewStyle().MarginTop(1).MarginLeft(2).
		Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).
		Padding(0, 5, 0, 5).Render(heading))

	fmt.Fprintf(&b, "\n\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(m.input.View()))
	fmt.Fprintf(&b, "\n\n%s\n", folderHelp.Render("[Enter] save   [Esc] cancel"))

	if m.err != "" {
		fmt.Fprintf(&b, "\n%s\n", errorText.Render(m.err))
	}
	return b.String()
}
