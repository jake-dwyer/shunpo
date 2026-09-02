package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jake-dwyer/clackity/internal/words"
)

type state int

const (
	stateReady state = iota
	stateTyping
	stateDone
)

type testMode int

const (
	modeTime testMode = iota
	modeWords
)

type completedWord struct {
	typed  string
	target string
}

var timePresets = []int{10, 15, 30, 60, 120}
var wordPresets = []int{10, 25, 50, 100}

type tickMsg time.Time

// Config is the starting mode/duration, typically from CLI flags. It need
// not match a preset exactly; presets only drive the in-app settings
// palette (see menuFocused) and default to their closest entry.
type Config struct {
	Mode     testMode
	Seconds  int
	WordGoal int
}

func TimeConfig(seconds int) Config { return Config{Mode: modeTime, Seconds: seconds} }
func WordsConfig(words int) Config  { return Config{Mode: modeWords, WordGoal: words} }

type Model struct {
	cfg Config

	// timeIdx/wordsIdx track which preset is highlighted in the settings
	// palette; they're cosmetic and only meaningfully authoritative once
	// the user has picked a preset there.
	timeIdx  int
	wordsIdx int

	// menuFocused gates the settings palette. While true, every keystroke
	// is a palette command (preset digit, mode letter, close). While
	// false, every keystroke is typing input. Keeping these mutually
	// exclusive is what lets mode-switching live entirely on the
	// keyboard without ever hijacking a real word's first letter.
	menuFocused bool

	state state

	targetWords []string
	completed   []completedWord
	current     string
	wordIdx     int

	keyCorrect   int
	keyIncorrect int

	start   time.Time
	elapsed time.Duration
	result  Result
	width   int
	height  int
}

func New(cfg Config) Model {
	m := Model{
		cfg:      cfg,
		timeIdx:  nearestIndex(timePresets, cfg.Seconds, 2),
		wordsIdx: nearestIndex(wordPresets, cfg.WordGoal, 1),
	}
	m.reset()
	return m
}

func nearestIndex(presets []int, val, def int) int {
	for i, p := range presets {
		if p == val {
			return i
		}
	}
	return def
}

func (m *Model) reset() {
	n := 200
	if m.cfg.Mode == modeWords {
		n = m.cfg.WordGoal
	}
	m.targetWords = words.Generate(n)
	m.completed = nil
	m.current = ""
	m.wordIdx = 0
	m.keyCorrect = 0
	m.keyIncorrect = 0
	m.state = stateReady
	m.elapsed = 0
}

func (m Model) Init() tea.Cmd {
	return nil
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/10, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m *Model) timeLimit() time.Duration {
	return time.Duration(m.cfg.Seconds) * time.Second
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil

	case tickMsg:
		if m.state != stateTyping {
			return m, nil
		}
		m.elapsed = time.Since(m.start)
		if m.cfg.Mode == modeTime && m.elapsed >= m.timeLimit() {
			m.finish()
			return m, nil
		}
		return m, tickCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return m, tea.Quit

	case tea.KeyEsc:
		if m.menuFocused {
			m.menuFocused = false
			return m, nil
		}
		return m, tea.Quit

	case tea.KeyEnter:
		if m.menuFocused {
			m.menuFocused = false
			return m, nil
		}
		m.reset()
		return m, nil

	case tea.KeyTab:
		if m.state == stateDone {
			m.reset()
			return m, nil
		}
		m.menuFocused = !m.menuFocused
		return m, nil
	}

	if m.menuFocused {
		return m.handleMenuKey(msg)
	}

	if m.state == stateDone {
		return m, nil
	}
	return m.handleTypingKey(msg)
}

func (m Model) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type != tea.KeyRunes || len(msg.Runes) != 1 {
		return m, nil
	}
	switch r := msg.Runes[0]; r {
	case 't':
		m.cfg.Mode = modeTime
		m.cfg.Seconds = timePresets[m.timeIdx]
		m.reset()
	case 'w':
		m.cfg.Mode = modeWords
		m.cfg.WordGoal = wordPresets[m.wordsIdx]
		m.reset()
	case '1', '2', '3', '4', '5':
		idx := int(r - '1')
		if m.cfg.Mode == modeTime && idx < len(timePresets) {
			m.timeIdx = idx
			m.cfg.Seconds = timePresets[idx]
			m.reset()
		} else if m.cfg.Mode == modeWords && idx < len(wordPresets) {
			m.wordsIdx = idx
			m.cfg.WordGoal = wordPresets[idx]
			m.reset()
		}
	}
	return m, nil
}

func (m Model) handleTypingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	target := m.currentTarget()

	switch msg.Type {
	case tea.KeyBackspace:
		if m.current == "" {
			if m.wordIdx > 0 && len(m.completed) > 0 {
				last := m.completed[len(m.completed)-1]
				m.completed = m.completed[:len(m.completed)-1]
				m.current = last.typed
				m.wordIdx--
			}
			return m, nil
		}
		m.current = m.current[:len(m.current)-1]
		return m, nil

	case tea.KeySpace:
		if m.current == "" {
			return m, nil
		}
		m.completed = append(m.completed, completedWord{typed: m.current, target: target})
		m.current = ""
		m.wordIdx++
		if m.cfg.Mode == modeWords && m.wordIdx >= len(m.targetWords) {
			m.finish()
			return m, nil
		}
		return m, nil

	case tea.KeyRunes:
		var cmd tea.Cmd
		if m.state == stateReady {
			m.state = stateTyping
			m.start = time.Now()
			cmd = tickCmd()
		}
		for _, r := range msg.Runes {
			pos := len(m.current)
			if pos < len(target) {
				if byte(r) == target[pos] {
					m.keyCorrect++
				} else {
					m.keyIncorrect++
				}
			} else {
				m.keyIncorrect++
			}
			m.current += string(r)
		}
		return m, cmd
	}
	return m, nil
}

func (m *Model) currentTarget() string {
	if m.wordIdx < len(m.targetWords) {
		return m.targetWords[m.wordIdx]
	}
	return ""
}

func (m *Model) finish() {
	m.state = stateDone
	dur := time.Since(m.start)
	if m.cfg.Mode == modeTime {
		dur = m.timeLimit()
	}
	m.elapsed = dur
	m.result = computeResult(m.completed, m.current, m.currentTarget(), m.keyCorrect, m.keyIncorrect, dur)
}

func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	switch m.state {
	case stateDone:
		return m.viewResults()
	default:
		return m.viewTyping()
	}
}

func (m Model) viewHeader() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("clackity") + "  ")

	if m.cfg.Mode == modeTime {
		b.WriteString(styleModeActive.Render(fmt.Sprintf("time %d", m.cfg.Seconds)))
	} else {
		b.WriteString(styleModeActive.Render(fmt.Sprintf("words %d", m.cfg.WordGoal)))
	}
	if !m.menuFocused {
		b.WriteString(styleHelp.Render("   tab: settings"))
	}
	return b.String()
}

func (m Model) viewMenu() string {
	renderColumn := func(label string, presets []int, idx int, active bool) string {
		var b strings.Builder
		labelStyle := styleModeInactive
		if active {
			labelStyle = styleModeActive
		}
		b.WriteString(labelStyle.Render(label) + "  ")
		for i, p := range presets {
			s := fmt.Sprintf("%d", p)
			if active && i == idx {
				b.WriteString(styleAccent.Render("[" + s + "]"))
			} else {
				b.WriteString(styleMuted.Render(s))
			}
			b.WriteString(" ")
		}
		return b.String()
	}

	var b strings.Builder
	b.WriteString(renderColumn("time", timePresets, m.timeIdx, m.cfg.Mode == modeTime))
	b.WriteString("\n")
	b.WriteString(renderColumn("words", wordPresets, m.wordsIdx, m.cfg.Mode == modeWords))
	b.WriteString("\n")
	b.WriteString(styleHelp.Render("t/w mode  ·  1-5 preset  ·  enter/esc/tab close"))
	return b.String()
}

func (m Model) viewTyping() string {
	var b strings.Builder
	b.WriteString(m.viewHeader())
	b.WriteString("\n\n")

	if m.menuFocused {
		b.WriteString(m.viewMenu())
		b.WriteString("\n\n")
	}

	if m.cfg.Mode == modeTime {
		remaining := m.timeLimit() - m.elapsed
		if remaining < 0 {
			remaining = 0
		}
		b.WriteString(styleAccent.Render(fmt.Sprintf("%d", int(remaining.Seconds()+0.999))))
	} else {
		b.WriteString(styleAccent.Render(fmt.Sprintf("%d/%d", m.wordIdx, len(m.targetWords))))
	}
	b.WriteString("\n\n")

	b.WriteString(m.viewWords())
	b.WriteString("\n")

	b.WriteString(styleHelp.Render("esc quit  ·  enter restart  ·  tab settings"))
	return b.String()
}

const wrapWidth = 70

func (m Model) viewWords() string {
	var lines []strings.Builder
	lines = append(lines, strings.Builder{})
	lineLen := 0

	push := func(s string, visLen int) {
		if lineLen+visLen > wrapWidth {
			lines = append(lines, strings.Builder{})
			lineLen = 0
		}
		lines[len(lines)-1].WriteString(s)
		lineLen += visLen
	}

	renderWord := func(idx int) {
		target := m.targetWords[idx]
		var typed string
		isCurrent := idx == m.wordIdx
		if isCurrent {
			typed = m.current
		} else if idx < len(m.completed) {
			typed = m.completed[idx].typed
		}

		var w strings.Builder
		for i, tc := range target {
			ch := string(tc)
			if i < len(typed) {
				if typed[i] == byte(tc) {
					w.WriteString(styleFg.Render(ch))
				} else {
					w.WriteString(styleError.Bold(true).Render(ch))
				}
			} else if isCurrent && i == len(typed) {
				w.WriteString(styleCursor.Render(ch))
			} else {
				w.WriteString(styleMuted.Render(ch))
			}
		}
		if len(typed) > len(target) {
			w.WriteString(styleError.Underline(true).Render(typed[len(target):]))
		}
		if isCurrent && len(typed) >= len(target) {
			w.WriteString(styleCursor.Render(" "))
		}

		push(w.String()+" ", len([]rune(target))+1)
	}

	limit := len(m.targetWords)
	if m.cfg.Mode == modeTime && limit > 60 {
		limit = 60
	}
	for i := 0; i < limit; i++ {
		renderWord(i)
	}

	var out []string
	for _, l := range lines {
		out = append(out, l.String())
	}
	return strings.Join(out, "\n")
}

func (m Model) viewResults() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("clackity") + "\n\n")

	r := m.result
	statStyle := lipgloss.NewStyle().Width(14)
	stat := func(label string, value string) string {
		return statStyle.Render(styleStatLabel.Render(label) + "\n" + styleBigStat.Render(value))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		stat("wpm", fmt.Sprintf("%.0f", r.WPM)),
		stat("acc", fmt.Sprintf("%.0f%%", r.Accuracy)),
		stat("raw", fmt.Sprintf("%.0f", r.RawWPM)),
		stat("time", fmt.Sprintf("%.0fs", r.Duration.Seconds())),
	)
	b.WriteString(row)
	b.WriteString("\n\n")
	b.WriteString(styleMuted.Render(fmt.Sprintf("correct %d  ·  incorrect %d", r.Correct, r.Incorrect)))
	b.WriteString("\n\n")
	b.WriteString(styleHelp.Render("tab restart  ·  esc quit"))
	return b.String()
}
