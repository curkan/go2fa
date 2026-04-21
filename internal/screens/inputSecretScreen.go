package screens

// A simple example demonstrating the use of multiple text input components
// from the Bubbles component library.

import (
	"fmt"
	"go2fa/internal/addkey"
	"go2fa/internal/storage"
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
	focusedButton = focusedStyle.Padding(0,2).Bold(true).Render("[Add]")
	blurredButton = blurredStyle.Padding(0,2).Render("[Add]")
	errorText = lipgloss.NewStyle().Padding(0,2).Bold(true).Foreground(lipgloss.Color("#FF7575"))

)

type (
	errMsg error
)

// The form has 5 focusable positions: 3 text inputs (indexes 0..2),
// the folder picker (index 3) and the submit button (index 4).
const (
	formFolderIdx = 3
	formButtonIdx = 4
)

type screenInputSecret struct {
	focusIndex       int
	textInputs       []textinput.Model
	folders          []storage.Folder
	folderIdx        int
	cursorMode       cursor.Mode
	err              error
	error            string
	returnFolderID   string
	returnFolderName string
	returnToFolders  bool
}

// ScreenInputSecret builds the Add key form.
//
// preselectFolderID — highlights this folder in the picker; when empty or
// unknown the Default folder is preselected.
// returnFolderID / returnFolderName — scope to return to on Esc / after save.
// returnToFolders — when true, Esc/save returns to the Folders screen instead
// of a scoped keys list (used when invoked from the Folders landing screen).
func ScreenInputSecret(preselectFolderID, returnFolderID, returnFolderName string, returnToFolders bool) screenInputSecret {
	m := screenInputSecret{
		textInputs:       make([]textinput.Model, 3),
		returnFolderID:   returnFolderID,
		returnFolderName: returnFolderName,
		returnToFolders:  returnToFolders,
	}

	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32

		switch i {
		case 0:
			t.Placeholder = "Title"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Description"
			t.CharLimit = 128
		case 2:
			t.Placeholder = "SecretKey"
			t.CharLimit = 64
		}

		m.textInputs[i] = t
	}

	// Load folders for the picker. If loading fails we silently fall back to
	// a single Default folder so the form is still usable.
	if store, err := storage.LoadStore(); err == nil {
		m.folders = store.Folders
	}
	if len(m.folders) == 0 {
		m.folders = []storage.Folder{{ID: storage.DefaultFolderID, Name: storage.DefaultFolderName}}
	}
	// Preselection: try the requested folder id first, fall back to Default.
	preselectID := preselectFolderID
	if preselectID == "" {
		preselectID = storage.DefaultFolderID
	}
	for i, f := range m.folders {
		if f.ID == preselectID {
			m.folderIdx = i
			break
		}
	}

	return m
}

// returnToOrigin produces the tea.Model transition back to the screen that
// launched this form (either the Folders landing or a scoped keys list).
func (m screenInputSecret) returnToOrigin() (tea.Model, tea.Cmd) {
	if m.returnToFolders {
		s := ListFoldersScreen()
		return RootScreen().SwitchScreen(&s)
	}
	s := ListKeysScreenScoped(m.returnFolderID, m.returnFolderName)
	return RootScreen().SwitchScreen(&s)
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
				return m.returnToOrigin()

			case tea.KeyCtrlC:
				output.ClearScreen()
				return m, tea.Quit

		}

		// Folder picker cycling (left/right) works only while the folder
		// field has focus, so it doesn't hijack normal typing.
		if m.focusIndex == formFolderIdx {
			switch msg.String() {
			case "left", "h":
				if len(m.folders) > 0 {
					m.folderIdx = (m.folderIdx - 1 + len(m.folders)) % len(m.folders)
				}
				return m, nil
			case "right", "l":
				if len(m.folders) > 0 {
					m.folderIdx = (m.folderIdx + 1) % len(m.folders)
				}
				return m, nil
			}
		}

		switch msg.String() {
			case "tab", "shift+tab", "enter", "up", "down":
				s := msg.String()

				if s == "enter" && m.focusIndex == formButtonIdx {
					folderID := ""
					if len(m.folders) > 0 {
						folderID = m.folders[m.folderIdx].ID
					}
					result := addkey.AddKey(m.textInputs, folderID)

					if !result {
						m.focusIndex = 1
						m.textInputs[2].SetValue("")
						m.error = "Only base32 symbols and not empty Title/Secret"
					} else {
						return m.returnToOrigin()
					}

				}

				// Cycle indexes across 5 positions (inputs + folder + button).
				if s == "up" || s == "shift+tab" {
					m.focusIndex--
				} else {
					m.focusIndex++
				}

				if m.focusIndex > formButtonIdx {
					m.focusIndex = 0
				} else if m.focusIndex < 0 {
					m.focusIndex = formButtonIdx
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
	b.WriteString(lipgloss.NewStyle().MarginTop(1).MarginLeft(2).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Padding(0, 5, 0, 5).Render(fmt.Sprintf("Add key")))

	fmt.Fprintf(&b, "\n\n")

	for i := range m.textInputs {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(m.textInputs[i].View()))
		b.WriteRune('\n')
	}

	// Folder picker row.
	folderName := ""
	if len(m.folders) > 0 {
		folderName = m.folders[m.folderIdx].Name
	}
	folderRow := fmt.Sprintf("Folder: < %s >", folderName)
	if m.focusIndex == formFolderIdx {
		folderRow = focusedStyle.Render(folderRow) + cursorModeHelpStyle.Render("  (←/→ to cycle)")
	} else {
		folderRow = blurredStyle.Render(folderRow)
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(folderRow))

	button := &blurredButton
	if m.focusIndex == formButtonIdx {
		button = &focusedButton
	}

	fmt.Fprintf(&b, "\n\n%s\n\n", *button)

	if m.error != "" {
		fmt.Fprintf(&b, "\n%s\n", errorText.Render(m.error))
	}

	return b.String()
}
