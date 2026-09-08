package app

import "time"

type Result struct {
	WPM       float64
	RawWPM    float64
	Accuracy  float64
	Correct   int // keystroke-level correct count, including ones later fixed
	Incorrect int // keystroke-level incorrect count, including ones later fixed
	Duration  time.Duration

	// Final-state character breakdown (monkeytype's "characters" grid):
	// counted from what actually ended up submitted, not every keystroke.
	CharsCorrect   int
	CharsIncorrect int
	CharsExtra     int
	CharsMissing   int

	IsPR     bool
	PrevBest float64
}

func computeResult(completed []completedWord, current string, target string, keyCorrect, keyIncorrect int, dur time.Duration) Result {
	if dur <= 0 {
		dur = time.Millisecond
	}
	minutes := dur.Minutes()

	correctChars := 0
	rawChars := 0
	finalCorrect, finalIncorrect, extra, missing := 0, 0, 0, 0

	tally := func(typed, tgt string) {
		n := len(typed)
		if len(tgt) < n {
			n = len(tgt)
		}
		for i := 0; i < n; i++ {
			if typed[i] == tgt[i] {
				finalCorrect++
			} else {
				finalIncorrect++
			}
		}
		if len(typed) > len(tgt) {
			extra += len(typed) - len(tgt)
		}
		if len(typed) < len(tgt) {
			missing += len(tgt) - len(typed)
		}
	}

	for _, w := range completed {
		rawChars += len(w.typed) + 1 // +1 for the trailing space
		correctChars += matchingChars(w.typed, w.target)
		if w.typed == w.target {
			correctChars++ // the space after a fully correct word counts too
		}
		tally(w.typed, w.target)
	}
	if current != "" {
		rawChars += len(current)
		correctChars += matchingChars(current, target)
		tally(current, target)
	}

	if minutes <= 0 {
		minutes = float64(time.Millisecond) / float64(time.Minute)
	}

	res := Result{
		RawWPM:         float64(rawChars) / 5 / minutes,
		WPM:            float64(correctChars) / 5 / minutes,
		Correct:        keyCorrect,
		Incorrect:      keyIncorrect,
		Duration:       dur,
		CharsCorrect:   finalCorrect,
		CharsIncorrect: finalIncorrect,
		CharsExtra:     extra,
		CharsMissing:   missing,
	}
	total := keyCorrect + keyIncorrect
	if total > 0 {
		res.Accuracy = float64(keyCorrect) / float64(total) * 100
	} else {
		res.Accuracy = 100
	}
	return res
}

func matchingChars(typed, target string) int {
	n := 0
	for i := 0; i < len(typed) && i < len(target); i++ {
		if typed[i] == target[i] {
			n++
		}
	}
	return n
}
