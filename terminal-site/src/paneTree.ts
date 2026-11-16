import type { PaneLeaf, PaneNode, PaneSplit, SectionId, Side } from './types'

let nextPaneId = 1
export const leaf = (sectionId: SectionId): PaneLeaf => ({
  type: 'leaf',
  paneId: nextPaneId++,
  sectionId,
})

export function sideToDir(side: Side): 'row' | 'col' {
  return side === 'left' || side === 'right' ? 'row' : 'col'
}

export function sideToFront(side: Side): boolean {
  return side === 'left' || side === 'top'
}

function findLeafPath(node: PaneNode, paneId: number, path: number[] = []): number[] | null {
  if (node.type === 'leaf') return node.paneId === paneId ? path : null
  for (let i = 0; i < 2; i++) {
    const found = findLeafPath(node.children[i], paneId, [...path, i])
    if (found) return found
  }
  return null
}

function getLeafAt(node: PaneNode, path: number[]): PaneLeaf {
  let cur = node
  for (const i of path) {
    if (cur.type !== 'split') throw new Error('invalid pane path')
    cur = cur.children[i]
  }
  if (cur.type !== 'leaf') throw new Error('expected leaf at path')
  return cur
}

function replaceAt(node: PaneNode, path: number[], replacement: PaneNode): PaneNode {
  if (path.length === 0) return replacement
  if (node.type !== 'split') return node
  const [head, ...rest] = path
  const children = [...node.children] as [PaneNode, PaneNode]
  children[head] = replaceAt(children[head], rest, replacement)
  return { ...node, children }
}

// A section can only live in one pane at a time - evict it from wherever
// else it is before (re)placing it next to a drop target.
function removeSection(node: PaneNode, sectionId: SectionId): PaneNode {
  if (node.type === 'leaf') return node
  for (let i = 0; i < 2; i++) {
    const child = node.children[i]
    if (child.type === 'leaf' && child.sectionId === sectionId) return node.children[1 - i]
  }
  const children = node.children.map((c) => removeSection(c, sectionId)) as [PaneNode, PaneNode]
  return { ...node, children }
}

// Dropping a tab onto an existing pane nests another split inside it -
// that's how you get thirds, quarters, ... - same as a real terminal
// multiplexer. Dropping onto empty content area (tree === null, only
// possible on the very first split) starts the tree.
export function splitPane(
  tree: PaneNode | null,
  activePage: SectionId,
  targetPaneId: number | null,
  droppedId: SectionId,
  side: Side,
): PaneNode | null {
  const dir = sideToDir(side)
  const toFront = sideToFront(side)

  if (!tree) {
    if (droppedId === activePage) return tree
    const a = leaf(activePage)
    const b = leaf(droppedId)
    const children: [PaneNode, PaneNode] = toFront ? [b, a] : [a, b]
    return { type: 'split', dir, sizes: [0.5, 0.5], children }
  }

  if (targetPaneId == null) return tree
  const targetPath = findLeafPath(tree, targetPaneId)
  if (!targetPath) return tree
  const target = getLeafAt(tree, targetPath)
  if (target.sectionId === droppedId) return tree

  const evicted = removeSection(tree, droppedId)
  if (evicted.type === 'leaf') {
    const children: [PaneNode, PaneNode] = toFront ? [leaf(droppedId), evicted] : [evicted, leaf(droppedId)]
    return { type: 'split', dir, sizes: [0.5, 0.5], children }
  }

  // Eviction may have shortened paths above the target (a sibling split
  // collapsing away), so re-resolve the target's path in the evicted tree.
  const pathAfterEvict = findLeafPath(evicted, targetPaneId)
  if (!pathAfterEvict) return evicted
  const targetNode = getLeafAt(evicted, pathAfterEvict)
  const children: [PaneNode, PaneNode] = toFront ? [leaf(droppedId), targetNode] : [targetNode, leaf(droppedId)]
  return replaceAt(evicted, pathAfterEvict, { type: 'split', dir, sizes: [0.5, 0.5], children })
}

export type ClosePaneResult =
  | { exited: true; remainingSectionId: SectionId }
  | { exited: false; tree: PaneNode }

export function closePane(tree: PaneNode, paneId: number): ClosePaneResult {
  if (tree.type === 'leaf') return { exited: true, remainingSectionId: tree.sectionId }

  function recur(node: PaneNode): PaneNode {
    if (node.type !== 'split') return node
    for (let i = 0; i < 2; i++) {
      const child = node.children[i]
      if (child.type === 'leaf' && child.paneId === paneId) return node.children[1 - i]
    }
    const children = node.children.map(recur) as [PaneNode, PaneNode]
    return { ...node, children }
  }

  const next = recur(tree)
  if (next.type === 'leaf') return { exited: true, remainingSectionId: next.sectionId }
  return { exited: false, tree: next }
}

export function resizeSplitAt(tree: PaneNode, path: number[], sizes: [number, number]): PaneNode {
  if (path.length === 0) {
    if (tree.type !== 'split') return tree
    return { ...tree, sizes }
  }
  if (tree.type !== 'split') return tree
  const [head, ...rest] = path
  const children = [...tree.children] as [PaneNode, PaneNode]
  children[head] = resizeSplitAt(children[head], rest, sizes)
  return { ...tree, children }
}

export function getSplitAt(node: PaneNode, path: number[]): PaneSplit {
  let cur = node
  for (const i of path) {
    if (cur.type !== 'split') throw new Error('invalid pane path')
    cur = cur.children[i]
  }
  if (cur.type !== 'split') throw new Error('expected split at path')
  return cur
}
