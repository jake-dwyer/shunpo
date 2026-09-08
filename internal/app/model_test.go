package app

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/jake-dwyer/shunpo/internal/store"
)

func send(m Model, msg tea.KeyMsg) Model {
	next, _ := m.Update(msg)
	return next.(Model)
}

func typeString(m Model, s string) Model {
	for _, r := range s {
		if r == ' ' {
			m = send(m, tea.KeyMsg{Type: tea.KeySpace})
			continue
		}
		m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestTypeCorrectWord(t *testing.T) {
	m := New(WordsConfig(3))
	m.targetWords = []string{"the", "of", "and"}

	m = typeString(m, "the ")

	if m.state != stateTyping {
		t.Fatalf("expected stateTyping, got %v", m.state)
	}
	if m.wordIdx != 1 {
		t.Fatalf("expected wordIdx 1, got %d", m.wordIdx)
	}
	if m.keyCorrect != 3 || m.keyIncorrect != 0 {
		t.Fatalf("expected 3 correct 0 incorrect, got %d/%d", m.keyCorrect, m.keyIncorrect)
	}
}

func TestWordsStartingWithMenuLettersAndDigitsTypeNormally(t *testing.T) {
	// "the" and "was" start with letters that double as menu commands
	// ('t'/'w') when the settings palette is open; outside the palette
	// they must just type. Also make sure a raw digit types fine.
	m := New(WordsConfig(3))
	m.targetWords = []string{"the", "was", "end"}

	m = typeString(m, "the was")

	if m.keyIncorrect != 0 {
		t.Fatalf("expected 0 incorrect, got %d", m.keyIncorrect)
	}
	if m.current != "was" {
		t.Fatalf("expected current 'was', got %q", m.current)
	}
	if m.menuFocused {
		t.Fatal("typing 't'/'w' outside the palette must not open it")
	}
}

func TestBackspaceAcrossWordBoundary(t *testing.T) {
	m := New(WordsConfig(3))
	m.targetWords = []string{"the", "of", "and"}

	m = typeString(m, "the ")
	m = send(m, tea.KeyMsg{Type: tea.KeyBackspace})

	if m.wordIdx != 0 {
		t.Fatalf("expected wordIdx back to 0, got %d", m.wordIdx)
	}
	if m.current != "the" {
		t.Fatalf("expected current to restore 'the', got %q", m.current)
	}
}

func TestFinishOnLastWord(t *testing.T) {
	m := New(WordsConfig(2))
	m.targetWords = []string{"the", "of"}

	m = typeString(m, "the of ")

	if m.state != stateDone {
		t.Fatalf("expected stateDone, got %v", m.state)
	}
	if m.result.Accuracy != 100 {
		t.Fatalf("expected 100%% accuracy, got %v", m.result.Accuracy)
	}
	if m.result.WPM <= 0 {
		t.Fatalf("expected positive WPM, got %v", m.result.WPM)
	}
}

func TestIncorrectCharTracked(t *testing.T) {
	m := New(WordsConfig(1))
	m.targetWords = []string{"the"}

	m = typeString(m, "tha")

	if m.keyIncorrect != 1 || m.keyCorrect != 2 {
		t.Fatalf("expected 2 correct 1 incorrect, got %d/%d", m.keyCorrect, m.keyIncorrect)
	}
}

func TestTimeModeFinishesOnTick(t *testing.T) {
	m := New(TimeConfig(15))
	m.targetWords = []string{"the", "of", "and"}

	m = typeString(m, "the ")
	m.start = time.Now().Add(-16 * time.Second)

	next, _ := m.Update(tickMsg(time.Now()))
	m = next.(Model)

	if m.state != stateDone {
		t.Fatalf("expected stateDone after time limit exceeded, got %v", m.state)
	}
}

func TestEnterRestartsWithFreshWords(t *testing.T) {
	m := New(WordsConfig(3))
	m.targetWords = []string{"the", "of", "and"}
	m = typeString(m, "the")

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	if m.state != stateReady || m.current != "" || m.wordIdx != 0 {
		t.Fatalf("expected reset state after enter, got state=%v current=%q wordIdx=%d", m.state, m.current, m.wordIdx)
	}
}

func TestTabOpensMenuWithoutLosingProgress(t *testing.T) {
	m := New(WordsConfig(3))
	m.targetWords = []string{"apple", "of", "and"}
	m = typeString(m, "app")

	m = send(m, tea.KeyMsg{Type: tea.KeyTab})

	if !m.menuFocused {
		t.Fatal("expected tab to open the settings palette")
	}
	if m.current != "app" {
		t.Fatalf("expected in-progress word preserved while palette opens, got %q", m.current)
	}
}

func TestMenuPresetSelectionAppliesAndCloses(t *testing.T) {
	m := New(TimeConfig(30)) // default idx points at 30 in timePresets

	m = send(m, tea.KeyMsg{Type: tea.KeyTab}) // open palette
	if !m.menuFocused {
		t.Fatal("expected palette open")
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'1'}}) // preset index 0 -> 10s
	if m.cfg.Seconds != timePresets[0] {
		t.Fatalf("expected seconds %d, got %d", timePresets[0], m.cfg.Seconds)
	}
	if !m.menuFocused {
		t.Fatal("palette should stay open after a selection so multiple changes can be made")
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter}) // close palette
	if m.menuFocused {
		t.Fatal("expected enter to close the palette")
	}
	if m.state != stateReady {
		t.Fatalf("expected stateReady after closing palette, got %v", m.state)
	}
}

func TestMenuKeysDoNotLeakIntoTypingWhenClosed(t *testing.T) {
	m := New(WordsConfig(2))
	m.targetWords = []string{"twelve", "end"}

	m = typeString(m, "twelve")

	if m.current != "twelve" {
		t.Fatalf("expected 'twelve' typed literally, got %q", m.current)
	}
}

func TestEscClosesMenuWithoutQuitting(t *testing.T) {
	m := New(TimeConfig(30))
	m = send(m, tea.KeyMsg{Type: tea.KeyTab})
	if !m.menuFocused {
		t.Fatal("expected palette open")
	}

	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if cmd != nil {
		t.Fatal("expected esc to just close the palette, not quit, while it's open")
	}
	if m.menuFocused {
		t.Fatal("expected esc to close the palette")
	}
}

func TestTabOnResultsScreenRestartsImmediately(t *testing.T) {
	m := New(WordsConfig(1))
	m.targetWords = []string{"go"}
	m = typeString(m, "go ")
	if m.state != stateDone {
		t.Fatalf("expected stateDone, got %v", m.state)
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyTab})

	if m.state != stateReady {
		t.Fatalf("expected tab to restart immediately from results, got %v", m.state)
	}
	if m.menuFocused {
		t.Fatal("tab from results should restart, not open the palette")
	}
}

func TestViewWordsWraps(t *testing.T) {
	m := New(TimeConfig(30))
	m.width = 80
	out := m.viewWords(m.styles())
	if out == "" {
		t.Fatal("expected non-empty rendered words")
	}
}

func TestFinishesOnLastWordWithoutTrailingSpace(t *testing.T) {
	m := New(WordsConfig(2))
	m.targetWords = []string{"the", "of"}

	m = typeString(m, "the o")
	if m.state != stateReady && m.state != stateTyping {
		t.Fatalf("expected test still running after first word + partial last word, got %v", m.state)
	}

	m = typeString(m, "f")
	if m.state != stateDone {
		t.Fatalf("expected stateDone immediately once the last word reaches full length, got %v", m.state)
	}
	if m.result.WPM <= 0 {
		t.Fatalf("expected positive WPM, got %v", m.result.WPM)
	}
}

func TestThemeCycleWithinMenu(t *testing.T) {
	m := New(TimeConfig(30))
	m = send(m, tea.KeyMsg{Type: tea.KeyTab}) // open palette
	start := m.themeIdx

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})

	if m.themeIdx == start && len(themes) > 1 {
		t.Fatal("expected theme index to advance after pressing c in the palette")
	}
	if !m.menuFocused {
		t.Fatal("theme cycling should not close the palette")
	}
}

func TestNoDoubleSpaceAfterWordCompletion(t *testing.T) {
	// Regression: the separator after a fully-typed current word used to
	// be rendered twice (once as an inline cursor block, once as a plain
	// space), shifting every following word over by one column.
	m := New(WordsConfig(3))
	m.width = 80
	m.targetWords = []string{"the", "of", "and"}
	m = typeString(m, "the")

	out := m.viewWords(m.styles())
	if strings.Contains(stripANSI(out), "  ") {
		t.Fatalf("expected no double space in rendered words, got %q", stripANSI(out))
	}
}

func TestPlainNewDoesNotPersist(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := New(WordsConfig(1))
	m.targetWords = []string{"go"}
	m = typeString(m, "go")
	if m.state != stateDone {
		t.Fatalf("expected stateDone, got %v", m.state)
	}

	loaded := store.Load()
	if len(loaded.Records) != 0 {
		t.Fatalf("expected New() (non-persisting) to leave disk state untouched, got %+v", loaded.Records)
	}
}

func TestMenuChangePersistsSettings(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := NewWithStore(TimeConfig(30), store.Store{Records: map[string]store.Record{}, Typos: map[string]int{}})
	m = send(m, tea.KeyMsg{Type: tea.KeyTab})                       // open palette
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}}) // cycle theme

	loaded := store.Load()
	wantTheme := themes[m.themeIdx].Name
	if loaded.Settings.Theme != wantTheme {
		t.Fatalf("expected persisted theme %q, got %q", wantTheme, loaded.Settings.Theme)
	}
}

func TestFinishRecordsPRAndTypos(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := NewWithStore(WordsConfig(1), store.Store{Records: map[string]store.Record{}, Typos: map[string]int{}})
	m.targetWords = []string{"the"}
	m = typeString(m, "the")

	if m.state != stateDone {
		t.Fatalf("expected stateDone, got %v", m.state)
	}
	if !m.result.IsPR {
		t.Fatal("expected first completion of a config to count as a new best")
	}

	key := recordKey(m.cfg)
	loaded := store.Load()
	rec, ok := loaded.Records[key]
	if !ok || rec.WPM != m.result.WPM {
		t.Fatalf("expected persisted record for %q matching result, got %+v ok=%v", key, rec, ok)
	}
}

func TestFinishTalliesTyposByTargetChar(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	m := NewWithStore(WordsConfig(1), store.Store{Records: map[string]store.Record{}, Typos: map[string]int{}})
	m.targetWords = []string{"the"}
	m = typeString(m, "tha") // 'a' mistyped for target 'e' at index 2

	if m.state != stateDone {
		t.Fatalf("expected stateDone, got %v", m.state)
	}
	loaded := store.Load()
	if loaded.Typos["e"] != 1 {
		t.Fatalf("expected typo tally for 'e' to be 1, got %d (full map %+v)", loaded.Typos["e"], loaded.Typos)
	}
}

func TestSparklineProducesRequestedWidth(t *testing.T) {
	out := sparkline([]float64{10, 20, 30, 20, 10}, 8)
	if len([]rune(out)) != 8 {
		t.Fatalf("expected sparkline of width 8, got %d runes (%q)", len([]rune(out)), out)
	}
	if sparkline(nil, 8) != "" {
		t.Fatal("expected empty sparkline for no data")
	}
}

func TestTopWeakKeysSortedDescending(t *testing.T) {
	m := New(WordsConfig(1))
	m.store.Typos = map[string]int{"e": 5, "t": 9, "a": 1}

	top := m.topWeakKeys(2)
	if len(top) != 2 || top[0].char != "t" || top[1].char != "e" {
		t.Fatalf("expected [t(9) e(5)], got %+v", top)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func typeRunes(m Model, s string) Model {
	for _, r := range s {
		m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func TestThemeEditorOpensFromPalette(t *testing.T) {
	m := New(TimeConfig(30))
	m = send(m, tea.KeyMsg{Type: tea.KeyTab}) // open palette
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	if !m.themeEditorOpen {
		t.Fatal("expected 'e' in the palette to open the theme editor")
	}
	if m.editTheme.Name != themes[m.themeIdx].Name {
		t.Fatalf("expected editTheme to start as the active theme, got %q want %q", m.editTheme.Name, themes[m.themeIdx].Name)
	}
}

func TestThemeEditorEditsHexLiveAndCommitsOnEnter(t *testing.T) {
	m := New(TimeConfig(30))
	m = send(m, tea.KeyMsg{Type: tea.KeyTab})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	// editField 0 == bg
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter}) // start editing hex

	if !m.editingHex {
		t.Fatal("expected enter to start hex editing")
	}

	m = typeRunes(m, "ff00aa")
	if m.editTheme.Bg != "#ff00aa" {
		t.Fatalf("expected live preview to update while typing, got %q", m.editTheme.Bg)
	}

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter}) // commit
	if m.editingHex {
		t.Fatal("expected enter to stop hex editing")
	}
	if m.editTheme.Bg != "#ff00aa" {
		t.Fatalf("expected committed bg #ff00aa, got %q", m.editTheme.Bg)
	}
}

// Typing all 6 digits applies live (covered by
// TestThemeEditorEditsHexLiveAndCommitsOnEnter) - that's the point of a
// live preview. Esc only truly "cancels" an incomplete entry, since a
// completed one is already applied by the time you'd press it.
func TestThemeEditorEscOnIncompleteHexLeavesValueUntouched(t *testing.T) {
	m := New(TimeConfig(30))
	m = send(m, tea.KeyMsg{Type: tea.KeyTab})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	original := m.editTheme.Bg

	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = typeRunes(m, "abc") // fewer than 6 digits - never applied
	m = send(m, tea.KeyMsg{Type: tea.KeyEsc})

	if m.editingHex {
		t.Fatal("expected esc to stop hex editing")
	}
	if m.editTheme.Bg != original {
		t.Fatalf("expected incomplete edit to leave bg unchanged at %q, got %q", original, m.editTheme.Bg)
	}
	if !m.themeEditorOpen {
		t.Fatal("expected esc from hex-editing to only cancel the field, not close the editor")
	}
}

func TestThemeEditorSavesOverrideOnClose(t *testing.T) {
	m := New(TimeConfig(30))
	themeName := themes[m.themeIdx].Name
	m = send(m, tea.KeyMsg{Type: tea.KeyTab})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = typeRunes(m, "123456")
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter}) // commit field

	m = send(m, tea.KeyMsg{Type: tea.KeyEsc}) // close editor, save override

	if m.themeEditorOpen {
		t.Fatal("expected esc to close the editor")
	}
	ov, ok := m.store.ThemeOverrides[themeName]
	if !ok || ov.Bg != "#123456" {
		t.Fatalf("expected saved override with bg #123456, got %+v ok=%v", ov, ok)
	}
	if m.theme().Bg != "#123456" {
		t.Fatalf("expected m.theme() to reflect the saved override, got %q", m.theme().Bg)
	}
}

func TestThemeEditorResetRemovesOverride(t *testing.T) {
	m := New(TimeConfig(30))
	themeName := themes[m.themeIdx].Name
	factoryBg := themes[m.themeIdx].Bg

	m = send(m, tea.KeyMsg{Type: tea.KeyTab})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})
	m = typeRunes(m, "654321")
	m = send(m, tea.KeyMsg{Type: tea.KeyEnter})

	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}) // reset

	if _, ok := m.store.ThemeOverrides[themeName]; ok {
		t.Fatal("expected reset to remove the stored override")
	}
	if m.editTheme.Bg != factoryBg {
		t.Fatalf("expected editTheme reset to factory bg %q, got %q", factoryBg, m.editTheme.Bg)
	}
}

func TestThemeEditorArrowsNavigateFieldsWithWraparound(t *testing.T) {
	m := New(TimeConfig(30))
	m = send(m, tea.KeyMsg{Type: tea.KeyTab})
	m = send(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})

	if m.editField != 0 {
		t.Fatalf("expected editField to start at 0, got %d", m.editField)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyUp}) // wrap to last field
	if m.editField != 4 {
		t.Fatalf("expected wraparound to field 4, got %d", m.editField)
	}
	m = send(m, tea.KeyMsg{Type: tea.KeyDown}) // wrap back to first
	if m.editField != 0 {
		t.Fatalf("expected wraparound back to field 0, got %d", m.editField)
	}
}
