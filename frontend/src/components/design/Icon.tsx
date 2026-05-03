import type { CSSProperties, ReactNode, SVGProps } from 'react'

export type DesignIconName =
  | 'search'
  | 'pin'
  | 'star'
  | 'heart'
  | 'fire'
  | 'bolt'
  | 'gift'
  | 'check'
  | 'checkc'
  | 'arrow'
  | 'arrl'
  | 'caret'
  | 'user'
  | 'cal'
  | 'clock'
  | 'bell'
  | 'msg'
  | 'card'
  | 'filter'
  | 'grid'
  | 'list'
  | 'map'
  | 'plus'
  | 'minus'
  | 'x'
  | 'more'
  | 'set'
  | 'rocket'
  | 'shield'
  | 'out'
  | 'trash'
  | 'pencil'
  | 'upload'
  | 'globe'
  | 'phone'
  | 'chat'
  | 'car'
  | 'qr'
  | 'trend'
  | 'download'
  | 'info'
  | 'calculator'
  | 'camera'
  | 'x-circle'
  | 'code'
  | 'copy'
  | 'crown'
  | 'wallet'
  | 'alert'
  | 'eye'
  | 'file'
  | 'flag'
  | 'inbox'
  | 'key'
  | 'laptop'
  | 'link'
  | 'lock'
  | 'mail'
  | 'pause'
  | 'percent'
  | 'play'
  | 'help'
  | 'reload'
  | 'robot'
  | 'save'
  | 'send'
  | 'share'
  | 'shop'
  | 'cart'
  | 'smile'
  | 'sort'
  | 'split'
  | 'stop'
  | 'swap'
  | 'tag'
  | 'team'
  | 'trophy'
  | 'unlock'
  | 'user-plus'
  | 'user-minus'
  | 'video'

const ICON_PATHS: Record<DesignIconName, ReactNode> = {
  search: <><circle cx="11" cy="11" r="7" /><path d="m21 21-4.3-4.3" /></>,
  pin: <><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 1 1 16 0z" /><circle cx="12" cy="10" r="3" /></>,
  star: <><path d="m12 2 3 7 7 .8-5.5 4.7 1.7 6.9L12 17.8 5.8 21.4l1.7-6.9L2 9.8 9 9z" /></>,
  heart: <path d="M20.84 4.6a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.07a5.5 5.5 0 0 0-7.78 7.78L12 21.23l8.84-8.85a5.5 5.5 0 0 0 0-7.78z" />,
  fire: <path d="M8.5 14.5C5 16 5 22 12 22s7-6 3.5-7.5C18 12 19 9 17.5 6.5c-2 6-9 4-9 8z" />,
  bolt: <path d="M13 2 3 14h8l-1 8 10-12h-8z" />,
  gift: <><rect x="3" y="8" width="18" height="13" rx="2" /><path d="M3 8V6a3 3 0 0 1 3-3h12a3 3 0 0 1 3 3v2M12 8v13" /></>,
  check: <path d="M20 6 9 17l-5-5" />,
  checkc: <><circle cx="12" cy="12" r="10" /><path d="m9 12 2 2 4-4" /></>,
  arrow: <><path d="M5 12h14M13 5l7 7-7 7" /></>,
  arrl: <><path d="M19 12H5M11 5l-7 7 7 7" /></>,
  caret: <path d="m6 9 6 6 6-6" />,
  user: <><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></>,
  cal: <><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M16 3v4M8 3v4M3 11h18" /></>,
  clock: <><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></>,
  bell: <><path d="M6 8a6 6 0 1 1 12 0c0 7 3 7 3 9H3c0-2 3-2 3-9z" /><path d="M10 21a2 2 0 0 0 4 0" /></>,
  msg: <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />,
  card: <><rect x="2" y="6" width="20" height="13" rx="2" /><path d="M2 11h20" /></>,
  filter: <path d="M3 5h18l-7 9v6l-4-2v-4z" />,
  grid: <><rect x="3" y="3" width="7" height="7" rx="1" /><rect x="14" y="3" width="7" height="7" rx="1" /><rect x="3" y="14" width="7" height="7" rx="1" /><rect x="14" y="14" width="7" height="7" rx="1" /></>,
  list: <><path d="M8 6h13M8 12h13M8 18h13" /><circle cx="3.5" cy="6" r="1" /><circle cx="3.5" cy="12" r="1" /><circle cx="3.5" cy="18" r="1" /></>,
  map: <><path d="m9 4-6 2v14l6-2 6 2 6-2V4l-6 2z" /><path d="M9 4v14M15 6v14" /></>,
  plus: <path d="M12 5v14M5 12h14" />,
  minus: <path d="M5 12h14" />,
  x: <path d="m6 6 12 12M18 6 6 18" />,
  more: <><circle cx="5" cy="12" r="1.4" /><circle cx="12" cy="12" r="1.4" /><circle cx="19" cy="12" r="1.4" /></>,
  set: <><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82L4.21 7.12a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09c0 .67.4 1.27 1 1.51a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82c.24.6.84 1 1.51 1H21a2 2 0 0 1 0 4h-.09c-.67 0-1.27.4-1.51 1z" /></>,
  rocket: <><path d="M5 13c-1 4-2 6-2 6s2-1 6-2" /><path d="M14 6s4 0 6 2-2 6-2 6" /><path d="M9 11s2-7 9-9c0 7-2 9-2 9z" /><circle cx="14" cy="9" r="1.2" /></>,
  shield: <path d="m12 2 8 4v6c0 5-4 9-8 10-4-1-8-5-8-10V6z" />,
  out: <><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" /></>,
  trash: <><path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6" /></>,
  pencil: <><path d="M12 20h9" /><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4z" /></>,
  upload: <><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12" /></>,
  globe: <><circle cx="12" cy="12" r="9" /><path d="M3 12h18M12 3a14 14 0 0 1 0 18 14 14 0 0 1 0-18z" /></>,
  phone: <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.13 4.18 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.96.36 1.9.69 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.91.33 1.85.56 2.81.69A2 2 0 0 1 22 16.92z" />,
  chat: <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8z" />,
  car: <><path d="M3 16h18M5 16V9l2-4h10l2 4v7M7 16v3M17 16v3" /><circle cx="7.5" cy="13.5" r="1.2" /><circle cx="16.5" cy="13.5" r="1.2" /></>,
  qr: <><rect x="3" y="3" width="7" height="7" /><rect x="14" y="3" width="7" height="7" /><rect x="3" y="14" width="7" height="7" /><path d="M14 14h2v2h-2zM18 14h3M14 18h3M18 18v3" /></>,
  trend: <><path d="m3 17 6-6 4 4 8-8" /><path d="M14 7h7v7" /></>,
  download: <><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3" /></>,
  info: <><circle cx="12" cy="12" r="10" /><path d="M12 16v-4M12 8h.01" /></>,
  calculator: <><rect x="4" y="2" width="16" height="20" rx="2" /><path d="M8 6h8M8 10h.01M12 10h.01M16 10h.01M8 14h.01M12 14h.01M16 14h.01M8 18h.01M12 18h.01M16 18h.01" /></>,
  camera: <><path d="M4 7h3l2-3h6l2 3h3a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2z" /><circle cx="12" cy="13" r="4" /></>,
  'x-circle': <><circle cx="12" cy="12" r="10" /><path d="m15 9-6 6M9 9l6 6" /></>,
  code: <><path d="m16 18 6-6-6-6M8 6l-6 6 6 6" /></>,
  copy: <><rect x="8" y="8" width="13" height="13" rx="2" /><path d="M4 16c-1.1 0-2-.9-2-2V5c0-1.1.9-2 2-2h9c1.1 0 2 .9 2 2" /></>,
  crown: <><path d="M3 8l5 4 4-7 4 7 5-4-2 11H5z" /><path d="M5 19h14" /></>,
  wallet: <><path d="M3 7a3 3 0 0 1 3-3h13v16H6a3 3 0 0 1-3-3z" /><path d="M3 8h18M16 13h2" /></>,
  alert: <><path d="m12 3 10 18H2z" /><path d="M12 9v5M12 17h.01" /></>,
  eye: <><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z" /><circle cx="12" cy="12" r="3" /></>,
  file: <><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><path d="M14 2v6h6" /></>,
  flag: <><path d="M4 22V4s2-1 5-1 5 2 8 1 3-1 3-1v12s-2 1-5 1-5-2-8-1-3 1-3 1" /></>,
  inbox: <><path d="M22 12h-6l-2 3h-4l-2-3H2" /><path d="m5 4-3 8v6a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2v-6l-3-8z" /></>,
  key: <><circle cx="7.5" cy="15.5" r="5.5" /><path d="m12 11 8-8M17 3l4 4M15 5l4 4" /></>,
  laptop: <><rect x="4" y="5" width="16" height="11" rx="2" /><path d="M2 20h20" /></>,
  link: <><path d="M10 13a5 5 0 0 0 7.1 0l2-2a5 5 0 1 0-7.1-7.1l-1.1 1.1" /><path d="M14 11a5 5 0 0 0-7.1 0l-2 2A5 5 0 0 0 12 20.1l1.1-1.1" /></>,
  lock: <><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></>,
  mail: <><rect x="3" y="5" width="18" height="14" rx="2" /><path d="m3 7 9 6 9-6" /></>,
  pause: <><circle cx="12" cy="12" r="10" /><path d="M10 8v8M14 8v8" /></>,
  percent: <><path d="m19 5-14 14" /><circle cx="7" cy="7" r="2" /><circle cx="17" cy="17" r="2" /></>,
  play: <><circle cx="12" cy="12" r="10" /><path d="m10 8 6 4-6 4z" /></>,
  help: <><circle cx="12" cy="12" r="10" /><path d="M9.5 9a2.8 2.8 0 0 1 5 1.8c0 2-2.5 2.2-2.5 4.2M12 18h.01" /></>,
  reload: <><path d="M21 12a9 9 0 1 1-2.64-6.36" /><path d="M21 3v6h-6" /></>,
  robot: <><rect x="5" y="8" width="14" height="11" rx="2" /><path d="M12 8V4M9 14h.01M15 14h.01M8 19v2M16 19v2" /></>,
  save: <><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" /><path d="M17 21v-8H7v8M7 3v5h8" /></>,
  send: <><path d="m22 2-7 20-4-9-9-4z" /><path d="M22 2 11 13" /></>,
  share: <><circle cx="18" cy="5" r="3" /><circle cx="6" cy="12" r="3" /><circle cx="18" cy="19" r="3" /><path d="m8.6 13.5 6.8 4M15.4 6.5l-6.8 4" /></>,
  shop: <><path d="M4 10h16l-1-6H5z" /><path d="M6 10v10h12V10M9 20v-6h6v6" /></>,
  cart: <><circle cx="9" cy="21" r="1" /><circle cx="20" cy="21" r="1" /><path d="M1 1h4l2.7 13.4a2 2 0 0 0 2 1.6H19a2 2 0 0 0 2-1.6L23 6H6" /></>,
  smile: <><circle cx="12" cy="12" r="10" /><path d="M8 14s1.5 2 4 2 4-2 4-2M9 9h.01M15 9h.01" /></>,
  sort: <><path d="M11 5h10M11 12h7M11 19h4M4 7l3-3 3 3M7 4v16" /></>,
  split: <><rect x="3" y="4" width="7" height="16" rx="1" /><rect x="14" y="4" width="7" height="16" rx="1" /></>,
  stop: <><circle cx="12" cy="12" r="10" /><path d="M9 9h6v6H9z" /></>,
  swap: <><path d="M7 7h14l-4-4M17 17H3l4 4" /></>,
  tag: <><path d="M20 12 12 20 4 12V4h8z" /><path d="M8 8h.01" /></>,
  team: <><path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M22 21v-2a4 4 0 0 0-3-3.87M16 3.13a4 4 0 0 1 0 7.75" /></>,
  trophy: <><path d="M8 21h8M12 17v4" /><path d="M7 4h10v5a5 5 0 0 1-10 0z" /><path d="M5 6H3v1a4 4 0 0 0 4 4M19 6h2v1a4 4 0 0 1-4 4" /></>,
  unlock: <><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 7.8-1.2" /></>,
  'user-plus': <><circle cx="9" cy="8" r="4" /><path d="M2 21a7 7 0 0 1 14 0M19 8v6M16 11h6" /></>,
  'user-minus': <><circle cx="9" cy="8" r="4" /><path d="M2 21a7 7 0 0 1 14 0M16 11h6" /></>,
  video: <><rect x="3" y="6" width="13" height="12" rx="2" /><path d="m16 10 5-3v10l-5-3z" /></>,
}

interface DesignIconProps extends Omit<SVGProps<SVGSVGElement>, 'name'> {
  name: DesignIconName
  size?: number
  style?: CSSProperties
  filled?: boolean
}

export default function DesignIcon({ name, size = 18, className = '', style, filled = false, ...props }: DesignIconProps) {
  const isDecorative = props['aria-label'] == null && props.role == null

  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      fill={filled ? 'currentColor' : 'none'}
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={`rh-icon ${className}`.trim()}
      style={{ flexShrink: 0, ...style }}
      aria-hidden={isDecorative ? 'true' : undefined}
      {...props}
    >
      {ICON_PATHS[name]}
    </svg>
  )
}
