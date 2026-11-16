import { useEffect, useRef, useState } from 'react'
import { PaneView } from './components/PaneView'
import { SplitHint, type SplitHintRect } from './components/SplitHint'
import { TabBar } from './components/TabBar'
import { closePane, resizeSplitAt, splitPane } from './paneTree'
import { ReadmeSection } from './sections/ReadmeSection'
import { TerminalSection } from './sections/TerminalSection'
import { UsageSection } from './sections/UsageSection'
import { SECTION_IDS, type PaneNode, type SectionId, type Side } from './types'
import { useHashPage } from './useHashPage'

export default function App() {
  const [page, navigateHash] = useHashPage()
  const [tabs, setTabs] = useState<SectionId[]>(SECTION_IDS)
  const [paneTree, setPaneTree] = useState<PaneNode | null>(null)
  const [draggedId, setDraggedId] = useState<SectionId | null>(null)
  const [splitHintRect, setSplitHintRect] = useState<SplitHintRect | null>(null)
  const mainRef = useRef<HTMLElement>(null)
  // dragover/drop can fire faster than a setState driven by the same drag
  // gesture commits, so the values they need to read synchronously (not
  // just for the next render) live in a ref instead of state.
  const dragRef = useRef<{ id: SectionId | null; hoverPaneId: number | null; side: Side }>({
    id: null,
    hoverPaneId: null,
    side: 'left',
  })

  const beginDrag = (id: SectionId) => {
    dragRef.current.id = id
    setDraggedId(id)
  }

  useEffect(() => {
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && paneTree) setPaneTree(null)
    }
    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [paneTree])

  const navigate = (id: SectionId) => {
    setPaneTree(null)
    navigateHash(id)
  }

  const closeTab = (id: SectionId) => {
    setTabs((prev) => prev.filter((t) => t !== id))
    setPaneTree(null)
  }

  const handleClosePane = (paneId: number) => {
    if (!paneTree) return
    const result = closePane(paneTree, paneId)
    if (result.exited) {
      setPaneTree(null)
      navigate(result.remainingSectionId)
    } else {
      setPaneTree(result.tree)
    }
  }

  const handleResize = (path: number[], sizes: [number, number]) => {
    setPaneTree((tree) => (tree ? resizeSplitAt(tree, path, sizes) : tree))
  }

  const handleMainDragOver = (e: React.DragEvent<HTMLElement>) => {
    if (!dragRef.current.id) return
    e.preventDefault()
    const paneEl = (e.target as HTMLElement).closest<HTMLElement>('[data-pane-id]')
    const paneId = paneEl ? Number(paneEl.dataset.paneId) : null
    dragRef.current.hoverPaneId = paneId

    const rect = (paneEl ?? mainRef.current!).getBoundingClientRect()
    const fx = (e.clientX - rect.left) / rect.width
    const fy = (e.clientY - rect.top) / rect.height
    const edge = Math.min(fx, 1 - fx, fy, 1 - fy)
    const side: Side = edge === fx ? 'left' : edge === 1 - fx ? 'right' : edge === fy ? 'top' : 'bottom'
    dragRef.current.side = side

    let hx = rect.left
    let hy = rect.top
    let hw = rect.width
    let hh = rect.height
    if (side === 'left') hw /= 2
    else if (side === 'right') {
      hx += rect.width / 2
      hw /= 2
    } else if (side === 'top') hh /= 2
    else {
      hy += rect.height / 2
      hh /= 2
    }
    setSplitHintRect({ left: hx, top: hy, width: hw, height: hh })
  }

  const handleMainDragLeave = (e: React.DragEvent<HTMLElement>) => {
    if (e.target === mainRef.current) setSplitHintRect(null)
  }

  const handleMainDrop = (e: React.DragEvent<HTMLElement>) => {
    const { id, hoverPaneId, side } = dragRef.current
    if (!id) return
    e.preventDefault()
    setSplitHintRect(null)
    if (paneTree && hoverPaneId == null) return
    setPaneTree(splitPane(paneTree, page, hoverPaneId, id, side))
    dragRef.current.id = null
    dragRef.current.hoverPaneId = null
    setDraggedId(null)
  }

  const handleDragEnd = () => {
    dragRef.current.id = null
    dragRef.current.hoverPaneId = null
    setDraggedId(null)
    setSplitHintRect(null)
  }

  return (
    <>
      <TabBar
        tabs={tabs}
        activeTabId={page}
        draggedId={draggedId}
        onNavigate={navigate}
        onReorder={setTabs}
        onClose={closeTab}
        onDragStart={beginDrag}
        onDragEnd={handleDragEnd}
      />
      <SplitHint rect={splitHintRect} />
      <main
        ref={mainRef}
        onDragOver={handleMainDragOver}
        onDragLeave={handleMainDragLeave}
        onDrop={handleMainDrop}
        className={
          paneTree
            ? 'flex min-h-0 flex-1 flex-col p-0'
            : 'flex-1 p-[clamp(48px,8vh,90px)_40px_60px]'
        }
      >
        {paneTree ? (
          <PaneView node={paneTree} path={[]} onClosePane={handleClosePane} onResize={handleResize} />
        ) : page === 'terminal' ? (
          <TerminalSection />
        ) : page === 'usage' ? (
          <UsageSection />
        ) : (
          <ReadmeSection />
        )}
      </main>
      <div id="announcement" className="sr-only" role="status" aria-live="polite" />
    </>
  )
}
