# clackity

A [monkeytype](https://monkeytype.com)-style typing test for your terminal, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Made for typing tests in a spare terminal pane while something else (a build, a long-running agent, whatever) is working.

![mode](https://img.shields.io/badge/mode-time%20%7C%20words-e2b714)

## Install

```
go install github.com/jake-dwyer/clackity@latest
```

Or build from source:

```
git clone https://github.com/jake-dwyer/clackity
cd clackity
go build -o clackity .
```

## Usage

```
clackity                # 30 second test (default)
clackity -time 60       # 60 second test
clackity -words 50      # 50 word test
```

While typing:

- Just type — words are colored as you go (correct/incorrect), extra characters are underlined.
- `esc` / `ctrl+c` — quit
- `tab` — restart with a fresh set of words (same mode/settings)

Mode and duration/word-count are set via flags rather than in-app hotkeys — a "press `1`-`4`/`t`/`w` to change settings" scheme would hijack real keystrokes any time a word in the test happens to start with one of those characters.

## Stats

After a test finishes you get:

- **wpm** — net words per minute (correct characters / 5 / minutes)
- **acc** — accuracy across every keystroke, including ones you later fixed
- **raw** — raw words per minute (all typed characters / 5 / minutes, no correctness penalty)
- **time** — test duration

## Development

```
go test ./...
go build ./...
```
