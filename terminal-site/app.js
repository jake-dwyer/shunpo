const sections = [...document.querySelectorAll('main section')];
const links = [...document.querySelectorAll('nav a')];

function navigate() {
  const requested = location.hash.slice(1) || 'terminal';
  const page = sections.some(section => section.id === requested) ? requested : 'terminal';
  for (const section of sections) section.hidden = section.id !== page;
  for (const link of links) {
    const active = link.hash === `#${page}`;
    link.classList.toggle('active', active);
    if (active) link.setAttribute('aria-current', 'page');
    else link.removeAttribute('aria-current');
  }
}

window.addEventListener('hashchange', navigate);
navigate();

document.querySelectorAll('[data-copy]').forEach(button => {
  button.addEventListener('click', async () => {
    try {
      await navigator.clipboard.writeText(button.dataset.copy);
      button.textContent = '[copied]';
      document.querySelector('#announcement').textContent = 'Command copied to clipboard.';
    } catch {
      button.textContent = '[select to copy]';
      const range = document.createRange();
      range.selectNodeContents(button.previousElementSibling);
      const selection = window.getSelection();
      selection.removeAllRanges();
      selection.addRange(range);
      document.querySelector('#announcement').textContent = 'Command selected. Press your system copy shortcut.';
    }
    setTimeout(() => { button.textContent = '[copy]'; }, 2200);
  });
});

const themes = [
  {
    "name": "serika",
    "bg": "#232323",
    "panel": "#2c2c2c",
    "accent": "#e2b714",
    "fg": "#d1d0c5",
    "muted": "#646669",
    "error": "#ca4754"
  },
  {
    "name": "dracula",
    "bg": "#282a36",
    "panel": "#343746",
    "accent": "#bd93f9",
    "fg": "#f8f8f2",
    "muted": "#6272a4",
    "error": "#ff5555"
  },
  {
    "name": "nord",
    "bg": "#2e3440",
    "panel": "#3b4252",
    "accent": "#88c0d0",
    "fg": "#e5e9f0",
    "muted": "#4c566a",
    "error": "#bf616a"
  },
  {
    "name": "gruvbox",
    "bg": "#282828",
    "panel": "#3c3836",
    "accent": "#fabd2f",
    "fg": "#ebdbb2",
    "muted": "#928374",
    "error": "#fb4934"
  },
  {
    "name": "monokai",
    "bg": "#272822",
    "panel": "#3e3d32",
    "accent": "#a6e22e",
    "fg": "#f8f8f2",
    "muted": "#75715e",
    "error": "#f92672"
  },
  {
    "name": "tokyonight",
    "bg": "#1a1b26",
    "panel": "#24283b",
    "accent": "#7aa2f7",
    "fg": "#c0caf5",
    "muted": "#565f89",
    "error": "#f7768e"
  },
  {
    "name": "catppuccin",
    "bg": "#1e1e2e",
    "panel": "#313244",
    "accent": "#f5c2e7",
    "fg": "#cdd6f4",
    "muted": "#6c7086",
    "error": "#f38ba8"
  },
  {
    "name": "solarized",
    "bg": "#002b36",
    "panel": "#073642",
    "accent": "#b58900",
    "fg": "#839496",
    "muted": "#586e75",
    "error": "#dc322f"
  },
  {
    "name": "manga",
    "bg": "#0d0d0c",
    "panel": "#1a1a18",
    "accent": "#f2f0e6",
    "fg": "#c9c7ba",
    "muted": "#4a4944",
    "error": "#8b2419"
  },
  {
    "name": "chidori",
    "bg": "#090c14",
    "panel": "#131a2b",
    "accent": "#4fd1ff",
    "fg": "#eaf6ff",
    "muted": "#35415c",
    "error": "#ff7a3d"
  },
  {
    "name": "gojo",
    "bg": "#0a0f14",
    "panel": "#131c22",
    "accent": "#34e2e2",
    "fg": "#eafffb",
    "muted": "#3d5654",
    "error": "#ff5c8a"
  },
  {
    "name": "susanoo",
    "bg": "#140d08",
    "panel": "#241608",
    "accent": "#ff9d2e",
    "fg": "#f7e6cf",
    "muted": "#6b5439",
    "error": "#d1453b"
  }
];

function applyTheme(element, theme) {
  for (const key of ['bg', 'fg', 'accent', 'muted', 'error', 'panel']) {
    element.style.setProperty(`--${key}`, theme[key]);
  }
  element.querySelector('.demo-theme').textContent = theme.name;
}

const motionPreference = window.matchMedia('(prefers-reduced-motion: reduce)');
const demos = [...document.querySelectorAll('[data-demo]')].map(element => {
  element.innerHTML = `<div class="terminal-viewport"><pre class="terminal-frame" aria-hidden="true"></pre></div>
    <p class="sr-only">Example rendered by Shunpo: a timed typing test, a corrected mistake, and results with speed, accuracy, a trend chart, and weak keys.</p>
    <div class="demo-bottom"><span class="demo-theme"></span><button class="demo-toggle" type="button">[pause]</button></div>`;
  const demo = { element, elapsed: motionPreference.matches ? 3600 : 0, paused: motionPreference.matches, themeIndex: 0 };
  const button = element.querySelector('.demo-toggle');
  button.textContent = demo.paused ? '[play]' : '[pause]';
  button.addEventListener('click', () => {
    demo.paused = !demo.paused;
    button.textContent = demo.paused ? '[play]' : '[pause]';
  });
  applyTheme(element, themes[0]);
  return demo;
});

function renderDemo(demo) {
  const index = demo.elapsed >= 10000
    ? demoFrames.length - 1
    : Math.min(demoFrames.length - 2, Math.floor(demo.elapsed / 10000 * (demoFrames.length - 1)));
  if (demo.frameIndex === index) return;
  demo.element.querySelector('.terminal-frame').innerHTML = demoFrames[index];
  demo.frameIndex = index;
}

for (const demo of demos) renderDemo(demo);

let previousTick = performance.now();
setInterval(() => {
  const now = performance.now();
  const delta = Math.min(now - previousTick, 200);
  previousTick = now;
  for (const demo of demos) {
    if (demo.paused || document.hidden || demo.element.closest('section').hidden) continue;
    demo.elapsed += delta * (demo.elapsed < 10000 ? 2 : 1);
    if (demo.elapsed >= 12500) {
      demo.elapsed = 0;
      if (demo.element.dataset.demo === 'auto') {
        demo.themeIndex = (demo.themeIndex + 1) % themes.length;
        applyTheme(demo.element, themes[demo.themeIndex]);
      }
    }
    renderDemo(demo);
  }
}, 80);

motionPreference.addEventListener('change', event => {
  if (!event.matches) return;
  for (const demo of demos) {
    demo.paused = true;
    demo.element.querySelector('.demo-toggle').textContent = '[play]';
  }
});

for (const theme of themes) {
  const button = document.createElement('button');
  const label = document.createElement('span');
  label.textContent = theme.name;
  const swatches = document.createElement('span');
  swatches.className = 'swatches';
  swatches.setAttribute('aria-hidden', 'true');
  for (const color of [theme.bg, theme.accent, theme.fg, theme.error]) {
    const swatch = document.createElement('i');
    swatch.style.backgroundColor = color;
    swatches.append(swatch);
  }
  button.append(label, swatches);
  button.setAttribute('aria-pressed', String(theme.name === 'serika'));
  button.addEventListener('click', () => {
    document.querySelectorAll('#theme-options button').forEach(option => option.setAttribute('aria-pressed', String(option === button)));
    applyTheme(document.querySelector('[data-demo="selected"]'), theme);
    const command = `shunpo -theme ${theme.name}`;
    document.querySelector('#theme-command').textContent = command;
    document.querySelector('#copy-theme').dataset.copy = command;
  });
  document.querySelector('#theme-options').append(button);
}
