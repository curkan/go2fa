package main

import (
	"fmt"
	screens "go2fa/internal/screens"
	"go2fa/internal/vault"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

func main() {
	screen_y := screens.ListMethodsScreen()

	output := termenv.NewOutput(os.Stdout)
	output.ClearScreen()

	homeDir := os.Getenv("HOME")
	filePath := filepath.Join(homeDir, ".local", "share", "go2fa", "stores", "vault.json")
	_, err := os.Open(filePath)

	if err != nil {
		vault.Create()
	}

	if _, err := tea.NewProgram(screen_y).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

