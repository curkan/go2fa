package screens

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// Shared list-screen styles and layout constants.
const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("#FFFFFF")).Padding(0, 5, 0, 5)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#54B575"))
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#A1FCC0"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
)
