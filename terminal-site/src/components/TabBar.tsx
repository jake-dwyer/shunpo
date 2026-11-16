import { Chevron } from './Chevron'
import { SECTION_LABEL, type SectionId } from '../types'

const TAB_DOT: Record<SectionId, string> = {
  terminal: '#e2b714',
  usage: '#4fd1ff',
  readme: '#ff5c8a',
}

interface TabBarProps {
  tabs: SectionId[]
  activeTabId: SectionId | null
  draggedId: SectionId | null
  onNavigate: (id: SectionId) => void
  onReorder: (order: SectionId[]) => void
  onClose: (id: SectionId) => void
  onDragStart: (id: SectionId) => void
  onDragEnd: () => void
}

export function TabBar({ tabs, activeTabId, draggedId, onNavigate, onReorder, onClose, onDragStart, onDragEnd }: TabBarProps) {
  return (
    <header className="flex items-end justify-between border-b border-border bg-surface px-[10px] pt-[10px]">
      <nav
        aria-label="Main navigation"
        id="tabs"
        className="flex items-end gap-[4px] max-[700px]:flex-1 max-[700px]:gap-[2px]"
        onDragOver={(e) => {
          if (!draggedId) return
          e.preventDefault()
          const container = e.currentTarget
          const siblings = [...container.querySelectorAll<HTMLElement>('[data-tab]')].filter(
            (el) => el.dataset.tab !== draggedId,
          )
          const next = siblings.find((el) => e.clientX < el.getBoundingClientRect().left + el.offsetWidth / 2)
          const nextId = (next?.dataset.tab as SectionId | undefined) ?? null
          const withoutDragged = tabs.filter((id) => id !== draggedId)
          const insertAt = nextId ? withoutDragged.indexOf(nextId) : withoutDragged.length
          const reordered = [...withoutDragged.slice(0, insertAt), draggedId, ...withoutDragged.slice(insertAt)]
          if (reordered.join() !== tabs.join()) onReorder(reordered)
        }}
      >
        {tabs.map((id) => {
          const active = id === activeTabId
          return (
            <div
              key={id}
              data-tab={id}
              draggable
              onDragStart={(e) => {
                e.dataTransfer.effectAllowed = 'move'
                onDragStart(id)
              }}
              onDragEnd={onDragEnd}
              className={`group relative flex items-stretch rounded-t-[6px] border border-b-0 border-border bg-tab [cursor:grab] ${
                active ? 'top-[1px] bg-black' : ''
              } ${draggedId === id ? 'cursor-grabbing opacity-45' : ''}`}
            >
              <a
                href={`#${id}`}
                draggable={false}
                aria-current={active ? 'page' : undefined}
                onClick={(e) => {
                  e.preventDefault()
                  onNavigate(id)
                }}
                className={`flex items-center gap-[7px] py-[9px] pr-[8px] pl-[12px] text-[13px] font-semibold text-[#6f7376] no-underline max-[700px]:justify-center max-[700px]:p-[10px_6px] max-[700px]:text-[12px] ${
                  active ? 'text-white' : 'group-hover:text-[#c7cacc]'
                }`}
              >
                <span aria-hidden="true" className="h-[7px] w-[7px] shrink-0 rounded-[2px]" style={{ background: TAB_DOT[id] }} />
                {SECTION_LABEL[id]}
                {id === 'terminal' && (
                  <span aria-hidden="true" className="animate-blink text-white">
                    _
                  </span>
                )}
              </a>
              <button
                type="button"
                aria-label={`Close ${SECTION_LABEL[id]} tab`}
                onClick={() => onClose(id)}
                className="mr-[8px] h-[16px] w-[16px] self-center rounded-[3px] border-0 bg-transparent p-0 text-[13px] leading-none text-[#55585a] hover:bg-border hover:text-white max-[700px]:mr-[4px]"
              >
                ×
              </button>
            </div>
          )
        })}
        <button
          type="button"
          tabIndex={-1}
          aria-hidden="true"
          className="mb-[1px] h-[27px] w-[26px] self-end rounded-t-[6px] border border-transparent bg-transparent text-[15px] text-[#6f7376] [cursor:default] max-[700px]:hidden"
        >
          +
        </button>
      </nav>
      <a
        className="mx-[20px] self-center text-[13px] text-[#909397] no-underline hover:text-white max-[700px]:hidden"
        href="https://github.com/jake-dwyer/shunpo"
        target="_blank"
        rel="noreferrer"
      >
        github
        <Chevron />
      </a>
    </header>
  )
}
