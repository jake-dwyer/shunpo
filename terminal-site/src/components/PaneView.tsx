import { useCallback, useRef } from 'react'
import type { PaneNode } from '../types'
import { SECTION_LABEL } from '../types'
import { ReadmeSection } from '../sections/ReadmeSection'
import { TerminalSection } from '../sections/TerminalSection'
import { UsageSection } from '../sections/UsageSection'

interface PaneViewProps {
  node: PaneNode
  path: number[]
  onClosePane: (paneId: number) => void
  onResize: (path: number[], sizes: [number, number]) => void
}

export function PaneView({ node, path, onClosePane, onResize }: PaneViewProps) {
  if (node.type === 'leaf') {
    const section =
      node.sectionId === 'terminal' ? (
        <TerminalSection />
      ) : node.sectionId === 'usage' ? (
        <UsageSection />
      ) : (
        <ReadmeSection />
      )
    return (
      <div data-pane-id={node.paneId} className="flex min-h-0 min-w-0 flex-1 flex-col">
        <div className="flex shrink-0 items-center justify-between border-b border-border bg-surface py-[5px] pr-[5px] pl-[12px] text-[11px] text-[#73787b]">
          <span>{SECTION_LABEL[node.sectionId]}</span>
          <button
            type="button"
            aria-label={`Close ${SECTION_LABEL[node.sectionId]} pane`}
            onClick={() => onClosePane(node.paneId)}
            className="h-[16px] w-[16px] rounded-[3px] border-0 bg-transparent p-0 text-[13px] leading-none text-[#55585a] hover:bg-border hover:text-white"
          >
            ×
          </button>
        </div>
        <div className="min-h-0 flex-1 overflow-auto p-[clamp(24px,4vh,48px)_28px] [&>section]:max-w-none">
          {section}
        </div>
      </div>
    )
  }

  return (
    <SplitView node={node} path={path} onClosePane={onClosePane} onResize={onResize} />
  )
}

function SplitView({ node, path, onClosePane, onResize }: PaneViewProps & { node: Extract<PaneNode, { type: 'split' }> }) {
  const ref = useRef<HTMLDivElement>(null)

  const startResize = useCallback(
    (e: React.MouseEvent) => {
      e.preventDefault()
      const splitEl = ref.current
      if (!splitEl) return
      const rect = splitEl.getBoundingClientRect()
      const horizontal = node.dir === 'row'
      const total = horizontal ? rect.width : rect.height
      const start = horizontal ? e.clientX : e.clientY
      const startA = node.sizes[0]
      const dividerEl = e.currentTarget as HTMLElement
      dividerEl.classList.add('resizing-divider')

      const onMove = (ev: MouseEvent) => {
        const pos = horizontal ? ev.clientX : ev.clientY
        const a = Math.min(0.85, Math.max(0.15, startA + (pos - start) / total))
        onResize(path, [a, 1 - a])
      }
      const onUp = () => {
        dividerEl.classList.remove('resizing-divider')
        window.removeEventListener('mousemove', onMove)
        window.removeEventListener('mouseup', onUp)
      }
      window.addEventListener('mousemove', onMove)
      window.addEventListener('mouseup', onUp)
    },
    [node.dir, node.sizes, onResize, path],
  )

  const [a, b] = node.children
  const horizontal = node.dir === 'row'

  return (
    <div ref={ref} className={`flex min-h-0 min-w-0 flex-1 ${horizontal ? 'flex-row' : 'flex-col'} max-[700px]:flex-col`}>
      <div className="flex min-h-0 min-w-0 overflow-hidden" style={{ flexGrow: node.sizes[0], flexBasis: 0, flexShrink: 1 }}>
        <PaneView node={a} path={[...path, 0]} onClosePane={onClosePane} onResize={onResize} />
      </div>
      <div
        onMouseDown={startResize}
        className={`relative shrink-0 grow-0 basis-[1px] bg-border hover:bg-accent [&.resizing-divider]:bg-accent ${
          horizontal ? 'cursor-col-resize max-[700px]:cursor-row-resize' : 'cursor-row-resize'
        } before:absolute before:content-[''] ${
          horizontal
            ? 'before:inset-y-0 before:-left-[3px] before:-right-[3px]'
            : 'before:inset-x-0 before:-top-[3px] before:-bottom-[3px]'
        }`}
      />
      <div className="flex min-h-0 min-w-0 overflow-hidden" style={{ flexGrow: node.sizes[1], flexBasis: 0, flexShrink: 1 }}>
        <PaneView node={b} path={[...path, 1]} onClosePane={onClosePane} onResize={onResize} />
      </div>
    </div>
  )
}
