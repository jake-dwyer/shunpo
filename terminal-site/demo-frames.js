// Pre-rendered animation frames for the homepage terminal demo.
// app.js steps through these over a ~10s cycle, then holds on the
// final (results) frame for 2.5s before resetting. Markup reuses the
// site's .ansi-fg-*/.ansi-bg-panel classes, which mirror the real
// app's own color roles (fg/muted/accent/error/panel) so a theme swap
// just needs the CSS custom properties, not different markup.
const demoFrames = (() => {
  const esc = s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  const span = (cls, text) => `<span class="${cls}">${esc(text)}</span>`;
  const fg = t => span('ansi-fg-fg', t);
  const muted = t => span('ansi-fg-muted', t);
  const accent = t => span('ansi-fg-accent ansi-bold', t);
  const err = t => span('ansi-fg-error ansi-bold', t);
  const kbd = t => span('ansi-bg-panel ansi-fg-fg', ` ${t} `);

  const header = seconds =>
    `${accent('瞬歩 shunpo')}  ${accent('✿')}  ${accent('time 30')}\n\n${accent(String(seconds))}\n\n`;
  const hint = `\n\n${kbd('esc')} quit   ${kbd('enter')} restart   ${kbd('tab')} settings`;

  const WORDS = [
    'necessary', 'fact', 'ran', 'fast', 'sand', 'when', 'draw', 'multiply',
    'family', 'cross', 'lie', 'too', 'position', 'speak', 'quick', 'water',
    'floor', 'hair', 'are', 'rock',
  ];

  function wordsLine(typedCount, mistakeAt, current) {
    return WORDS.map((w, i) => {
      if (i === mistakeAt) {
        return fg(w.slice(0, 2)) + err(w[2] === 'a' ? 'e' : 'a') + muted(w.slice(3));
      }
      if (i < typedCount) return fg(w);
      if (i === current) return accent('|') + muted(w);
      return muted(w);
    }).join(' ');
  }

  const frames = [];
  const steps = 9;
  for (let i = 0; i < steps; i++) {
    const seconds = Math.max(30 - i * 3, 6);
    const mistakeAt = i === 5 ? 5 : -1;
    frames.push(header(seconds) + wordsLine(i, mistakeAt, i) + hint);
  }

  frames.push(
    `${accent('瞬歩 shunpo')}\n\n` +
    `${accent('new best!')}  ${muted('(was 96 wpm)')}\n\n` +
    `${muted('wpm')}        ${muted('acc')}        ${muted('raw')}        ${muted('time')}\n` +
    `${accent('104')}        ${accent('97%')}       ${accent('109')}       ${accent('30s')}\n\n` +
    `${accent('▂▃▅▆▆▇█▇▆▇█▇▆▇█▇▆▅▆▇█▇▆▇█▇▆▇█▇')}\n\n` +
    `${muted('characters')} ${fg('312')}${muted('/')}${err('9')}${muted('/')}${accent('2')}${muted('/')}${muted('0')}\n` +
    `${muted('weak keys')}  ${fg('e')}${muted('(14)')} ${fg('t')}${muted('(9)')} ${fg('i')}${muted('(6)')}\n\n` +
    `${kbd('tab')} restart   ${kbd('esc')} quit`
  );

  return frames;
})();
