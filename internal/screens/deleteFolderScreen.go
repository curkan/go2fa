package screens

import (
	"fmt"
	"go2fa/internal/storage"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

type deleteFolderScreen struct {
	folderID   string
	folderName string
	itemCount  int
	err        string
}

// ScreenDeleteFolder opens confirmation for deleting folderID. itemCount is
// purely for display; actual reassignment is done by storage.DeleteFolder.
func ScreenDeleteFolder(folderID, folderName string, itemCount int) deleteFolderScreen {
	return deleteFolderScreen{
		folderID:   folderID,
		folderName: folderName,
		itemCount:  itemCount,
	}
}

func (m deleteFolderScreen) Init() tea.Cmd { return nil }

func (m deleteFolderScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		switch msg.String() {
		case "q", "ctrl+c":
			output.ClearScreen()
			return m, tea.Quit
		case "esc":
			screen := ListFoldersScreen()
			return RootScreen().SwitchScreen(&screen)
		case "enter":
			store, err := storage.LoadStore()
			if err != nil {
				m.err = err.Error()
				return m, nil
			}
			if err := storage.DeleteFolder(&store, m.folderID, ""); err != nil {
				m.err = err.Error()
				return m, nil
			}
			if err := storage.SaveStore(store); err != nil {
				m.err = err.Error()
				return m, nil
			}
			screen := ListFoldersScreen()
			return RootScreen().SwitchScreen(&screen)
		}
	}
	return m, nil
}

func (m deleteFolderScreen) View() string {
	var b strings.Builder

	var msg string
	if m.itemCount == 0 {
		msg = fmt.Sprintf("Delete empty folder %q?", m.folderName)
	} else {
		msg = fmt.Sprintf("Delete folder %q? %d key(s) will be moved to %s.",
			m.folderName, m.itemCount, storage.DefaultFolderName)
	}
	b.WriteString(warning.Render(msg))

	fmt.Fprintf(&b, "\n\n%s\n", help.Render("To confirm - [Enter]\nTo cancel  - [Esc]"))
	if m.err != "" {
		fmt.Fprintf(&b, "\n%s\n", errorText.Render(m.err))
	}
	return b.String()
}
