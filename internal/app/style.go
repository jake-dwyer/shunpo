package app

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name   string
	Bg     string
	Panel  string // slightly-lighter-than-bg surface, used for kbd badges/borders
	Accent string
	Fg     string
	Muted  string
	Error  string
}

var themes = []Theme{
	{Name: "serika", Bg: "#232323", Panel: "#2c2c2c", Accent: "#e2b714", Fg: "#d1d0c5", Muted: "#646669", Error: "#ca4754"},
	{Name: "dracula", Bg: "#282a36", Panel: "#343746", Accent: "#bd93f9", Fg: "#f8f8f2", Muted: "#6272a4", Error: "#ff5555"},
	{Name: "nord", Bg: "#2e3440", Panel: "#3b4252", Accent: "#88c0d0", Fg: "#e5e9f0", Muted: "#4c566a", Error: "#bf616a"},
	{Name: "gruvbox", Bg: "#282828", Panel: "#3c3836", Accent: "#fabd2f", Fg: "#ebdbb2", Muted: "#928374", Error: "#fb4934"},
	{Name: "monokai", Bg: "#272822", Panel: "#3e3d32", Accent: "#a6e22e", Fg: "#f8f8f2", Muted: "#75715e", Error: "#f92672"},
	{Name: "tokyonight", Bg: "#1a1b26", Panel: "#24283b", Accent: "#7aa2f7", Fg: "#c0caf5", Muted: "#565f89", Error: "#f7768e"},
	{Name: "catppuccin", Bg: "#1e1e2e", Panel: "#313244", Accent: "#f5c2e7", Fg: "#cdd6f4", Muted: "#6c7086", Error: "#f38ba8"},
	{Name: "solarized", Bg: "#002b36", Panel: "#073642", Accent: "#b58900", Fg: "#839496", Muted: "#586e75", Error: "#dc322f"},

	// Anime-inspired themes.
	{Name: "vagabond", Bg: "#0d0d0c", Panel: "#1a1a18", Accent: "#f2f0e6", Fg: "#c9c7ba", Muted: "#4a4944", Error: "#8b2419"},
	{Name: "chidori", Bg: "#090c14", Panel: "#131a2b", Accent: "#4fd1ff", Fg: "#eaf6ff", Muted: "#35415c", Error: "#ff7a3d"},
	{Name: "gojo", Bg: "#0a0f14", Panel: "#131c22", Accent: "#34e2e2", Fg: "#eafffb", Muted: "#3d5654", Error: "#ff5c8a"},
	{Name: "susanoo", Bg: "#140d08", Panel: "#241608", Accent: "#ff9d2e", Fg: "#f7e6cf", Muted: "#6b5439", Error: "#d1453b"},
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
	bg     lipgloss.Color
	panel  lipgloss.Color
	accent lipgloss.Style
	muted  lipgloss.Style
	fg     lipgloss.Style
	err    lipgloss.Style
	caret  lipgloss.Style // thin insertion-point bar, not a block cursor

	title lipgloss.Style
	help  lipgloss.Style
	kbd   lipgloss.Style // small pill badge for a key name, e.g. "esc"

	modeOn  lipgloss.Style
	modeOff lipgloss.Style

	bigStat   lipgloss.Style
	statLabel lipgloss.Style

	border lipgloss.Style // rounded-border card wrapper (settings palette)
}

func buildStyles(t Theme) styleSet {
	bg := lipgloss.Color(t.Bg)
	panel := lipgloss.Color(t.Panel)
	accent := lipgloss.Color(t.Accent)
	fg := lipgloss.Color(t.Fg)
	muted := lipgloss.Color(t.Muted)
	errC := lipgloss.Color(t.Error)

	return styleSet{
		bg:     bg,
		panel:  panel,
		accent: lipgloss.NewStyle().Foreground(accent).Bold(true),
		muted:  lipgloss.NewStyle().Foreground(muted),
		fg:     lipgloss.NewStyle().Foreground(fg),
		err:    lipgloss.NewStyle().Foreground(errC),
		caret:  lipgloss.NewStyle().Foreground(accent),

		title: lipgloss.NewStyle().Foreground(accent).Bold(true),
		help:  lipgloss.NewStyle().Foreground(muted),
		kbd:   lipgloss.NewStyle().Foreground(fg).Background(panel).Padding(0, 1),

		modeOn:  lipgloss.NewStyle().Foreground(accent).Bold(true),
		modeOff: lipgloss.NewStyle().Foreground(muted),

		bigStat:   lipgloss.NewStyle().Foreground(accent).Bold(true),
		statLabel: lipgloss.NewStyle().Foreground(muted),

		border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(muted).Padding(0, 2),
	}
}

// kbd renders a small "keycap" badge, e.g. kbd(s, "esc") -> a padded pill.
func kbd(s styleSet, label string) string {
	return s.kbd.Render(label)
}

// hintLine joins alternating kbd-badge/plain-text fragments with the
// muted style, e.g. hintLine(s, "esc", " quit  ", "tab", " settings").
func hintLine(s styleSet, parts ...string) string {
	var out string
	for i, p := range parts {
		if i%2 == 0 {
			out += kbd(s, p)
		} else {
			out += s.help.Render(p)
		}
	}
	return out
}
