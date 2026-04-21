package screens

import (
	"fmt"
	"go2fa/internal/storage"
	"io"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
)

// Custom bindings surfaced through the list's built-in help bar.
var (
	folderKeyOpen      = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open"))
	folderKeyPick      = key.NewBinding(key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"), key.WithHelp("0-9", "pick"))
	folderKeyAddKey    = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add key"))
	folderKeyNewFolder = key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new folder"))
	folderKeyRename    = key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "rename"))
	folderKeyDelete    = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete"))
	folderKeyBack      = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "quit"))
)

// folderItem represents one row on the folders screen. An empty ID denotes
// the synthetic "All keys" entry.
type folderItem struct {
	id    string
	name  string
	count int
}

func (f folderItem) FilterValue() string { return f.name }

type folderItemDelegate struct{}

func (d folderItemDelegate) Height() int                             { return 1 }
func (d folderItemDelegate) Spacing() int                            { return 0 }
func (d folderItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d folderItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	f, ok := listItem.(folderItem)
	if !ok {
		return
	}

	// Slot label self-documents the 0-9 quick-pick binding.
	slot := ""
	if index <= 9 {
		slot = fmt.Sprintf("[%d] ", index)
	}
	line := fmt.Sprintf("%s%s (%d)", slot, f.name, f.count)
	if index == m.Index() {
		fmt.Fprint(w, selectedItemStyle.Render("→ "+line))
		return
	}
	fmt.Fprint(w, itemStyle.Render(line))
}

var folderHelp = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.Color("#D2D2D2"))

var folderShortHelp = []key.Binding{folderKeyOpen, folderKeyPick, folderKeyAddKey, folderKeyNewFolder, folderKeyRename, folderKeyDelete, folderKeyBack}

type listFoldersModel struct {
	list list.Model
}

func (m listFoldersModel) Init() tea.Cmd { return nil }

func (m listFoldersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		switch msg.String() {
		case "q", "ctrl+c":
			output.ClearScreen()
			return m, tea.Quit
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			n, _ := strconv.Atoi(msg.String())
			items := m.list.Items()
			if n >= len(items) {
				return m, nil
			}
			f, ok := items[n].(folderItem)
			if !ok {
				return m, nil
			}
			screen := ListKeysScreenScoped(f.id, f.name)
			return RootScreen().SwitchScreen(&screen)
		case "a":
			f, _ := m.list.SelectedItem().(folderItem)
			// For the synthetic "All keys" row f.id == "" → Default will be preselected.
			screen := ScreenInputSecret(f.id, "", "", true)
			return RootScreen().SwitchScreen(&screen)
		case "n":
			screen := ScreenCreateFolder()
			return RootScreen().SwitchScreen(&screen)
		case "e":
			f, ok := m.list.SelectedItem().(folderItem)
			if !ok || f.id == "" {
				return m, nil
			}
			screen := ScreenRenameFolder(f.id, f.name)
			return RootScreen().SwitchScreen(&screen)
		case "d":
			f, ok := m.list.SelectedItem().(folderItem)
			if !ok || f.id == "" || f.id == storage.DefaultFolderID {
				return m, nil
			}
			screen := ScreenDeleteFolder(f.id, f.name, f.count)
			return RootScreen().SwitchScreen(&screen)
		}

		switch msg.Type {
		case tea.KeyEsc:
			output.ClearScreen()
			return m, tea.Quit

		case tea.KeyEnter:
			f, ok := m.list.SelectedItem().(folderItem)
			if !ok {
				return m, nil
			}
			screen := ListKeysScreenScoped(f.id, f.name)
			return RootScreen().SwitchScreen(&screen)
		}

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listFoldersModel) View() string {
	return "\n" + m.list.View()
}

// ListFoldersScreen builds the folders list (with synthetic All keys on top).
func ListFoldersScreen() listFoldersModel {
	store, err := storage.LoadStore()
	if err != nil {
		logrus.Fatal(err)
	}

	counts := storage.CountByFolder(store)
	totalItems := len(store.Items)

	items := []list.Item{
		folderItem{id: "", name: "All keys", count: totalItems},
	}
	for _, f := range store.Folders {
		items = append(items, folderItem{id: f.ID, name: f.Name, count: counts[f.ID]})
	}

	l := list.New(items, folderItemDelegate{}, 40, listHeight)
	l.Title = "go2fa"
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = func() []key.Binding { return folderShortHelp }
	l.AdditionalFullHelpKeys = func() []key.Binding { return folderShortHelp }

	return listFoldersModel{list: l}
}
