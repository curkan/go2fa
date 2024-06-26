package main

import (
	"fmt"
	screens "go2fa/internal/screens"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

func main() {
	screen_y := screens.ListMethodsScreen()

	output := termenv.NewOutput(os.Stdout)
	output.ClearScreen()

	if _, err := tea.NewProgram(screen_y).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

