package main

import (
	"fmt"
	"os"
	tea "github.com/charmbracelet/bubbletea"
	screens "go2fa/internal/screens"
)

func main() {
	screen_y := screens.ListMethodsScreen()

	if _, err := tea.NewProgram(screen_y).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

