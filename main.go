package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jake-dwyer/clackity/internal/app"
)

func main() {
	timeFlag := flag.Int("time", 30, "test duration in seconds (ignored if -words is set)")
	wordsFlag := flag.Int("words", 0, "number of words per test (overrides -time)")
	flag.Parse()

	var cfg app.Config
	if *wordsFlag > 0 {
		cfg = app.WordsConfig(*wordsFlag)
	} else {
		cfg = app.TimeConfig(*timeFlag)
	}

	p := tea.NewProgram(app.New(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "clackity:", err)
		os.Exit(1)
	}
}
