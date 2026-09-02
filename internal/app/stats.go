package app

import "time"

type Result struct {
	WPM       float64
	RawWPM    float64
	Accuracy  float64
	Correct   int
	Incorrect int
	Duration  time.Duration
}

func computeResult(completed []completedWord, current string, target string, keyCorrect, keyIncorrect int, dur time.Duration) Result {
	if dur <= 0 {
		dur = time.Millisecond
	}
	minutes := dur.Minutes()

	correctChars := 0
	rawChars := 0

	for _, w := range completed {
		rawChars += len(w.typed) + 1 // +1 for the trailing space
		correctChars += matchingChars(w.typed, w.target)
		if w.typed == w.target {
			correctChars++ // the space after a fully correct word counts too
		}
	}
	if current != "" {
		rawChars += len(current)
		correctChars += matchingChars(current, target)
	}

	if minutes <= 0 {
		minutes = float64(time.Millisecond) / float64(time.Minute)
	}

	res := Result{
		RawWPM:    float64(rawChars) / 5 / minutes,
		WPM:       float64(correctChars) / 5 / minutes,
		Correct:   keyCorrect,
		Incorrect: keyIncorrect,
		Duration:  dur,
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
