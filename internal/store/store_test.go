package store

import "testing"

func TestSaveLoadRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	st := empty()
	st.Settings = Settings{Mode: "time", Seconds: 60, Theme: "dracula"}
	st.Records[RecordKey("time", 60)] = Record{WPM: 80, Accuracy: 97, Date: "2026-09-08"}
	st.Typos["e"] = 3

	if err := st.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded := Load()
	if loaded.Settings != st.Settings {
		t.Fatalf("settings mismatch: got %+v want %+v", loaded.Settings, st.Settings)
	}
	rec, ok := loaded.Records[RecordKey("time", 60)]
	if !ok || rec.WPM != 80 {
		t.Fatalf("expected record to round-trip, got %+v ok=%v", rec, ok)
	}
	if loaded.Typos["e"] != 3 {
		t.Fatalf("expected typo count 3, got %d", loaded.Typos["e"])
	}
}

func TestLoadWithNoFileReturnsEmptyStore(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	st := Load()
	if st.Records == nil || st.Typos == nil {
		t.Fatal("expected non-nil maps even with no saved state")
	}
	if len(st.Records) != 0 || len(st.Typos) != 0 {
		t.Fatal("expected empty maps with no saved state")
	}
}
