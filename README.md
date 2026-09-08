# shunpo

*Flash step.* A [monkeytype](https://monkeytype.com)-style typing test for your terminal, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Made for typing tests in a spare terminal pane while something else (a build, a long-running agent, whatever) is working.

![mode](https://img.shields.io/badge/mode-time%20%7C%20words-e2b714)

## Install

```
go install github.com/jake-dwyer/shunpo@latest
```

Or build from source:

```
git clone https://github.com/jake-dwyer/shunpo
cd shunpo
go build -o shunpo .
```

## Usage

```
shunpo                    # 30 second test (default)
shunpo -time 60           # 60 second test
shunpo -words 50          # 50 word test
shunpo -theme dracula     # start with a specific theme
```

While typing:

- Just type — words are colored as you go (correct/incorrect), extra characters are underlined.
- Word mode finishes the instant the last word reaches full length — no need to hit space after it.
- `esc` — quit (or close the settings palette, if it's open) · `ctrl+c` — quit, always
- `enter` — restart with a fresh set of words (same settings)
- `tab` — open the settings palette (or, from the results screen, restart immediately)

### Settings palette (`tab`)

Everything is keyboard-driven — no mouse required. With the palette open:

- `t` / `w` — switch between time mode and word-count mode
- `1`-`5` — pick a preset for whichever mode is active
- `c` — cycle color theme (serika, dracula, nord, gruvbox, monokai, tokyonight, catppuccin, solarized)
- `enter` / `esc` / `tab` — close the palette

Settings apply immediately as you change them, so you can keep adjusting before closing the palette. This all lives behind an explicit `tab` press rather than always-on hotkeys — letting `1`-`5`/`t`/`w` mean something while you're mid-word would hijack real keystrokes any time a test word happens to start with one of those characters.

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
