// Icon — власний набір тонких SVG-іконок (без емоджі та сторонніх бібліотек).
const PATHS = {
  dashboard: (
    <>
      <rect x="3" y="3" width="7" height="7" rx="1.5" />
      <rect x="14" y="3" width="7" height="7" rx="1.5" />
      <rect x="3" y="14" width="7" height="7" rx="1.5" />
      <rect x="14" y="14" width="7" height="7" rx="1.5" />
    </>
  ),
  graph: (
    <>
      <circle cx="5" cy="6" r="2.6" />
      <circle cx="19" cy="8" r="2.6" />
      <circle cx="10" cy="19" r="2.6" />
      <line x1="7.4" y1="6.6" x2="16.4" y2="7.6" />
      <line x1="5.9" y1="8.4" x2="9.2" y2="16.6" />
      <line x1="17.3" y1="10" x2="12" y2="17.4" />
    </>
  ),
  table: (
    <>
      <rect x="3" y="4" width="18" height="16" rx="2" />
      <line x1="3" y1="9.5" x2="21" y2="9.5" />
      <line x1="3" y1="15" x2="21" y2="15" />
      <line x1="9.5" y1="9.5" x2="9.5" y2="20" />
    </>
  ),
  call: (
    <path d="M5 4h4l1.6 4.4L8.2 10a12.5 12.5 0 0 0 5.8 5.8l1.6-2.4L20 15v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z" />
  ),
  report: (
    <>
      <path d="M6 2.5h8.5L19.5 7.5V21.5h-13.5z" />
      <path d="M14 2.5v5.5h5.5" />
      <line x1="9" y1="13" x2="16" y2="13" />
      <line x1="9" y1="17" x2="15" y2="17" />
    </>
  ),
  ai: <path d="M12 3.5l1.7 4.8 4.8 1.7-4.8 1.7L12 16.5l-1.7-4.8-4.8-1.7 4.8-1.7z" />,
  settings: (
    <>
      <line x1="4" y1="7" x2="20" y2="7" />
      <circle cx="9" cy="7" r="2.2" />
      <line x1="4" y1="17" x2="20" y2="17" />
      <circle cx="15" cy="17" r="2.2" />
    </>
  ),
  edit: (
    <>
      <path d="M12 20h9" />
      <path d="M16.6 3.6a2.1 2.1 0 0 1 3 3L8 18.5l-4.2 1.2L5 15.5z" />
    </>
  ),
  trash: (
    <>
      <path d="M3.5 6h17" />
      <path d="M8.5 6V4h7v2" />
      <path d="M6 6l1 14.5h10L18 6" />
      <line x1="10" y1="10.5" x2="10" y2="17" />
      <line x1="14" y1="10.5" x2="14" y2="17" />
    </>
  ),
  copy: (
    <>
      <rect x="9" y="9" width="11" height="11" rx="2" />
      <path d="M5 15V5a2 2 0 0 1 2-2h10" />
    </>
  ),
  refresh: (
    <>
      <path d="M21 12a9 9 0 1 1-2.7-6.4" />
      <path d="M21 3.5V9h-5.5" />
    </>
  ),
  kb: (
    <>
      <path d="M4 19V5a2 2 0 0 1 2-2h14v18H6a2 2 0 0 1-2-2z" />
      <path d="M4 19a2 2 0 0 1 2-2h14" />
    </>
  ),
  flame: (
    <path d="M12 2.5c1.8 3.8 6 5.8 6 10.5a6 6 0 0 1-12 0c0-1.9.9-3.4 1.8-5 .6 1.6 1.2 3 2.2 3 1.5 0 .8-4.5 2-8.5z" />
  ),
  bot: (
    <>
      <rect x="4" y="8.5" width="16" height="10.5" rx="2.5" />
      <line x1="12" y1="8.5" x2="12" y2="5" />
      <circle cx="12" cy="4" r="1.1" />
      <circle cx="9" cy="13" r="0.9" />
      <circle cx="15" cy="13" r="0.9" />
      <path d="M9.5 16.2h5" />
    </>
  ),
  chat: (
    <path d="M21 11.5a8.5 8.5 0 0 1-8.5 8.5H4.5l1.9-3.1A8.5 8.5 0 1 1 21 11.5z" />
  ),
  send: (
    <>
      <path d="M21 3L10.6 13.4" />
      <path d="M21 3l-6.8 18-3.6-7.6L3 10z" />
    </>
  ),
  flask: (
    <>
      <path d="M9.5 3h5" />
      <path d="M10.5 3v5.5L4.8 17.8A2.4 2.4 0 0 0 7 21.5h10a2.4 2.4 0 0 0 2.2-3.7L13.5 8.5V3" />
      <line x1="7.5" y1="15" x2="16.5" y2="15" />
    </>
  ),
}

export default function Icon({ name, size = 15 }) {
  return (
    <span className="icon" style={{ width: size, height: size }}>
      <svg
        viewBox="0 0 24 24"
        width={size}
        height={size}
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        {PATHS[name] || null}
      </svg>
    </span>
  )
}
