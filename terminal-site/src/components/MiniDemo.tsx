import { useEffect, useRef, useState } from 'react'

// Steps through WORDS one character at a time (with a single simulated
// mistake + correction), then holds a short results line before looping.
// Pauses when off-screen, hidden, or reduced motion is set.
const WORD_POOL = [
  'necessary', 'fact', 'ran', 'fast', 'sand', 'when', 'draw', 'multiply',
  'family', 'cross', 'lie', 'too', 'position', 'speak', 'quick', 'water',
  'floor', 'hair', 'are', 'rock', 'build', 'agent', 'pane', 'spare',
  'flow', 'finger', 'light', 'string', 'focus', 'branch',
]
const WORD_LEN = 10

function pickWords(): string[] {
  const pool = [...WORD_POOL]
  for (let i = pool.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[pool[i], pool[j]] = [pool[j], pool[i]]
  }
  return pool.slice(0, WORD_LEN)
}

interface MistakeSpot {
  wordIndex: number
  charIndex: number
  used: boolean
}

function pickMistake(words: string[]): MistakeSpot {
  const candidates = words.map((w, i) => ({ i, w })).filter(({ w }) => w.length >= 3)
  const { i, w } = candidates[Math.floor(Math.random() * candidates.length)]
  return { wordIndex: i, charIndex: Math.floor(Math.random() * (w.length - 2)) + 1, used: false }
}

type Phase = 'typing' | 'result'

interface DemoState {
  words: string[]
  mistakeSpot: MistakeSpot
  wordIndex: number
  charIndex: number
  mistake: string | null
  phase: Phase
  seconds: number
}

const START_SECONDS = 30
const MIN_SECONDS = 4

function initialState(): DemoState {
  const words = pickWords()
  return {
    words,
    mistakeSpot: pickMistake(words),
    wordIndex: 0,
    charIndex: 0,
    mistake: null,
    phase: 'typing',
    seconds: START_SECONDS,
  }
}

function reset(): DemoState {
  return initialState()
}

function step(state: DemoState): DemoState {
  const word = state.words[state.wordIndex]
  const { mistakeSpot } = state
  const atMistake = !mistakeSpot.used && state.wordIndex === mistakeSpot.wordIndex && state.charIndex === mistakeSpot.charIndex

  if (atMistake && !state.mistake) {
    return {
      ...state,
      mistakeSpot: { ...mistakeSpot, used: true },
      mistake: word[state.charIndex] === 'e' ? 'r' : 'e',
    }
  }

  let { wordIndex, charIndex, phase } = state
  charIndex++
  if (charIndex >= word.length) {
    wordIndex++
    charIndex = 0
    if (wordIndex >= WORD_LEN) phase = 'result'
  }
  return { ...state, wordIndex, charIndex, phase, mistake: null }
}

function Header({ seconds }: { seconds: number }) {
  return (
    <>
      <span className="text-accent">瞬歩 shunpo</span>
      {'  '}
      <span className="text-[#55585a]">✿</span>
      {'  '}
      <span className="text-accent">{`time 30`}</span>
      {'\n\n'}
      <span className="text-accent">{seconds}</span>
      {'\n\n'}
    </>
  )
}

function TypingView({ state }: { state: DemoState }) {
  return (
    <>
      <Header seconds={state.seconds} />
      {state.words.map((w, i) => {
        if (i < state.wordIndex) return <span key={i}>{w} </span>
        if (i === state.wordIndex) {
          return (
            <span key={i}>
              {w.slice(0, state.charIndex)}
              {state.mistake && <span className="text-err">{state.mistake}</span>}
              <span className="text-accent">|</span>
              <span className="text-[#55585a]">{w.slice(state.charIndex)}</span>{' '}
            </span>
          )
        }
        return (
          <span key={i} className="text-[#55585a]">
            {w}{' '}
          </span>
        )
      })}
      {'\n\n'}
      <span className="text-[#55585a]">esc</span> quit <span className="text-[#55585a]">enter</span> restart{' '}
      <span className="text-[#55585a]">tab</span> settings
    </>
  )
}

function ResultView() {
  return (
    <>
      <span className="text-accent">瞬歩 shunpo</span>
      {'\n\n'}
      <span className="text-accent">new best!</span> <span className="text-[#55585a]">(was 96 wpm)</span>
      {'\n\n'}
      <span className="text-[#55585a]">wpm</span>   <span className="text-[#55585a]">acc</span>   <span className="text-[#55585a]">raw</span>   <span className="text-[#55585a]">time</span>
      {'\n'}
      <span className="text-accent">104</span>   <span className="text-accent">97%</span>  <span className="text-accent">109</span>  <span className="text-accent">30s</span>
      {'\n\n'}
      <span className="text-[#55585a]">tab</span> restart   <span className="text-[#55585a]">esc</span> quit
    </>
  )
}

export function MiniDemo() {
  const reduceMotion = typeof window !== 'undefined' && window.matchMedia('(prefers-reduced-motion: reduce)').matches
  const [state, setState] = useState<DemoState>(() => {
    const s = initialState()
    return reduceMotion ? { ...s, phase: 'result' } : s
  })
  const ref = useRef<HTMLPreElement>(null)

  useEffect(() => {
    const motionPreference = window.matchMedia('(prefers-reduced-motion: reduce)')
    if (motionPreference.matches) return

    let timer: ReturnType<typeof setTimeout>
    // The delay before each tick depends on the state THAT TICK IS ABOUT TO
    // PRODUCE (e.g. holding 2600ms once phase becomes 'result'), so `current`
    // is threaded through explicitly rather than read back from a ref after
    // setState - a ref wouldn't reflect the new state until React re-renders,
    // which hasn't happened yet at this point in the same callback.
    const schedule = (current: DemoState) => {
      const delay = current.phase === 'result' ? 2600 : current.mistake ? 260 : 60 + Math.random() * 90
      timer = setTimeout(() => {
        const section = ref.current?.closest('section')
        const visible = !document.hidden && !!section && !section.hidden
        if (!visible) {
          schedule(current)
          return
        }
        const startingNewCycle = current.phase === 'result'
        const next = step(startingNewCycle ? reset() : current)
        // `seconds` is owned by the real-time countdown effect below, not
        // by this char-by-char progression - preserve whatever it just set,
        // except when a fresh cycle starts and the countdown really does
        // reset to the top.
        setState((prev) => (startingNewCycle ? next : { ...next, seconds: prev.seconds }))
        schedule(next)
      }, delay)
    }
    schedule(state)

    const onMotionChange = (event: MediaQueryListEvent) => {
      if (!event.matches) return
      clearTimeout(timer)
      setState((prev) => ({ ...prev, phase: 'result' }))
    }
    motionPreference.addEventListener('change', onMotionChange)

    return () => {
      clearTimeout(timer)
      motionPreference.removeEventListener('change', onMotionChange)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // The countdown reflects real elapsed time, not how many characters have
  // been typed, so it ticks on its own independent 1-second interval rather
  // than inside step().
  useEffect(() => {
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    const interval = setInterval(() => {
      const section = ref.current?.closest('section')
      if (document.hidden || !section || section.hidden) return
      setState((prev) => (prev.phase === 'typing' ? { ...prev, seconds: Math.max(prev.seconds - 1, MIN_SECONDS) } : prev))
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  return (
    <>
      <pre ref={ref} aria-hidden="true" className="m-0 whitespace-pre-wrap">
        {state.phase === 'result' ? <ResultView /> : <TypingView state={state} />}
      </pre>
      <span className="sr-only">Example of shunpo: words typed with a corrected mistake, then results.</span>
    </>
  )
}
