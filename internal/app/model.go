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
	Theme    string
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
	themeIdx int

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
		themeIdx: themeIndexByName(cfg.Theme),
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
	// Time mode has no natural end to the word list, so it needs enough
	// words that even a very fast typist won't run out before the timer
	// does. ~6 chars/word average and a generous WPM ceiling of 300 over
	// the longest test (120s) is ~600 words; 800 leaves headroom.
	n := 800
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
	case 'c':
		m.themeIdx = (m.themeIdx + 1) % len(themes)
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

// submitWord commits the current input as a completed word and advances
// to the next one.
func (m *Model) submitWord() {
	m.completed = append(m.completed, completedWord{typed: m.current, target: m.currentTarget()})
	m.current = ""
	m.wordIdx++
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
		m.submitWord()
		if m.cfg.Mode == modeWords && m.wordIdx >= len(m.targetWords) {
			m.finish()
		}
		return m, nil

	case tea.KeyRunes:
		var cmd tea.Cmd
		if m.state == stateReady {
			m.state = stateTyping
			m.start = time.Now()
			cmd = tickCmd()
		}
		isLastWord := m.cfg.Mode == modeWords && m.wordIdx == len(m.targetWords)-1
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

			// Word mode ends the instant the final word reaches full
			// length - no trailing space required to submit it.
			if isLastWord && len(target) > 0 && len(m.current) == len(target) {
				m.submitWord()
				m.finish()
				return m, nil
			}
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

func (m Model) theme() Theme {
	return themes[m.themeIdx]
}

func (m Model) styles() styleSet {
	return buildStyles(m.theme())
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

func (m Model) viewHeader(s styleSet) string {
	var b strings.Builder
	b.WriteString(s.title.Render("clackity") + "  ")

	if m.cfg.Mode == modeTime {
		b.WriteString(s.modeOn.Render(fmt.Sprintf("time %d", m.cfg.Seconds)))
	} else {
		b.WriteString(s.modeOn.Render(fmt.Sprintf("words %d", m.cfg.WordGoal)))
	}
	if !m.menuFocused {
		b.WriteString(s.help.Render("   tab: settings"))
	}
	return b.String()
}

func (m Model) viewMenu(s styleSet) string {
	renderRow := func(label string, presets []int, idx int, active bool) string {
		var b strings.Builder
		labelStyle := s.modeOff
		if active {
			labelStyle = s.modeOn
		}
		b.WriteString(labelStyle.Render(label) + "  ")
		for i, p := range presets {
			ps := fmt.Sprintf("%d", p)
			if active && i == idx {
				b.WriteString(s.accent.Render("[" + ps + "]"))
			} else {
				b.WriteString(s.muted.Render(ps))
			}
			b.WriteString(" ")
		}
		return b.String()
	}

	renderThemeRow := func() string {
		var b strings.Builder
		b.WriteString(s.modeOn.Render("theme") + "  ")
		for i, t := range themes {
			if i == m.themeIdx {
				b.WriteString(s.accent.Render("[" + t.Name + "]"))
			} else {
				b.WriteString(s.muted.Render(t.Name))
			}
			b.WriteString(" ")
		}
		return b.String()
	}

	var b strings.Builder
	b.WriteString(renderRow("time ", timePresets, m.timeIdx, m.cfg.Mode == modeTime))
	b.WriteString("\n")
	b.WriteString(renderRow("words", wordPresets, m.wordsIdx, m.cfg.Mode == modeWords))
	b.WriteString("\n")
	b.WriteString(renderThemeRow())
	b.WriteString("\n")
	b.WriteString(s.help.Render("t/w mode  ·  1-5 preset  ·  c theme  ·  enter/esc/tab close"))
	return b.String()
}

func (m Model) viewTyping() string {
	s := m.styles()
	var b strings.Builder
	b.WriteString(m.viewHeader(s))
	b.WriteString("\n\n")

	if m.menuFocused {
		b.WriteString(m.viewMenu(s))
		b.WriteString("\n\n")
	}

	if m.cfg.Mode == modeTime {
		remaining := m.timeLimit() - m.elapsed
		if remaining < 0 {
			remaining = 0
		}
		b.WriteString(s.accent.Render(fmt.Sprintf("%d", int(remaining.Seconds()+0.999))))
	} else {
		b.WriteString(s.accent.Render(fmt.Sprintf("%d/%d", m.wordIdx, len(m.targetWords))))
	}
	b.WriteString("\n\n")

	b.WriteString(m.viewWords(s))
	b.WriteString("\n")

	b.WriteString(s.help.Render("esc quit  ·  enter restart  ·  tab settings"))
	return b.String()
}

const wrapWidth = 70

func (m Model) viewWords(s styleSet) string {
	var lines []strings.Builder
	lines = append(lines, strings.Builder{})
	lineLen := 0

	push := func(content string, sep string, visLen int) {
		if lineLen+visLen > wrapWidth {
			lines = append(lines, strings.Builder{})
			lineLen = 0
		}
		lines[len(lines)-1].WriteString(content)
		lines[len(lines)-1].WriteString(sep)
		lineLen += visLen
	}

	// renderWord returns the styled word content, its visible width
	// (excluding the trailing separator), and whether the cursor
	// currently sits right after it (i.e. it's the current word and
	// fully typed) - in which case the single separator space that
	// follows should be rendered as the cursor block rather than a
	// second, plain space stacked on top of it.
	renderWord := func(idx int) (string, int, bool) {
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
			switch {
			case i < len(typed):
				if typed[i] == byte(tc) {
					w.WriteString(s.fg.Render(ch))
				} else {
					w.WriteString(s.err.Bold(true).Render(ch))
				}
			case isCurrent && i == len(typed):
				w.WriteString(s.cursor.Render(ch))
			default:
				w.WriteString(s.muted.Render(ch))
			}
		}
		visLen := len([]rune(target))
		if len(typed) > len(target) {
			extra := typed[len(target):]
			w.WriteString(s.err.Underline(true).Render(extra))
			visLen += len(extra)
		}

		cursorAtEnd := isCurrent && len(typed) >= len(target)
		return w.String(), visLen, cursorAtEnd
	}

	limit := len(m.targetWords)
	if m.cfg.Mode == modeTime && limit > 60 {
		limit = 60
	}
	for i := 0; i < limit; i++ {
		content, visLen, cursorAtEnd := renderWord(i)
		sep := " "
		if cursorAtEnd {
			sep = s.cursor.Render(" ")
		}
		push(content, sep, visLen+1)
	}

	var out []string
	for _, l := range lines {
		out = append(out, l.String())
	}
	return strings.Join(out, "\n")
}

func (m Model) viewResults() string {
	s := m.styles()
	var b strings.Builder
	b.WriteString(s.title.Render("clackity") + "\n\n")

	r := m.result
	statStyle := lipgloss.NewStyle().Width(14)
	stat := func(label string, value string) string {
		return statStyle.Render(s.statLabel.Render(label) + "\n" + s.bigStat.Render(value))
	}

	row := lipgloss.JoinHorizontal(lipgloss.Top,
		stat("wpm", fmt.Sprintf("%.0f", r.WPM)),
		stat("acc", fmt.Sprintf("%.0f%%", r.Accuracy)),
		stat("raw", fmt.Sprintf("%.0f", r.RawWPM)),
		stat("time", fmt.Sprintf("%.0fs", r.Duration.Seconds())),
	)
	b.WriteString(row)
	b.WriteString("\n\n")
	b.WriteString(s.muted.Render(fmt.Sprintf("correct %d  ·  incorrect %d", r.Correct, r.Incorrect)))
	b.WriteString("\n\n")
	b.WriteString(s.help.Render("tab restart  ·  esc quit"))
	return b.String()
}
