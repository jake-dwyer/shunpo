package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/jake-dwyer/shunpo/internal/store"
	"github.com/jake-dwyer/shunpo/internal/words"
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

// Config is the starting mode/duration, typically from CLI flags or saved
// settings. It need not match a preset exactly; presets only drive the
// in-app settings palette (see menuFocused) and default to their closest
// entry.
type Config struct {
	Mode     testMode
	Seconds  int
	WordGoal int
	Theme    string
}

func TimeConfig(seconds int) Config { return Config{Mode: modeTime, Seconds: seconds} }
func WordsConfig(words int) Config  { return Config{Mode: modeWords, WordGoal: words} }

// ToSettings converts a Config into the persisted shape, so whatever a
// process launches with - via flags or remembered state - can be written
// back as the new remembered default even if the user quits without
// finishing a test or touching the settings palette.
func (c Config) ToSettings() store.Settings {
	mode := "time"
	if c.Mode == modeWords {
		mode = "words"
	}
	return store.Settings{Mode: mode, Seconds: c.Seconds, WordGoal: c.WordGoal, Theme: c.Theme}
}

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

	// themeEditorOpen gates the theme editor overlay, opened with 'e'
	// from the settings palette. editTheme is a working copy (starting
	// from the active, already-overridden theme) that edits apply to
	// live; it's only written back to the store on close. editField
	// indexes into {bg, fg, muted, accent, error}. editingHex is true
	// while actively typing a replacement hex value into hexBuf.
	themeEditorOpen bool
	editTheme       Theme
	editField       int
	editingHex      bool
	hexBuf          string

	state state

	targetWords []string
	completed   []completedWord
	current     string
	wordIdx     int

	keyCorrect   int
	keyIncorrect int
	typoTally    map[string]int // per-test; merged into store.Typos on finish

	wpmHistory []float64 // ~1 sample/sec while typing, for the results sparkline
	lastSample time.Time

	start   time.Time
	elapsed time.Duration
	result  Result
	width   int
	height  int

	// store is shunpo's cross-session state (settings/records/typos).
	// persist gates whether changes get written to disk: true for the
	// real CLI (NewWithStore), false for New(), which tests use and
	// which must never touch the user's actual state file.
	store   store.Store
	persist bool
}

func New(cfg Config) Model {
	m := newModel(cfg, store.Store{Records: map[string]store.Record{}, Typos: map[string]int{}})
	m.persist = false
	return m
}

func NewWithStore(cfg Config, st store.Store) Model {
	m := newModel(cfg, st)
	m.persist = true
	return m
}

func newModel(cfg Config, st store.Store) Model {
	m := Model{
		cfg:      cfg,
		timeIdx:  nearestIndex(timePresets, cfg.Seconds, 2),
		wordsIdx: nearestIndex(wordPresets, cfg.WordGoal, 1),
		themeIdx: themeIndexByName(cfg.Theme),
		store:    st,
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
	m.typoTally = map[string]int{}
	m.wpmHistory = nil
	m.lastSample = time.Time{}
	m.state = stateReady
	m.elapsed = 0
	m.result = Result{}
}

func (m *Model) syncSettings() {
	mode := "time"
	if m.cfg.Mode == modeWords {
		mode = "words"
	}
	m.store.Settings = store.Settings{
		Mode:     mode,
		Seconds:  m.cfg.Seconds,
		WordGoal: m.cfg.WordGoal,
		Theme:    themes[m.themeIdx].Name,
	}
}

func (m *Model) maybeSave() {
	if m.persist {
		_ = m.store.Save()
	}
}

func recordKey(cfg Config) string {
	if cfg.Mode == modeTime {
		return store.RecordKey("time", cfg.Seconds)
	}
	return store.RecordKey("words", cfg.WordGoal)
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

// liveWPM is the net WPM so far, using the same formula as the final
// result - used to sample the in-progress sparkline.
func (m Model) liveWPM() float64 {
	r := computeResult(m.completed, m.current, m.currentTarget(), m.keyCorrect, m.keyIncorrect, m.elapsed)
	return r.WPM
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
		if time.Since(m.lastSample) >= time.Second {
			m.wpmHistory = append(m.wpmHistory, m.liveWPM())
			m.lastSample = time.Now()
		}
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
	if msg.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	if m.themeEditorOpen {
		return m.handleEditorKey(msg)
	}

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
		m.syncSettings()
		m.maybeSave()
	case 'w':
		m.cfg.Mode = modeWords
		m.cfg.WordGoal = wordPresets[m.wordsIdx]
		m.reset()
		m.syncSettings()
		m.maybeSave()
	case 'c':
		m.themeIdx = (m.themeIdx + 1) % len(themes)
		m.syncSettings()
		m.maybeSave()
	case '1', '2', '3', '4', '5':
		idx := int(r - '1')
		if m.cfg.Mode == modeTime && idx < len(timePresets) {
			m.timeIdx = idx
			m.cfg.Seconds = timePresets[idx]
			m.reset()
			m.syncSettings()
			m.maybeSave()
		} else if m.cfg.Mode == modeWords && idx < len(wordPresets) {
			m.wordsIdx = idx
			m.cfg.WordGoal = wordPresets[idx]
			m.reset()
			m.syncSettings()
			m.maybeSave()
		}
	case 'e':
		m.editTheme = m.theme()
		m.editField = 0
		m.editingHex = false
		m.hexBuf = ""
		m.themeEditorOpen = true
	}
	return m, nil
}

// setEditField writes hex into whichever of editTheme's five editable
// colors editField currently points at.
func (m *Model) setEditField(hex string) {
	switch m.editField {
	case 0:
		m.editTheme.Bg = hex
	case 1:
		m.editTheme.Fg = hex
	case 2:
		m.editTheme.Muted = hex
	case 3:
		m.editTheme.Accent = hex
	default:
		m.editTheme.Error = hex
	}
}

func isHexDigit(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}

func (m Model) handleEditorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editingHex {
		switch msg.Type {
		case tea.KeyEnter:
			if len(m.hexBuf) == 6 {
				m.setEditField("#" + strings.ToLower(m.hexBuf))
			}
			m.editingHex = false
			return m, nil
		case tea.KeyEsc:
			m.editingHex = false
			return m, nil
		case tea.KeyBackspace:
			if len(m.hexBuf) > 0 {
				m.hexBuf = m.hexBuf[:len(m.hexBuf)-1]
			}
			return m, nil
		case tea.KeyRunes:
			for _, r := range msg.Runes {
				if len(m.hexBuf) < 6 && isHexDigit(r) {
					m.hexBuf += string(r)
				}
			}
			if len(m.hexBuf) == 6 {
				m.setEditField("#" + strings.ToLower(m.hexBuf))
			}
			return m, nil
		}
		return m, nil
	}

	switch msg.Type {
	case tea.KeyEsc:
		// Commit the working copy as this theme's saved override and
		// close, back to the settings palette.
		if m.store.ThemeOverrides == nil {
			m.store.ThemeOverrides = map[string]store.ThemeColors{}
		}
		m.store.ThemeOverrides[m.editTheme.Name] = store.ThemeColors{
			Bg: m.editTheme.Bg, Fg: m.editTheme.Fg, Muted: m.editTheme.Muted,
			Accent: m.editTheme.Accent, Error: m.editTheme.Error,
		}
		m.maybeSave()
		m.themeEditorOpen = false
		return m, nil
	case tea.KeyUp:
		m.editField = (m.editField + 4) % 5
		return m, nil
	case tea.KeyDown:
		m.editField = (m.editField + 1) % 5
		return m, nil
	case tea.KeyEnter:
		// Start empty rather than pre-filled: hexBuf only supports
		// append/backspace-from-end (no cursor), so typing a fresh value
		// is the only editing model that makes sense here.
		m.editingHex = true
		m.hexBuf = ""
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'r' {
			delete(m.store.ThemeOverrides, m.editTheme.Name)
			m.maybeSave()
			m.editTheme = themes[m.themeIdx]
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
			m.lastSample = m.start
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
					m.typoTally[string(target[pos])]++
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

	for ch, n := range m.typoTally {
		m.store.Typos[ch] += n
	}

	key := recordKey(m.cfg)
	prev, existed := m.store.Records[key]
	m.result.PrevBest = prev.WPM
	if !existed || m.result.WPM > prev.WPM {
		m.result.IsPR = true
		m.store.Records[key] = store.Record{
			WPM:      m.result.WPM,
			Accuracy: m.result.Accuracy,
			Date:     time.Now().Format("2006-01-02"),
		}
	}

	m.maybeSave()
}

// theme returns the active theme with any saved user override layered
// on top of the built-in preset.
func (m Model) theme() Theme {
	t := themes[m.themeIdx]
	if ov, ok := m.store.ThemeOverrides[t.Name]; ok {
		t.Bg, t.Fg, t.Muted, t.Accent, t.Error = ov.Bg, ov.Fg, ov.Muted, ov.Accent, ov.Error
	}
	return t
}

func (m Model) styles() styleSet {
	return buildStyles(m.theme())
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	s := m.styles()
	var content string
	switch m.state {
	case stateDone:
		content = m.viewResults(s)
	default:
		content = m.viewTyping(s)
	}
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content,
		lipgloss.WithWhitespaceBackground(s.bg))
}

func (m Model) viewHeader(s styleSet) string {
	var b strings.Builder
	b.WriteString(s.title.Render("瞬歩 shunpo") + s.muted.Render("  "))
	b.WriteString(s.accent.Render("☆") + s.muted.Render("  "))

	if m.cfg.Mode == modeTime {
		b.WriteString(s.modeOn.Render(fmt.Sprintf("time %d", m.cfg.Seconds)))
	} else {
		b.WriteString(s.modeOn.Render(fmt.Sprintf("words %d", m.cfg.WordGoal)))
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
		b.WriteString(labelStyle.Render(label) + s.muted.Render("  "))
		for i, p := range presets {
			ps := fmt.Sprintf("%d", p)
			if active && i == idx {
				b.WriteString(s.accent.Render("[" + ps + "]"))
			} else {
				b.WriteString(s.muted.Render(ps))
			}
			b.WriteString(s.muted.Render(" "))
		}
		return b.String()
	}

	// The theme row can outgrow a narrow terminal, so it wraps onto
	// further lines indented to line up under the label - the same
	// pattern viewWords uses for wrapping the word list itself.
	// indentWidth tracks the plain visible width for wrap math;
	// indentStyled is the actual (background-styled) string written.
	renderThemeRow := func(maxWidth int) string {
		label := "theme"
		indentWidth := len(label) + 2
		indentStyled := s.muted.Render(strings.Repeat(" ", indentWidth))
		lines := []strings.Builder{{}}
		lines[0].WriteString(s.modeOn.Render(label) + s.muted.Render("  "))
		curLen := indentWidth

		for i, t := range themes {
			entry := t.Name
			if i == m.themeIdx {
				entry = "[" + t.Name + "]"
			}
			entryLen := len(entry) + 1
			if curLen+entryLen > maxWidth && curLen > indentWidth {
				lines = append(lines, strings.Builder{})
				lines[len(lines)-1].WriteString(indentStyled)
				curLen = indentWidth
			}
			if i == m.themeIdx {
				lines[len(lines)-1].WriteString(s.accent.Render(entry))
			} else {
				lines[len(lines)-1].WriteString(s.muted.Render(entry))
			}
			lines[len(lines)-1].WriteString(s.muted.Render(" "))
			curLen += entryLen
		}

		out := make([]string, len(lines))
		for i, l := range lines {
			out[i] = l.String()
		}
		return strings.Join(out, "\n")
	}

	menuWidth := m.wrapWidth() - 6 // border chars + padding

	var b strings.Builder
	b.WriteString(renderRow("time ", timePresets, m.timeIdx, m.cfg.Mode == modeTime))
	b.WriteString("\n")
	b.WriteString(renderRow("words", wordPresets, m.wordsIdx, m.cfg.Mode == modeWords))
	b.WriteString("\n")
	b.WriteString(renderThemeRow(menuWidth))
	b.WriteString("\n\n")
	b.WriteString(hintLine(s, "t", "/", "w", " mode   ", "1-5", " preset   ", "c", " theme   ", "e", " edit   ", "enter", " close"))

	return s.border.Render(b.String())
}

var editFieldLabels = []string{"bg", "fg", "muted", "accent", "error"}

// viewThemeEditor renders live-editable swatches for the theme currently
// being tuned. Chrome (labels, borders, hints) uses the stable, already
// -committed theme so the editor itself never becomes unreadable while
// mid-edit of a bad value; only the sample line previews the in-progress
// edit (m.editTheme).
func (m Model) viewThemeEditor(outer styleSet) string {
	live := buildStyles(m.editTheme)
	values := []string{m.editTheme.Bg, m.editTheme.Fg, m.editTheme.Muted, m.editTheme.Accent, m.editTheme.Error}

	var b strings.Builder
	b.WriteString(outer.title.Render("theme editor") + outer.muted.Render("  "+m.editTheme.Name))
	b.WriteString("\n\n")

	for i, label := range editFieldLabels {
		cursor := outer.muted.Render("  ")
		if i == m.editField {
			cursor = outer.accent.Render("> ")
		}
		swatch := lipgloss.NewStyle().Background(lipgloss.Color(values[i])).Render("  ")
		row := cursor + outer.modeOff.Render(fmt.Sprintf("%-7s", label)) + swatch + outer.muted.Render(" ")
		if i == m.editField && m.editingHex {
			row += outer.fg.Render("#"+m.hexBuf) + outer.caret.Render("_")
		} else {
			row += outer.fg.Render(values[i])
		}
		b.WriteString(row)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	sample := live.fg.Render("correct") + live.muted.Render(" ") +
		live.err.Bold(true).Render("wrong") + live.muted.Render(" ") +
		live.muted.Render("upcoming") + live.muted.Render("   ") +
		live.accent.Render("42")
	b.WriteString(lipgloss.NewStyle().Background(lipgloss.Color(m.editTheme.Bg)).Padding(0, 1).Render(sample))
	b.WriteString("\n\n")

	b.WriteString(hintLine(outer, "up/down", " field   ", "enter", " edit hex   ", "r", " reset theme   ", "esc", " save & close"))
	return outer.border.Render(b.String())
}

func (m Model) viewTyping(s styleSet) string {
	var b strings.Builder
	b.WriteString(m.viewHeader(s))
	b.WriteString("\n\n")

	if m.themeEditorOpen {
		b.WriteString(m.viewThemeEditor(s))
		b.WriteString("\n\n")
	} else if m.menuFocused {
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
	b.WriteString("\n\n")

	b.WriteString(hintLine(s, "esc", " quit   ", "enter", " restart   ", "tab", " settings"))
	return b.String()
}

// caretGlyph is a thin insertion-point bar, matching monkeytype's caret
// style, rather than a solid block covering a character cell. Plain ASCII
// on purpose: box-drawing bars like "│" fall in Unicode's East-Asian
// "ambiguous width" category, so some terminal/font combinations render
// them double-width while our own width math assumes single-width - that
// mismatch is what caused the earlier stray background banding.
const caretGlyph = "|"

func (m Model) wrapWidth() int {
	w := m.width - 12
	if w > 100 {
		w = 100
	}
	if w < 30 {
		w = 30
	}
	return w
}

func (m Model) viewWords(s styleSet) string {
	wrapWidth := m.wrapWidth()
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
	// (excluding the trailing separator), and whether the caret
	// currently sits right after it (i.e. it's the current word and
	// fully typed) - in which case the caret is drawn as part of the
	// separator instead of a second element stacked on top of it.
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
		visLen := 0
		for i, tc := range target {
			ch := string(tc)
			if i < len(typed) {
				if typed[i] == byte(tc) {
					w.WriteString(s.fg.Render(ch))
				} else {
					w.WriteString(s.err.Bold(true).Render(ch))
				}
				visLen++
				continue
			}
			if isCurrent && i == len(typed) {
				w.WriteString(s.caret.Render(caretGlyph))
				visLen++
			}
			w.WriteString(s.muted.Render(ch))
			visLen++
		}
		if len(typed) > len(target) {
			extra := typed[len(target):]
			w.WriteString(s.err.Underline(true).Render(extra))
			visLen += len(extra)
		}

		caretAtEnd := isCurrent && len(typed) >= len(target)
		return w.String(), visLen, caretAtEnd
	}

	limit := len(m.targetWords)
	if m.cfg.Mode == modeTime && limit > 60 {
		limit = 60
	}
	for i := 0; i < limit; i++ {
		content, visLen, caretAtEnd := renderWord(i)
		sep := s.muted.Render(" ")
		sepLen := 1
		if caretAtEnd {
			sep = s.caret.Render(caretGlyph) + s.muted.Render(" ")
			sepLen = 2
		}
		push(content, sep, visLen+sepLen)
	}

	var out []string
	for _, l := range lines {
		out = append(out, l.String())
	}
	return strings.Join(out, "\n")
}

var sparkChars = []rune("▁▂▃▄▅▆▇█")

func sparkline(vals []float64, width int) string {
	if len(vals) == 0 || width <= 0 {
		return ""
	}
	minV, maxV := vals[0], vals[0]
	for _, v := range vals {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	span := maxV - minV
	if span == 0 {
		span = 1
	}
	n := len(vals)
	out := make([]rune, width)
	for i := 0; i < width; i++ {
		idx := i * n / width
		if idx >= n {
			idx = n - 1
		}
		level := int((vals[idx] - minV) / span * float64(len(sparkChars)-1))
		if level < 0 {
			level = 0
		}
		if level > len(sparkChars)-1 {
			level = len(sparkChars) - 1
		}
		out[i] = sparkChars[level]
	}
	return string(out)
}

type weakKey struct {
	char  string
	count int
}

func (m Model) topWeakKeys(n int) []weakKey {
	out := make([]weakKey, 0, len(m.store.Typos))
	for ch, c := range m.store.Typos {
		if c > 0 {
			out = append(out, weakKey{char: ch, count: c})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].count != out[j].count {
			return out[i].count > out[j].count
		}
		return out[i].char < out[j].char
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}

func (m Model) viewResults(s styleSet) string {
	r := m.result
	var b strings.Builder
	b.WriteString(s.title.Render("瞬歩 shunpo") + "\n\n")

	if r.IsPR {
		if r.PrevBest > 0 {
			b.WriteString(s.accent.Render(fmt.Sprintf("new best!  (was %.0f wpm)", r.PrevBest)))
		} else {
			b.WriteString(s.accent.Render("new best!"))
		}
		b.WriteString("\n\n")
	}

	statStyle := lipgloss.NewStyle().Width(14).Background(s.bg)
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

	if spark := sparkline(m.wpmHistory, 40); spark != "" {
		b.WriteString(s.accent.Render(spark))
		b.WriteString("\n\n")
	}

	b.WriteString(s.statLabel.Render("characters ") + s.fg.Render(fmt.Sprintf("%d", r.CharsCorrect)) +
		s.muted.Render("/") + s.err.Render(fmt.Sprintf("%d", r.CharsIncorrect)) +
		s.muted.Render("/") + s.accent.Render(fmt.Sprintf("%d", r.CharsExtra)) +
		s.muted.Render("/") + s.muted.Render(fmt.Sprintf("%d", r.CharsMissing)))
	b.WriteString("\n")

	if !r.IsPR && r.PrevBest > 0 {
		b.WriteString(s.statLabel.Render("best       ") + s.muted.Render(fmt.Sprintf("%.0f wpm", r.PrevBest)))
		b.WriteString("\n")
	}

	if weak := m.topWeakKeys(5); len(weak) > 0 {
		var w strings.Builder
		for i, wk := range weak {
			if i > 0 {
				w.WriteString(s.muted.Render(" "))
			}
			w.WriteString(s.fg.Render(wk.char) + s.muted.Render(fmt.Sprintf("(%d)", wk.count)))
		}
		b.WriteString(s.statLabel.Render("weak keys  ") + w.String())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(hintLine(s, "tab", " restart   ", "esc", " quit"))
	return b.String()
}
