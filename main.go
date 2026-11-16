package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jake-dwyer/shunpo/internal/app"
	"github.com/jake-dwyer/shunpo/internal/store"
)

const defaultSeconds = 30

func main() {
	st := store.Load()

	timeFlag := flag.Int("time", defaultSeconds, "test duration in seconds (ignored if -words is set)")
	wordsFlag := flag.Int("words", 0, "number of words per test, must be positive (overrides -time)")
	themeFlag := flag.String("theme", "serika", "color theme (serika, dracula, nord, gruvbox, monokai, tokyonight, catppuccin, solarized, manga, chidori, gojo, susanoo)")
	flag.Parse()

	explicit := map[string]bool{}
	flag.Visit(func(f *flag.Flag) { explicit[f.Name] = true })

	// A non-positive -words/-time doesn't describe a test that can ever
	// finish (no last word to reach full length, or no time limit to
	// hit), so it's treated as if the flag were never passed rather than
	// handed to the model, which would just hang until ctrl+c.
	if explicit["words"] && *wordsFlag <= 0 {
		fmt.Fprintln(os.Stderr, "shunpo: -words must be positive, ignoring")
		explicit["words"] = false
	}
	if explicit["time"] && *timeFlag <= 0 {
		fmt.Fprintln(os.Stderr, "shunpo: -time must be positive, ignoring")
		explicit["time"] = false
	}

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
		cfg = app.TimeConfig(defaultSeconds)
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
