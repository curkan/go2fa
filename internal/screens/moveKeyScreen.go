package screens

import (
	"fmt"
	"go2fa/internal/storage"
	"go2fa/internal/structure"
	"io"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

var (
	moveKeyMove   = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "move"))
	moveKeyCancel = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))
)

var moveShortHelp = []key.Binding{moveKeyMove, moveKeyCancel}

type moveTargetItem struct {
	id   string
	name string
}

func (t moveTargetItem) FilterValue() string { return t.name }

type moveTargetDelegate struct{}

func (d moveTargetDelegate) Height() int                             { return 1 }
func (d moveTargetDelegate) Spacing() int                            { return 0 }
func (d moveTargetDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d moveTargetDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	t, ok := listItem.(moveTargetItem)
	if !ok {
		return
	}
	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("→ "+t.name))
		return
	}
	fmt.Fprint(w, itemStyle.Render(t.name))
}

type moveKeyScreen struct {
	list       list.Model
	target     structure.TwoFactorItem
	fromFolder string
	err        string
}

// ScreenMoveKey opens a folder picker. After the user confirms, the selected
// folder id is assigned to the first item that matches `target` by identity
// (title+desc+secret+folder_id).
func ScreenMoveKey(target structure.TwoFactorItem, fromFolder string) moveKeyScreen {
	store, err := storage.LoadStore()
	items := []list.Item{}
	if err == nil {
		for _, f := range store.Folders {
			if f.ID == target.FolderID {
				continue
			}
			items = append(items, moveTargetItem{id: f.ID, name: f.Name})
		}
	}

	l := list.New(items, moveTargetDelegate{}, 40, listHeight)
	l.Title = "Move to folder"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = func() []key.Binding { return moveShortHelp }
	l.AdditionalFullHelpKeys = func() []key.Binding { return moveShortHelp }

	m := moveKeyScreen{list: l, target: target, fromFolder: fromFolder}
	if err != nil {
		m.err = err.Error()
	}
	return m
}

func (m moveKeyScreen) Init() tea.Cmd { return nil }

func (m moveKeyScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		switch msg.String() {
		case "q", "ctrl+c":
			output.ClearScreen()
			return m, tea.Quit
		}

		switch msg.Type {
		case tea.KeyEsc:
			screen := ListKeysScreenScoped(m.fromFolder, folderNameFor(m.fromFolder))
			return RootScreen().SwitchScreen(&screen)
		case tea.KeyEnter:
			t, ok := m.list.SelectedItem().(moveTargetItem)
			if !ok {
				return m, nil
			}
			if err := applyMove(m.target, t.id); err != nil {
				m.err = err.Error()
				return m, nil
			}
			screen := ListKeysScreenScoped(m.fromFolder, folderNameFor(m.fromFolder))
			return RootScreen().SwitchScreen(&screen)
		}
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m moveKeyScreen) View() string {
	out := "\n" + m.list.View()
	if m.err != "" {
		out += "\n" + errorText.Render(m.err)
	}
	return out
}

// applyMove locates the target item in the store and reassigns it.
func applyMove(target structure.TwoFactorItem, newFolderID string) error {
	store, err := storage.LoadStore()
	if err != nil {
		return err
	}
	idx := -1
	for i, it := range store.Items {
		if it.Title == target.Title && it.Desc == target.Desc &&
			it.Secret == target.Secret && it.FolderID == target.FolderID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("item not found")
	}
	if err := storage.MoveItem(&store, idx, newFolderID); err != nil {
		return err
	}
	return storage.SaveStore(store)
}

// folderNameFor looks up a folder's display name; used to repopulate the
// scoped keys list after a move.
func folderNameFor(folderID string) string {
	if folderID == "" {
		return ""
	}
	store, err := storage.LoadStore()
	if err != nil {
		return ""
	}
	if f, ok := storage.FindFolderByID(store, folderID); ok {
		return f.Name
	}
	return ""
}
