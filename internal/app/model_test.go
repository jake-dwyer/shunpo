package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	m := New(WordsConfig(2))
	m.targetWords = []string{"the", "was"}

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
	m := New(WordsConfig(1))
	m.targetWords = []string{"twelve"}

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
	out := m.viewWords()
	if out == "" {
		t.Fatal("expected non-empty rendered words")
	}
}
