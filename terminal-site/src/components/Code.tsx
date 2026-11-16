import type { ReactNode } from 'react'

export function CodeList({ children, className = '' }: { children: ReactNode; className?: string }) {
  return (
    <ol
      className={`m-0 list-none p-0 text-[clamp(14px,1vw,17px)] leading-[1.85] tracking-[-0.3px] [counter-reset:line] max-[700px]:text-[16px] max-[700px]:leading-[1.95] max-[700px]:tracking-[-0.5px] ${className}`}
    >
      {children}
    </ol>
  )
}

export function CodeLine({
  children,
  blank = false,
  className = '',
}: {
  children?: ReactNode
  blank?: boolean
  className?: string
}) {
  return (
    <li
      aria-hidden={blank || undefined}
      className={`relative min-h-[1.85em] pl-[48px] [counter-increment:line] before:absolute before:top-0 before:left-0 before:w-[25px] before:translate-y-0 before:text-right before:text-[0.9em] before:tracking-normal before:text-[#606568] before:select-none before:content-[counter(line)] max-[700px]:min-h-[1.95em] max-[700px]:pl-[38px] max-[700px]:before:w-[16px] ${className}`}
    >
      {blank ? ' ' : children}
    </li>
  )
}

export function Comment({ children }: { children: ReactNode }) {
  return <span className="text-[#85898c]">{children}</span>
}

export function Annotation({ children }: { children: ReactNode }) {
  return (
    <span className="ml-[35px] inline-block text-[0.72em] tracking-normal text-[#85898c] max-[700px]:m-0 max-[700px]:block">
      {children}
    </span>
  )
}

export function Kbd({ children }: { children: ReactNode }) {
  return <kbd className="inline-block min-w-[5ch] font-[inherit]">{children}</kbd>
}

export function BlockCursor() {
  return (
    <span
      aria-hidden="true"
      className="animate-blink ml-[-12px] inline-block h-[1.1em] w-[11px] self-center bg-[#ddd] max-[700px]:ml-0"
    />
  )
}

export function Command({ command }: { command: string }) {
  return (
    <div className="flex items-baseline gap-[22px] max-[700px]:flex-wrap max-[700px]:gap-[13px]">
      <code className="font-[inherit] [overflow-wrap:anywhere]">{command}</code>
      <BlockCursor />
    </div>
  )
}

export function CopyCommand({ command }: { command: string }) {
  const onCopy = async (e: React.MouseEvent<HTMLButtonElement>) => {
    const button = e.currentTarget
    const announcement = document.getElementById('announcement')
    try {
      await navigator.clipboard.writeText(command)
      button.textContent = '[copied]'
      if (announcement) announcement.textContent = 'Command copied to clipboard.'
    } catch {
      button.textContent = '[select to copy]'
      const codeEl = button.previousElementSibling
      if (codeEl) {
        const range = document.createRange()
        range.selectNodeContents(codeEl)
        const selection = window.getSelection()
        selection?.removeAllRanges()
        selection?.addRange(range)
      }
      if (announcement) announcement.textContent = 'Command selected. Press your system copy shortcut.'
    }
    setTimeout(() => {
      button.textContent = '[copy]'
    }, 2200)
  }

  return (
    <div className="flex items-baseline gap-[22px] max-[700px]:flex-wrap max-[700px]:gap-[13px]">
      <code className="[overflow-wrap:anywhere] font-[inherit]">{command}</code>
      <button
        type="button"
        onClick={onCopy}
        aria-label="Copy install command"
        className="shrink-0 cursor-pointer border-0 bg-transparent p-0 text-[12px] tracking-normal text-[#747a7d] hover:text-white max-[700px]:leading-[1.4]"
      >
        [copy]
      </button>
    </div>
  )
}
