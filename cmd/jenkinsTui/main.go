package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanjaykishor/JenkinsTui.git/internal/tui"
)

func main() {
	// Create a new instance of our application
	app, err := tui.New()
	if err != nil {
		fmt.Println("Error creating application:", err)
		os.Exit(1)
	}
	// Create a new Bubble Tea program with our model
	program := tea.NewProgram(
		app,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Start the program
	if _, err := program.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
