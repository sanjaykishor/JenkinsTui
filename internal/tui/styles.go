package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors used throughout the application
var (
	ColorPrimary   = lipgloss.Color("#0D8AC9") // Jenkins blue
	ColorSecondary = lipgloss.Color("#CC0000") // Error/failure red
	ColorSuccess   = lipgloss.Color("#339900") // Success green
	ColorWarning   = lipgloss.Color("#F0AD4E") // Warning yellow
	ColorGray      = lipgloss.Color("#666666") // Neutral gray
	ColorDarkGray  = lipgloss.Color("#333333") // Dark gray for backgrounds
	ColorLightGray = lipgloss.Color("#CCCCCC") // Light gray for borders
	ColorWhite     = lipgloss.Color("#FFFFFF") // White
	ColorBlack     = lipgloss.Color("#000000") // Black
)

// Common Styles
var (
	// Base text styles
	NormalText = lipgloss.NewStyle().
			Foreground(ColorWhite)

	BoldText = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Bold(true)

	HeaderText = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)

	// Status styles
	SuccessText = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	FailureText = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	WarningText = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)

	// Container styles
	AppContainer = lipgloss.NewStyle().
			Padding(1, 2)

	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 2).
		Margin(0, 1)

	// Tab styles
	ActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorBlack).
			Background(ColorPrimary).
			Padding(0, 3)

	InactiveTab = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorDarkGray).
			Padding(0, 3)

	// Status bar styles
	StatusBar = lipgloss.NewStyle().
			Foreground(ColorWhite).
			Background(ColorDarkGray).
			Padding(0, 1)

	// Help style
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorLightGray).
			MarginLeft(1)
)

