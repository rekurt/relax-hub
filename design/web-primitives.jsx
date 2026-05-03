/* global React */
// Shared primitives for RelaxHUB web mocks.
// Available as window.RH.* — used by every screen in ui_kits/web/.

const { useState } = React;

// ── Lucide-style outline icons ───────────────────────────────────
// (Stroke-only, 1.75 weight. Match production @ant-design/icons in feel.)
function Icon({ name, size = 18, className = '', style = {} }) {
  const paths = ICON_PATHS[name];
  if (!paths) return null;
  return (
    <svg
      viewBox="0 0 24 24"
      width={size}
      height={size}
      fill="none"
      stroke="currentColor"
      strokeWidth="1.75"
      strokeLinecap="round"
      strokeLinejoin="round"
      className={className}
      style={{ flexShrink: 0, ...style }}
    >
      {paths}
    </svg>
  );
}

const ICON_PATHS = {
  search: <><circle cx="11" cy="11" r="7"/><path d="m21 21-4.3-4.3"/></>,
  pin:    <><path d="M20 10c0 6-8 12-8 12s-8-6-8-12a8 8 0 1 1 16 0z"/><circle cx="12" cy="10" r="3"/></>,
  star:   <><path d="m12 2 3 7 7 .8-5.5 4.7 1.7 6.9L12 17.8 5.8 21.4l1.7-6.9L2 9.8 9 9z"/></>,
  heart:  <><path d="M20.84 4.6a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.07a5.5 5.5 0 0 0-7.78 7.78L12 21.23l8.84-8.85a5.5 5.5 0 0 0 0-7.78z"/></>,
  fire:   <><path d="M8.5 14.5C5 16 5 22 12 22s7-6 3.5-7.5C18 12 19 9 17.5 6.5c-2 6-9 4-9 8z"/></>,
  bolt:   <><path d="M13 2 3 14h8l-1 8 10-12h-8z"/></>,
  gift:   <><rect x="3" y="8" width="18" height="13" rx="2"/><path d="M3 8V6a3 3 0 0 1 3-3h12a3 3 0 0 1 3 3v2M12 8v13"/></>,
  check:  <><path d="M20 6 9 17l-5-5"/></>,
  checkc: <><circle cx="12" cy="12" r="10"/><path d="m9 12 2 2 4-4"/></>,
  arrow:  <><path d="M5 12h14M13 5l7 7-7 7"/></>,
  arrl:   <><path d="M19 12H5M11 5l-7 7 7 7"/></>,
  caret:  <><path d="m6 9 6 6 6-6"/></>,
  user:   <><circle cx="12" cy="8" r="4"/><path d="M4 21a8 8 0 0 1 16 0"/></>,
  cal:    <><rect x="3" y="5" width="18" height="16" rx="2"/><path d="M16 3v4M8 3v4M3 11h18"/></>,
  clock:  <><circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/></>,
  bell:   <><path d="M6 8a6 6 0 1 1 12 0c0 7 3 7 3 9H3c0-2 3-2 3-9z"/><path d="M10 21a2 2 0 0 0 4 0"/></>,
  msg:    <><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></>,
  card:   <><rect x="2" y="6" width="20" height="13" rx="2"/><path d="M2 11h20"/></>,
  filter: <><path d="M3 5h18l-7 9v6l-4-2v-4z"/></>,
  grid:   <><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></>,
  list:   <><path d="M8 6h13M8 12h13M8 18h13"/><circle cx="3.5" cy="6" r="1"/><circle cx="3.5" cy="12" r="1"/><circle cx="3.5" cy="18" r="1"/></>,
  map:    <><path d="m9 4-6 2v14l6-2 6 2 6-2V4l-6 2z"/><path d="M9 4v14M15 6v14"/></>,
  plus:   <><path d="M12 5v14M5 12h14"/></>,
  minus:  <><path d="M5 12h14"/></>,
  x:      <><path d="m6 6 12 12M18 6 6 18"/></>,
  more:   <><circle cx="5" cy="12" r="1.4"/><circle cx="12" cy="12" r="1.4"/><circle cx="19" cy="12" r="1.4"/></>,
  set:    <><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82L4.21 7.12a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09c0 .67.4 1.27 1 1.51a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82c.24.6.84 1 1.51 1H21a2 2 0 0 1 0 4h-.09c-.67 0-1.27.4-1.51 1z"/></>,
  rocket: <><path d="M5 13c-1 4-2 6-2 6s2-1 6-2"/><path d="M14 6s4 0 6 2-2 6-2 6"/><path d="M9 11s2-7 9-9c0 7-2 9-2 9z"/><circle cx="14" cy="9" r="1.2"/></>,
  shield: <><path d="m12 2 8 4v6c0 5-4 9-8 10-4-1-8-5-8-10V6z"/></>,
  out:    <><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/></>,
  trash:  <><path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/></>,
  pencil: <><path d="M12 20h9"/><path d="M16.5 3.5a2.12 2.12 0 0 1 3 3L7 19l-4 1 1-4z"/></>,
  upload: <><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M17 8l-5-5-5 5M12 3v12"/></>,
  globe:  <><circle cx="12" cy="12" r="9"/><path d="M3 12h18M12 3a14 14 0 0 1 0 18 14 14 0 0 1 0-18z"/></>,
  phone:  <><path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6A19.79 19.79 0 0 1 2.13 4.18 2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72c.13.96.36 1.9.69 2.81a2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45c.91.33 1.85.56 2.81.69A2 2 0 0 1 22 16.92z"/></>,
  chat:   <><path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8z"/></>,
  car:    <><path d="M3 16h18M5 16V9l2-4h10l2 4v7M7 16v3M17 16v3"/><circle cx="7.5" cy="13.5" r="1.2"/><circle cx="16.5" cy="13.5" r="1.2"/></>,
  qr:     <><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><path d="M14 14h2v2h-2zM18 14h3M14 18h3M18 18v3"/></>,
  trend:  <><path d="m3 17 6-6 4 4 8-8"/><path d="M14 7h7v7"/></>,
  download: <><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4M7 10l5 5 5-5M12 15V3"/></>,
};

// ── Brand lockup ─────────────────────────────────────────────────
function BrandLockup({ size = 22, dark = false, sub = false }) {
  return (
    <div style={{ display: 'inline-flex', flexDirection: 'column', lineHeight: 1 }}>
      <span style={{
        fontWeight: 800, fontSize: size, letterSpacing: '-0.06em', lineHeight: 0.94,
        color: dark ? '#fffdf8' : '#16212b',
      }}>RelaxHUB</span>
      {sub ? (
        <span style={{
          marginTop: 6, fontSize: 10, fontWeight: 700,
          letterSpacing: '0.14em', textTransform: 'uppercase',
          color: dark ? 'rgba(255,255,255,0.66)' : '#5f6877',
        }}>Booking · Бани и спа</span>
      ) : null}
    </div>
  );
}

// ── Avatar (initials, glass) ─────────────────────────────────────
function Avatar({ name = 'А А', size = 36, tone = 'teal' }) {
  const initials = name.split(' ').map(p => p[0]).slice(0, 2).join('');
  const palettes = {
    teal:  ['#0f766e', '#0a5f59'],
    amber: ['#d97706', '#b45309'],
    blue:  ['#2563eb', '#1d4ed8'],
    plum:  ['#7c3aed', '#5b21b6'],
    sand:  ['#e7d8b9', '#cdb585'],
  };
  const [a, b] = palettes[tone] || palettes.teal;
  return (
    <div style={{
      width: size, height: size, borderRadius: '999px',
      background: `linear-gradient(135deg, ${a}, ${b})`,
      color: '#fffdf8',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      fontSize: Math.round(size * 0.34), fontWeight: 700,
      letterSpacing: '0.02em',
      flexShrink: 0,
    }}>{initials}</div>
  );
}

// ── Generic photo placeholder ────────────────────────────────────
function PhotoPlaceholder({
  width = '100%', height = 200, label = 'ФОТО ОБЪЕКТА',
  variant = 'banya', radius = 20,
}) {
  const variants = {
    banya: 'linear-gradient(135deg, #10313a, #38606a 50%, #9a5c30)',
    forest: 'linear-gradient(135deg, #0f3a3a, #2a6a5e 60%, #5b8a6d)',
    city:   'linear-gradient(135deg, #1f2a3a, #3a4a5a 60%, #6a7a8a)',
    spa:    'linear-gradient(135deg, #2a1f3a, #5e3a6a 50%, #b06a8a)',
    sand:   'linear-gradient(135deg, #b08a4d, #d9b577 60%, #e7d8b9)',
  };
  return (
    <div style={{
      width, height, borderRadius: radius,
      background: variants[variant] || variants.banya,
      position: 'relative', overflow: 'hidden',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      color: 'rgba(255,255,255,0.62)', fontSize: 11, fontWeight: 700,
      letterSpacing: '0.18em',
    }}>
      <svg width="100%" height="100%" style={{ position: 'absolute', inset: 0, opacity: 0.16 }}>
        <circle cx="22%" cy="28%" r="60" stroke="#fff" strokeWidth="1" fill="none"/>
        <circle cx="78%" cy="76%" r="80" stroke="#fff" strokeWidth="1" fill="none"/>
      </svg>
      <span style={{ position: 'relative', zIndex: 1 }}>{label}</span>
    </div>
  );
}

// ── Stars row ────────────────────────────────────────────────────
function Stars({ value = 4.9, count }) {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, color: '#faad14', fontSize: 12, fontWeight: 600 }}>
      <Icon name="star" size={14} style={{ stroke: '#faad14', fill: '#faad14' }} />
      <span>{value.toFixed(1)}</span>
      {count != null ? <span style={{ color: '#5f6877', fontWeight: 500 }}>({count})</span> : null}
    </span>
  );
}

// ── Annotation pill (for kit context) ────────────────────────────
function KitNote({ children, style = {} }) {
  return (
    <div style={{
      position: 'absolute', top: 16, left: 16, zIndex: 10,
      padding: '6px 12px',
      borderRadius: 999,
      background: 'rgba(15, 118, 110, 0.92)',
      color: '#fffdf8',
      fontSize: 11, fontWeight: 700,
      letterSpacing: '0.10em',
      textTransform: 'uppercase',
      boxShadow: '0 8px 18px rgba(15, 118, 110, 0.30)',
      ...style,
    }}>{children}</div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, {
  Icon, BrandLockup, Avatar, PhotoPlaceholder, Stars, KitNote,
});
