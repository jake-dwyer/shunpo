package app

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name   string
	Bg     string
	Accent string
	Fg     string
	Muted  string
	Error  string
}

var themes = []Theme{
	{Name: "serika", Bg: "#232323", Accent: "#e2b714", Fg: "#d1d0c5", Muted: "#646669", Error: "#ca4754"},
	{Name: "dracula", Bg: "#282a36", Accent: "#bd93f9", Fg: "#f8f8f2", Muted: "#6272a4", Error: "#ff5555"},
	{Name: "nord", Bg: "#2e3440", Accent: "#88c0d0", Fg: "#e5e9f0", Muted: "#4c566a", Error: "#bf616a"},
	{Name: "gruvbox", Bg: "#282828", Accent: "#fabd2f", Fg: "#ebdbb2", Muted: "#928374", Error: "#fb4934"},
	{Name: "monokai", Bg: "#272822", Accent: "#a6e22e", Fg: "#f8f8f2", Muted: "#75715e", Error: "#f92672"},
	{Name: "tokyonight", Bg: "#1a1b26", Accent: "#7aa2f7", Fg: "#c0caf5", Muted: "#565f89", Error: "#f7768e"},
	{Name: "catppuccin", Bg: "#1e1e2e", Accent: "#f5c2e7", Fg: "#cdd6f4", Muted: "#6c7086", Error: "#f38ba8"},
	{Name: "solarized", Bg: "#002b36", Accent: "#b58900", Fg: "#839496", Muted: "#586e75", Error: "#dc322f"},
}

func themeIndexByName(name string) int {
	for i, t := range themes {
		if t.Name == name {
			return i
		}
	}
	return 0
}

type styleSet struct {
	accent    lipgloss.Style
	muted     lipgloss.Style
	fg        lipgloss.Style
	err       lipgloss.Style
	cursor    lipgloss.Style
	title     lipgloss.Style
	help      lipgloss.Style
	modeOn    lipgloss.Style
	modeOff   lipgloss.Style
	bigStat   lipgloss.Style
	statLabel lipgloss.Style
}

func buildStyles(t Theme) styleSet {
	bg := lipgloss.Color(t.Bg)
	accent := lipgloss.Color(t.Accent)
	fg := lipgloss.Color(t.Fg)
	muted := lipgloss.Color(t.Muted)
	errC := lipgloss.Color(t.Error)

	return styleSet{
		accent:    lipgloss.NewStyle().Foreground(accent).Bold(true),
		muted:     lipgloss.NewStyle().Foreground(muted),
		fg:        lipgloss.NewStyle().Foreground(fg),
		err:       lipgloss.NewStyle().Foreground(errC),
		cursor:    lipgloss.NewStyle().Foreground(bg).Background(accent),
		title:     lipgloss.NewStyle().Foreground(accent).Bold(true).MarginBottom(1),
		help:      lipgloss.NewStyle().Foreground(muted).MarginTop(1),
		modeOn:    lipgloss.NewStyle().Foreground(accent).Bold(true),
		modeOff:   lipgloss.NewStyle().Foreground(muted),
		bigStat:   lipgloss.NewStyle().Foreground(accent).Bold(true),
		statLabel: lipgloss.NewStyle().Foreground(muted),
	}
}
