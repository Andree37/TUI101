package main

import (
	"fmt"
	"os"

	"tui101/app"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Create the main application model
	model := app.NewModel()

	// Create the tea program with alt screen for full screen TUI
	program := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the program
	if _, err := program.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
