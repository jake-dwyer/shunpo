package app

import "github.com/charmbracelet/lipgloss"

var (
	colorBg       = lipgloss.Color("#232323")
	colorAccent   = lipgloss.Color("#e2b714")
	colorFg       = lipgloss.Color("#d1d0c5")
	colorMuted    = lipgloss.Color("#646669")
	colorError    = lipgloss.Color("#ca4754")
	colorErrorDim = lipgloss.Color("#7e2a30")

	styleAccent = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleMuted  = lipgloss.NewStyle().Foreground(colorMuted)
	styleFg     = lipgloss.NewStyle().Foreground(colorFg)
	styleError  = lipgloss.NewStyle().Foreground(colorError)
	styleCursor = lipgloss.NewStyle().Foreground(colorBg).Background(colorAccent)

	styleTitle = lipgloss.NewStyle().Foreground(colorAccent).Bold(true).MarginBottom(1)
	styleHelp  = lipgloss.NewStyle().Foreground(colorMuted).MarginTop(1)

	styleModeActive   = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleModeInactive = lipgloss.NewStyle().Foreground(colorMuted)

	styleBigStat   = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleStatLabel = lipgloss.NewStyle().Foreground(colorMuted)
)
