package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/sanjaykishor/JenkinsTui.git/internal/tui"
)

func main() {
	// Create a new model
	m, err := tui.New()
	if err != nil {
		fmt.Printf("Error initializing application: %v\n", err)
		os.Exit(1)
	}

	// Initialize program with model
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Start the application
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}
