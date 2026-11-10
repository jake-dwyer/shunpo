const sections = [...document.querySelectorAll('main section')];
const mainEl = document.querySelector('main');
const SECTION_LABEL = { terminal: 'shunpo', usage: 'usage', readme: 'readme' };

// paneTree is null in normal single-section mode. Once a tab is dropped
// into the content area it becomes a binary tree - { type:'split', dir,
// sizes:[a,b], children:[node,node] } with 'leaf' nodes at the ends
// holding a sectionId - so dropping onto an already-split pane nests
// another split inside it (that's how you get thirds, quarters, ...),
// same as a real terminal multiplexer.
let paneTree = null;
let nextPaneId = 1;
const leaf = sectionId => ({ type: 'leaf', paneId: nextPaneId++, sectionId });

function navigate() {
  paneTree = null;
  mainEl.classList.remove('panes');
  for (const child of [...mainEl.children]) {
    if (child.tagName !== 'SECTION') child.remove();
  }
  const requested = location.hash.slice(1) || 'terminal';
  const page = sections.some(section => section.id === requested) ? requested : 'terminal';
  for (const section of sections) {
    if (section.parentElement !== mainEl) mainEl.appendChild(section);
    section.hidden = section.id !== page;
  }
  for (const tab of document.querySelectorAll('#tabs .tab')) {
    const link = tab.querySelector('a');
    const active = link.hash === `#${page}`;
    tab.classList.toggle('active', active);
    if (active) link.setAttribute('aria-current', 'page');
    else link.removeAttribute('aria-current');
  }
}

window.addEventListener('hashchange', navigate);
navigate();

window.addEventListener('keydown', event => {
  if (event.key === 'Escape' && paneTree) navigate();
});

document.querySelectorAll('.tab-close').forEach(button => {
  button.addEventListener('click', () => {
    button.closest('.tab').remove();
    navigate();
  });
});

// Drag a tab to reorder it within the bar, or drop it into (or onto an
// existing pane of) the content area to split the view - snap preview
// follows the cursor, like a window manager's split-screen hint.
const tabsNav = document.querySelector('#tabs');
const splitHint = document.querySelector('#split-hint');
let dragTab = null;
let hoverPaneId = null;

function findLeaf(node, paneId) {
  if (node.type === 'leaf') return node.paneId === paneId ? node : null;
  for (const child of node.children) {
    const found = findLeaf(child, paneId);
    if (found) return found;
  }
  return null;
}

function removeSection(node, sectionId) {
  if (node.type === 'leaf') return node;
  for (let i = 0; i < 2; i++) {
    const child = node.children[i];
    if (child.type === 'leaf' && child.sectionId === sectionId) return node.children[1 - i];
  }
  node.children = node.children.map(child => removeSection(child, sectionId));
  return node;
}

function splitPane(targetPaneId, droppedId, side) {
  const dir = side === 'left' || side === 'right' ? 'row' : 'col';
  const toFront = side === 'left' || side === 'top';

  if (!paneTree) {
    const activePage = sections.some(s => s.id === location.hash.slice(1)) ? location.hash.slice(1) : 'terminal';
    if (droppedId === activePage) return;
    const a = leaf(activePage), b = leaf(droppedId);
    paneTree = { type: 'split', dir, sizes: [0.5, 0.5], children: toFront ? [b, a] : [a, b] };
    mainEl.classList.add('panes');
    renderPaneTree();
    return;
  }

  const target = findLeaf(paneTree, targetPaneId);
  if (!target || target.sectionId === droppedId) return;

  // A section can only live in one pane at a time - evict it from
  // wherever else it is before (re)placing it next to the target.
  paneTree = removeSection(paneTree, droppedId);
  if (paneTree.type === 'leaf') {
    paneTree = { type: 'split', dir, sizes: [0.5, 0.5], children: toFront ? [leaf(droppedId), paneTree] : [paneTree, leaf(droppedId)] };
    renderPaneTree();
    return;
  }

  function recur(node) {
    if (node.type === 'leaf') {
      if (node.paneId !== targetPaneId) return node;
      const kids = toFront ? [leaf(droppedId), node] : [node, leaf(droppedId)];
      return { type: 'split', dir, sizes: [0.5, 0.5], children: kids };
    }
    node.children = node.children.map(recur);
    return node;
  }
  paneTree = recur(paneTree);
  renderPaneTree();
}

function closePane(paneId) {
  if (paneTree.type === 'leaf') { navigate(); return; }
  function recur(node) {
    if (node.type !== 'split') return node;
    for (let i = 0; i < 2; i++) {
      const child = node.children[i];
      if (child.type === 'leaf' && child.paneId === paneId) return node.children[1 - i];
    }
    node.children = node.children.map(recur);
    return node;
  }
  paneTree = recur(paneTree);
  if (paneTree.type === 'leaf') {
    const sectionId = paneTree.sectionId;
    paneTree = null;
    mainEl.classList.remove('panes');
    if (location.hash.slice(1) === sectionId) navigate();
    else location.hash = sectionId;
    return;
  }
  renderPaneTree();
}

function startResize(node, splitEl) {
  return event => {
    event.preventDefault();
    const rect = splitEl.getBoundingClientRect();
    const horizontal = node.dir === 'row';
    const total = horizontal ? rect.width : rect.height;
    const start = horizontal ? event.clientX : event.clientY;
    const startA = node.sizes[0];
    const divider = event.currentTarget;
    divider.classList.add('resizing');
    function onMove(e) {
      const pos = horizontal ? e.clientX : e.clientY;
      const a = Math.min(0.85, Math.max(0.15, startA + (pos - start) / total));
      node.sizes = [a, 1 - a];
      const [childA, , childB] = splitEl.children;
      childA.style.flexGrow = String(a);
      childB.style.flexGrow = String(1 - a);
    }
    function onUp() {
      divider.classList.remove('resizing');
      window.removeEventListener('mousemove', onMove);
      window.removeEventListener('mouseup', onUp);
    }
    window.addEventListener('mousemove', onMove);
    window.addEventListener('mouseup', onUp);
  };
}

function buildPane(node) {
  if (node.type === 'leaf') {
    const wrap = document.createElement('div');
    wrap.className = 'pane';
    wrap.dataset.paneId = node.paneId;
    const bar = document.createElement('div');
    bar.className = 'pane-bar';
    const label = document.createElement('span');
    label.textContent = SECTION_LABEL[node.sectionId] || node.sectionId;
    const close = document.createElement('button');
    close.className = 'pane-close';
    close.setAttribute('aria-label', `Close ${label.textContent} pane`);
    close.textContent = '×';
    close.addEventListener('click', () => closePane(node.paneId));
    bar.append(label, close);
    const body = document.createElement('div');
    body.className = 'pane-body';
    const section = sections.find(s => s.id === node.sectionId);
    section.hidden = false;
    body.appendChild(section);
    wrap.append(bar, body);
    return wrap;
  }
  const splitEl = document.createElement('div');
  splitEl.className = `pane-split pane-split-${node.dir}`;
  const childA = buildPane(node.children[0]);
  childA.style.flexGrow = String(node.sizes[0]);
  const divider = document.createElement('div');
  divider.className = 'pane-divider';
  divider.addEventListener('mousedown', startResize(node, splitEl));
  const childB = buildPane(node.children[1]);
  childB.style.flexGrow = String(node.sizes[1]);
  splitEl.append(childA, divider, childB);
  return splitEl;
}

function renderPaneTree() {
  mainEl.innerHTML = '';
  mainEl.appendChild(buildPane(paneTree));
}

tabsNav.addEventListener('dragstart', event => {
  dragTab = event.target.closest('.tab');
  if (!dragTab) return;
  dragTab.classList.add('dragging');
  event.dataTransfer.effectAllowed = 'move';
});
tabsNav.addEventListener('dragend', () => {
  if (dragTab) dragTab.classList.remove('dragging');
  dragTab = null;
  splitHint.hidden = true;
});
tabsNav.addEventListener('dragover', event => {
  if (!dragTab) return;
  event.preventDefault();
  const siblings = [...tabsNav.querySelectorAll('.tab:not(.dragging)')];
  const next = siblings.find(tab => event.clientX < tab.getBoundingClientRect().left + tab.offsetWidth / 2);
  tabsNav.insertBefore(dragTab, next || document.querySelector('.tab-add'));
});

mainEl.addEventListener('dragover', event => {
  if (!dragTab) return;
  event.preventDefault();
  const paneEl = event.target.closest('.pane');
  hoverPaneId = paneEl ? paneEl.dataset.paneId : null;
  const rect = paneEl ? paneEl.getBoundingClientRect() : mainEl.getBoundingClientRect();
  const fx = (event.clientX - rect.left) / rect.width;
  const fy = (event.clientY - rect.top) / rect.height;
  const edge = Math.min(fx, 1 - fx, fy, 1 - fy);
  const side = edge === fx ? 'left' : edge === 1 - fx ? 'right' : edge === fy ? 'top' : 'bottom';
  splitHint.dataset.side = side;
  let hx = rect.left, hy = rect.top, hw = rect.width, hh = rect.height;
  if (side === 'left') hw /= 2;
  else if (side === 'right') { hx += rect.width / 2; hw /= 2; }
  else if (side === 'top') hh /= 2;
  else { hy += rect.height / 2; hh /= 2; }
  Object.assign(splitHint.style, { left: hx + 'px', top: hy + 'px', width: hw + 'px', height: hh + 'px' });
  splitHint.hidden = false;
});
mainEl.addEventListener('dragleave', event => {
  if (event.target === mainEl) splitHint.hidden = true;
});
mainEl.addEventListener('drop', event => {
  if (!dragTab) return;
  event.preventDefault();
  splitHint.hidden = true;
  if (paneTree && !hoverPaneId) return;
  const droppedId = dragTab.querySelector('a').hash.slice(1);
  splitPane(hoverPaneId ? Number(hoverPaneId) : null, droppedId, splitHint.dataset.side);
});

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

// Mini typing demo: steps through WORDS one character at a time (with a
// single simulated mistake + correction), then holds a short results line
// before looping. Pauses when off-screen, hidden, or reduced motion is set.
const miniDemo = document.querySelector('.mini-demo');
if (miniDemo) {
  const esc = s => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  const span = (cls, text) => cls ? `<span class="${cls}">${esc(text)}</span>` : esc(text);
  const muted = t => span('muted', t);
  const accent = t => span('accent', t);
  const err = t => span('err', t);

  const WORD_POOL = [
    'necessary', 'fact', 'ran', 'fast', 'sand', 'when', 'draw', 'multiply',
    'family', 'cross', 'lie', 'too', 'position', 'speak', 'quick', 'water',
    'floor', 'hair', 'are', 'rock', 'build', 'agent', 'pane', 'spare',
    'flow', 'finger', 'light', 'string', 'focus', 'branch',
  ];
  const WORD_LEN = 10;
  const motionPreference = window.matchMedia('(prefers-reduced-motion: reduce)');
  const section = miniDemo.closest('section');
  const state = { wordIndex: 0, charIndex: 0, mistake: null, phase: 'typing' };
  let WORDS = pickWords();
  let mistakeSpot = pickMistake();
  let timer = null;

  function pickWords() {
    const pool = [...WORD_POOL];
    for (let i = pool.length - 1; i > 0; i--) {
      const j = Math.floor(Math.random() * (i + 1));
      [pool[i], pool[j]] = [pool[j], pool[i]];
    }
    return pool.slice(0, WORD_LEN);
  }

  function pickMistake() {
    const candidates = WORDS.map((w, i) => ({ i, w })).filter(({ w }) => w.length >= 3);
    const { i, w } = candidates[Math.floor(Math.random() * candidates.length)];
    return { wordIndex: i, charIndex: Math.floor(Math.random() * (w.length - 2)) + 1 };
  }

  function header(seconds) {
    return `${accent('瞬歩 shunpo')}  ${muted('✿')}  ${accent('time 30')}\n\n${accent(String(seconds))}\n\n`;
  }

  function render() {
    if (state.phase === 'result') {
      miniDemo.innerHTML =
        `${accent('瞬歩 shunpo')}\n\n` +
        `${accent('new best!')}  ${muted('(was 96 wpm)')}\n\n` +
        `${muted('wpm')}   ${muted('acc')}   ${muted('raw')}   ${muted('time')}\n` +
        `${accent('104')}   ${accent('97%')}  ${accent('109')}  ${accent('30s')}\n\n` +
        `${muted('tab')} restart   ${muted('esc')} quit`;
      return;
    }
    const seconds = Math.max(30 - state.wordIndex * 3, 4);
    let out = header(seconds);
    for (let i = 0; i < WORD_LEN; i++) {
      const w = WORDS[i];
      if (i < state.wordIndex) {
        out += esc(w) + ' ';
      } else if (i === state.wordIndex) {
        out += esc(w.slice(0, state.charIndex));
        if (state.mistake) out += err(state.mistake);
        out += accent('|') + muted(w.slice(state.charIndex)) + ' ';
      } else {
        out += muted(w) + ' ';
      }
    }
    out += `\n\n${muted('esc')} quit   ${muted('enter')} restart   ${muted('tab')} settings`;
    miniDemo.innerHTML = out;
  }

  function reset() {
    state.wordIndex = 0;
    state.charIndex = 0;
    state.mistake = null;
    state.phase = 'typing';
    WORDS = pickWords();
    mistakeSpot = pickMistake();
  }

  function step() {
    const word = WORDS[state.wordIndex];
    const atMistake = !mistakeSpot.used && state.wordIndex === mistakeSpot.wordIndex && state.charIndex === mistakeSpot.charIndex;
    if (atMistake && !state.mistake) {
      mistakeSpot.used = true;
      state.mistake = word[state.charIndex] === 'e' ? 'r' : 'e';
    } else {
      if (state.mistake) state.mistake = null;
      state.charIndex++;
      if (state.charIndex >= word.length) {
        state.wordIndex++;
        state.charIndex = 0;
        if (state.wordIndex >= WORD_LEN) state.phase = 'result';
      }
    }
    render();
  }

  function schedule() {
    if (motionPreference.matches) return;
    const delay = state.phase === 'result' ? 2600 : (state.mistake ? 260 : 60 + Math.random() * 90);
    timer = setTimeout(() => {
      if (!document.hidden && !section.hidden) {
        if (state.phase === 'result') reset();
        step();
      }
      schedule();
    }, delay);
  }

  render();
  if (motionPreference.matches) {
    state.phase = 'result';
    render();
  } else {
    schedule();
  }

  motionPreference.addEventListener('change', event => {
    if (!event.matches) return;
    clearTimeout(timer);
    state.phase = 'result';
    render();
  });
}
