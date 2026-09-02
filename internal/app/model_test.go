package app

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func typeString(m Model, s string) Model {
	for _, r := range s {
		if r == ' ' {
			next, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
			m = next.(Model)
			continue
		}
		next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = next.(Model)
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

func TestWordsStartingWithHotkeyLettersTypeNormally(t *testing.T) {
	// "the" and "was" start with letters that used to double as mode
	// hotkeys ('t'/'w'); make sure typing them just types them.
	m := New(WordsConfig(2))
	m.targetWords = []string{"the", "was"}

	m = typeString(m, "the was")

	if m.keyIncorrect != 0 {
		t.Fatalf("expected 0 incorrect, got %d", m.keyIncorrect)
	}
	if m.current != "was" {
		t.Fatalf("expected current 'was', got %q", m.current)
	}
}

func TestBackspaceAcrossWordBoundary(t *testing.T) {
	m := New(WordsConfig(3))
	m.targetWords = []string{"the", "of", "and"}

	m = typeString(m, "the ")
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(Model)

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

func TestTabRestartsWithFreshWords(t *testing.T) {
	m := New(WordsConfig(3))
	m.targetWords = []string{"the", "of", "and"}
	m = typeString(m, "the")

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = next.(Model)

	if m.state != stateReady || m.current != "" || m.wordIdx != 0 {
		t.Fatalf("expected reset state after tab, got state=%v current=%q wordIdx=%d", m.state, m.current, m.wordIdx)
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
