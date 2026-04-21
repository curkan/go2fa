package screens

import (
	"fmt"
	"go2fa/internal/storage"
	"go2fa/internal/structure"
	"go2fa/internal/twofactor"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

// Custom key bindings surfaced through the list's built-in help bar.
var (
	keysKeyCopy    = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "copy"))
	keysKeyPick    = key.NewBinding(key.WithKeys("0", "1", "2", "3", "4", "5", "6", "7", "8", "9"), key.WithHelp("0-9", "pick"))
	keysKeyAdd     = key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add"))
	keysKeyEdit    = key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit"))
	keysKeyDelete  = key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "delete"))
	keysKeyMove    = key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move"))
	keysKeyReorder = key.NewBinding(key.WithKeys("shift+up", "shift+down", "K", "J"), key.WithHelp("shift+↑/↓ · J/K", "reorder"))
	keysKeyBack    = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))
)

var keysShortHelp = []key.Binding{keysKeyCopy, keysKeyPick, keysKeyAdd, keysKeyEdit, keysKeyDelete, keysKeyMove, keysKeyReorder, keysKeyBack}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

var globalCopied = false

type itemKey struct {
	title    string
	desc     string
	secret   string
	folderID string
}

type ItemDelegate struct{}
type tickMsg struct{}


func (d ItemDelegate) Height() int                             { return 1 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var (
		title, desc, code, secret string
		exp int64
		s = list.NewDefaultItemStyles()
	)

	s.SelectedTitle = s.SelectedTitle.Foreground(lipgloss.Color("#99dd99")).BorderLeftForeground(lipgloss.Color("#7aa37a"))
	s.SelectedDesc = s.SelectedDesc.Foreground(lipgloss.Color("#7aa37a")).BorderLeftForeground(lipgloss.Color("#7aa37a")).MarginBottom(1)
	s.NormalDesc = s.NormalDesc.MarginBottom(1)

	if i, ok := listItem.(itemKey); ok {
		title = i.title
		desc = i.desc
		secret = i.secret
		code, exp = twofactor.GenerateTOTP(secret)
	} else {
		return
	}

	// Slot label self-documents the 0-9 quick-pick binding.
	// Slots 1-9 hold the first nine items, slot 0 wraps around to the 10th.
	switch {
	case index < 9:
		title = fmt.Sprintf("[%d] %s", index+1, title)
	case index == 9:
		title = "[0] " + title
	}

	// Conditions
	var (
		isSelected  = index == m.Index()
	)

	if !isSelected {
		code = "******"
	}

	if globalCopied {
		s.SelectedTitle = s.SelectedTitle.BorderLeftBackground(lipgloss.Color("#7aa37a"))
		s.SelectedDesc = s.SelectedDesc.BorderLeftBackground(lipgloss.Color("#7aa37a"))
	}

	width, _, _ := term.GetSize(0)
	width = 50
	padding := 0

	until := exp - time.Now().Unix()

	codeStyle := lipgloss.NewStyle().Bold(true).Blink(true)

	if until <= 15 && until > 5 {
		codeStyle = codeStyle.Foreground(lipgloss.Color("#FFF0A1"))
	}

	if until <= 5 {
		codeStyle = codeStyle.Foreground(lipgloss.Color("#FF7575"))
	}

	if isSelected && m.FilterState() != list.Filtering {
		padding = width - len(title) - len(code) - 5
		title = s.SelectedTitle.Render(title)
		desc = s.SelectedDesc.Render(desc)
	} else {
		padding = width - len(title) - len(code) - 5
		title = s.NormalTitle.Render(title)
		desc = s.NormalDesc.Render(desc)
	}


	code = codeStyle.Render(code)

	fmt.Fprintf(w, "%s%*s%s %ds\n%s", title, padding, " ", code, until, desc)
}


func (i itemKey) Title() string       { return i.title }
func (i itemKey) Description() string { return i.desc }
func (i itemKey) FilterValue() string { return i.title }

type listKeysModel struct {
	list          list.Model
	itemsKeysList []structure.TwoFactorItem
	folderID      string
	folderName    string
}

func (m listKeysModel) Init() tea.Cmd {
	return tick()
}

func (m listKeysModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.list.FilterState() == list.Filtering {
	}
	switch msg := msg.(type) {

	case tea.KeyMsg:
		output := termenv.NewOutput(os.Stdout)

		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
				case "q", "ctrl+c":
					output.ClearScreen()

					return m, tea.Quit
				case "d":
					item, ok := m.list.SelectedItem().(itemKey)

					if !ok {
						return m, tick()
					}

					twoFactorItem := structure.TwoFactorItem{
						Title:    item.title,
						Desc:     item.desc,
						Secret:   item.secret,
						FolderID: item.folderID,
					}

					screen := ScreenDeleteKey(twoFactorItem, m.folderID, m.folderName)
					return RootScreen().SwitchScreen(&screen)
				case "m":
					item, ok := m.list.SelectedItem().(itemKey)
					if !ok {
						return m, tick()
					}
					twoFactorItem := structure.TwoFactorItem{
						Title:    item.title,
						Desc:     item.desc,
						Secret:   item.secret,
						FolderID: item.folderID,
					}
					screen := ScreenMoveKey(twoFactorItem, m.folderID)
					return RootScreen().SwitchScreen(&screen)
				case "e":
					item, ok := m.list.SelectedItem().(itemKey)
					if !ok {
						return m, tick()
					}
					twoFactorItem := structure.TwoFactorItem{
						Title:    item.title,
						Desc:     item.desc,
						Secret:   item.secret,
						FolderID: item.folderID,
					}
					screen := ScreenEditKey(twoFactorItem, m.folderID, m.folderName)
					return RootScreen().SwitchScreen(&screen)
				case "a":
					// Preselect the current folder (if scoped) so adding a key
					// to a folder you're already viewing is a single keystroke.
					screen := ScreenInputSecret(m.folderID, m.folderID, m.folderName, false)
					return RootScreen().SwitchScreen(&screen)
				case "shift+up", "K":
					return m.reorderSelected(-1)
				case "shift+down", "J":
					return m.reorderSelected(1)
				case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
					n, _ := strconv.Atoi(msg.String())
					// "0" wraps around to the 10th visible item; "1".."9" map 1:1.
					visible := n - 1
					if n == 0 {
						visible = 9
					}
					items := m.list.VisibleItems()
					if visible < 0 || visible >= len(items) {
						return m, tick()
					}
					item, ok := items[visible].(itemKey)
					if !ok {
						return m, tick()
					}
					m.list.Select(visible)
					code, _ := twofactor.GenerateTOTP(item.secret)
					clipboard.WriteAll(code)
					globalCopied = true
					return m, tick()
				}
		}

		switch msg.Type {
			case tea.KeyCtrlC:
				output.ClearScreen()
				return m, tea.Quit

			case tea.KeyEsc:
				screen := ListFoldersScreen()
				return RootScreen().SwitchScreen(&screen)

			case tea.KeyEnter:
				if m.list.FilterState() != list.Filtering {
					item, ok := m.list.SelectedItem().(itemKey)

					if !ok {
						return m, tick()
					}

					code, _ := twofactor.GenerateTOTP(item.secret)
					clipboard.WriteAll(code)
					globalCopied = true
				}
			default:
				globalCopied = false
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tickMsg:
		return m, tick()
	}


	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listKeysModel) View() string {
	return docStyle.Render(m.list.View())
}

// reorderSelected swaps the currently selected item with its neighbour
// (direction -1 = up, +1 = down) within the current folder scope, persists
// the new order, and keeps the cursor on the moved item.
func (m listKeysModel) reorderSelected(direction int) (tea.Model, tea.Cmd) {
	item, ok := m.list.SelectedItem().(itemKey)
	if !ok {
		return m, nil
	}

	store, err := storage.LoadStore()
	if err != nil {
		return m, nil
	}

	matcher := func(it structure.TwoFactorItem) bool {
		return it.Title == item.title && it.Desc == item.desc &&
			it.Secret == item.secret && it.FolderID == item.folderID
	}
	if !storage.ReorderItem(&store, matcher, direction, m.folderID) {
		return m, nil
	}
	if err := storage.SaveStore(store); err != nil {
		return m, nil
	}

	scoped := storage.ItemsInFolder(store, m.folderID)
	listItems := make([]list.Item, 0, len(scoped))
	for _, it := range scoped {
		listItems = append(listItems, itemKey{
			title:    it.Title,
			desc:     it.Desc,
			secret:   it.Secret,
			folderID: it.FolderID,
		})
	}
	m.itemsKeysList = scoped
	newIdx := m.list.Index() + direction
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(listItems) {
		newIdx = len(listItems) - 1
	}
	m.list.SetItems(listItems)
	m.list.Select(newIdx)
	return m, nil
}

// ListKeysScreen opens the unscoped (all keys) view. Kept for compatibility.
func ListKeysScreen() listKeysModel {
	return ListKeysScreenScoped("", "")
}

// ListKeysScreenScoped opens the list scoped to folderID. An empty folderID
// means "show all items" (the synthetic All keys scope).
func ListKeysScreenScoped(folderID, folderName string) listKeysModel {
	store, err := storage.LoadStore()
	if err != nil {
		logrus.Fatal(err)
	}

	itemsKeysList := storage.ItemsInFolder(store, folderID)

	itemKeys := make([]list.Item, 0, len(itemsKeysList))
	for _, item := range itemsKeysList {
		itemKeys = append(itemKeys, itemKey{
			title:    item.Title,
			desc:     item.Desc,
			secret:   item.Secret,
			folderID: item.FolderID,
		})
	}

	l := list.New(itemKeys, ItemDelegate{}, 30, 20)
	l.AdditionalShortHelpKeys = func() []key.Binding { return keysShortHelp }
	l.AdditionalFullHelpKeys = func() []key.Binding { return keysShortHelp }

	m := listKeysModel{
		list:          l,
		itemsKeysList: itemsKeysList,
		folderID:      folderID,
		folderName:    folderName,
	}

	title := "Active Keys"
	if folderName != "" {
		title = "Active Keys — " + folderName
	} else if folderID == "" {
		title = "Active Keys — All"
	}
	m.list.Title = title

	return m
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
