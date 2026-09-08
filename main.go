package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jake-dwyer/shunpo/internal/app"
	"github.com/jake-dwyer/shunpo/internal/store"
)

func main() {
	st := store.Load()

	timeFlag := flag.Int("time", 30, "test duration in seconds (ignored if -words is set)")
	wordsFlag := flag.Int("words", 0, "number of words per test (overrides -time)")
	themeFlag := flag.String("theme", "serika", "color theme (serika, dracula, nord, gruvbox, monokai, tokyonight, catppuccin, solarized)")
	flag.Parse()

	explicit := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })

	// Saved settings from the last run become the defaults; an explicit
	// flag on this invocation always wins.
	cfg := app.Config{}
	switch {
	case explicit["words"]:
		cfg = app.WordsConfig(*wordsFlag)
	case explicit["time"]:
		cfg = app.TimeConfig(*timeFlag)
	case st.Settings.Mode == "words" && st.Settings.WordGoal > 0:
		cfg = app.WordsConfig(st.Settings.WordGoal)
	case st.Settings.Seconds > 0:
		cfg = app.TimeConfig(st.Settings.Seconds)
	default:
		cfg = app.TimeConfig(*timeFlag)
	}

	switch {
	case explicit["theme"]:
		cfg.Theme = *themeFlag
	case st.Settings.Theme != "":
		cfg.Theme = st.Settings.Theme
	default:
		cfg.Theme = *themeFlag
	}

	// Whatever this process ends up running with becomes the new
	// remembered default, even if the user quits before finishing a
	// test or ever opening the settings palette.
	st.Settings = cfg.ToSettings()
	_ = st.Save()

	p := tea.NewProgram(app.NewWithStore(cfg, st), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "shunpo:", err)
		os.Exit(1)
	}
}
