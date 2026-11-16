export interface SplitHintRect {
  left: number
  top: number
  width: number
  height: number
}

export function SplitHint({ rect }: { rect: SplitHintRect | null }) {
  if (!rect) return null
  return (
    <div
      aria-hidden="true"
      className="pointer-events-none fixed z-50 border border-dashed border-accent/40 bg-accent/[0.07] transition-[left,top,width,height] duration-[60ms] ease-linear"
      style={{ left: rect.left, top: rect.top, width: rect.width, height: rect.height }}
    />
  )
}
