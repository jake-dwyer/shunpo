// Package store persists shunpo's cross-session state: the user's last
// settings (mode/duration/theme), a personal-best record per test config,
// and a running tally of which characters get mistyped most often.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	Mode     string `json:"mode"`
	Seconds  int    `json:"seconds"`
	WordGoal int    `json:"wordGoal"`
	Theme    string `json:"theme"`
}

type Record struct {
	WPM      float64 `json:"wpm"`
	Accuracy float64 `json:"accuracy"`
	Date     string  `json:"date"`
}

type Store struct {
	Settings Settings          `json:"settings"`
	Records  map[string]Record `json:"records"`
	Typos    map[string]int    `json:"typos"`
}

func empty() Store {
	return Store{Records: map[string]Record{}, Typos: map[string]int{}}
}

// RecordKey identifies a test configuration for personal-best tracking,
// e.g. "time:30" or "words:50" — records are kept per exact mode+length,
// matching how monkeytype tracks a separate best for each test length.
func RecordKey(mode string, value int) string {
	return fmt.Sprintf("%s:%d", mode, value)
}

func path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "shunpo", "state.json"), nil
}

// Load reads persisted state from disk, returning an empty (but non-nil)
// Store if nothing has been saved yet or the file can't be read/parsed.
func Load() Store {
	p, err := path()
	if err != nil {
		return empty()
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return empty()
	}
	st := empty()
	if err := json.Unmarshal(data, &st); err != nil {
		return empty()
	}
	if st.Records == nil {
		st.Records = map[string]Record{}
	}
	if st.Typos == nil {
		st.Typos = map[string]int{}
	}
	return st
}

// Save writes state to disk, creating the config directory if needed.
func (s Store) Save() error {
	p, err := path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}
