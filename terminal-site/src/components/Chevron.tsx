export function Chevron({ className = '' }: { className?: string }) {
  return (
    <svg
      className={`ml-[3px] inline align-[-1px] ${className}`}
      width="9"
      height="9"
      viewBox="0 0 9 9"
      aria-hidden="true"
    >
      <path
        d="M2 1.5H7.5V7M7.5 1.5L1.5 7.5"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.3"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
    </svg>
  )
}
