package screens

import (
	"fmt"
	"go2fa/internal/storage"
	"go2fa/internal/structure"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// The edit form has 3 focusable positions: Title (0), Desc (1), Save (2).
const (
	editButtonIdx = 2
)

var (
	focusedSaveButton = focusedStyle.Padding(0, 2).Bold(true).Render("[Save]")
	blurredSaveButton = blurredStyle.Padding(0, 2).Render("[Save]")
)

type editKeyScreen struct {
	focusIndex   int
	textInputs   []textinput.Model
	cursorMode   cursor.Mode
	target       structure.TwoFactorItem
	fromFolderID string
	fromFolder   string
	err          error
	error        string
}

// ScreenEditKey builds an edit screen for a single item. Only Title and Desc
// are editable. The item is located by (title, desc, secret, folder_id) just
// like in delete/move flows.
func ScreenEditKey(target structure.TwoFactorItem, fromFolderID, fromFolder string) editKeyScreen {
	m := editKeyScreen{
		textInputs:   make([]textinput.Model, 2),
		target:       target,
		fromFolderID: fromFolderID,
		fromFolder:   fromFolder,
	}

	var t textinput.Model
	for i := range m.textInputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle

		switch i {
		case 0:
			t.Placeholder = "Title"
			t.CharLimit = 32
			t.SetValue(target.Title)
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "Description"
			t.CharLimit = 128
			t.SetValue(target.Desc)
		}

		m.textInputs[i] = t
	}

	return m
}

func (m editKeyScreen) Init() tea.Cmd { return textinput.Blink }

func (m editKeyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	output := termenv.NewOutput(os.Stdout)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			screen := ListKeysScreenScoped(m.fromFolderID, m.fromFolder)
			return RootScreen().SwitchScreen(&screen)
		case tea.KeyCtrlC:
			output.ClearScreen()
			return m, tea.Quit
		}

		switch msg.String() {
		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			if s == "enter" && m.focusIndex == editButtonIdx {
				title := strings.TrimSpace(m.textInputs[0].Value())
				desc := m.textInputs[1].Value()
				if title == "" {
					m.error = "Title must not be empty"
					return m, nil
				}
				if err := applyEdit(m.target, title, desc); err != nil {
					m.error = err.Error()
					return m, nil
				}
				screen := ListKeysScreenScoped(m.fromFolderID, m.fromFolder)
				return RootScreen().SwitchScreen(&screen)
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > editButtonIdx {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = editButtonIdx
			}

			cmds := make([]tea.Cmd, len(m.textInputs))
			for i := range m.textInputs {
				if i == m.focusIndex {
					cmds[i] = m.textInputs[i].Focus()
					m.textInputs[i].PromptStyle = focusedStyle
					m.textInputs[i].TextStyle = focusedStyle
					continue
				}
				m.textInputs[i].Blur()
				m.textInputs[i].PromptStyle = noStyle
				m.textInputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *editKeyScreen) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.textInputs))
	for i := range m.textInputs {
		m.textInputs[i], cmds[i] = m.textInputs[i].Update(msg)
	}
	return tea.Batch(cmds...)
}

func (m editKeyScreen) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().MarginTop(1).MarginLeft(2).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Padding(0, 5, 0, 5).Render("Edit key"))

	fmt.Fprintf(&b, "\n\n")

	for i := range m.textInputs {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(m.textInputs[i].View()))
		b.WriteRune('\n')
	}

	button := &blurredSaveButton
	if m.focusIndex == editButtonIdx {
		button = &focusedSaveButton
	}

	fmt.Fprintf(&b, "\n%s\n\n", *button)

	if m.error != "" {
		fmt.Fprintf(&b, "\n%s\n", errorText.Render(m.error))
	}

	return b.String()
}

// applyEdit locates the target item by identity and updates title/desc.
func applyEdit(target structure.TwoFactorItem, title, desc string) error {
	store, err := storage.LoadStore()
	if err != nil {
		return err
	}
	match := func(it structure.TwoFactorItem) bool {
		return it.Title == target.Title && it.Desc == target.Desc &&
			it.Secret == target.Secret && it.FolderID == target.FolderID
	}
	if !storage.UpdateItemMeta(&store, match, title, desc) {
		return fmt.Errorf("item not found")
	}
	return storage.SaveStore(store)
}
