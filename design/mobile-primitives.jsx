/* global React */
// Mobile primitives shared by client + owner screens.
// Mounted on window.RHM (RelaxHUB Mobile) so they don't collide with web RH.

// Reuse exact icon paths from web kit but inline minimal needed
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
  map:    <><path d="m9 4-6 2v14l6-2 6 2 6-2V4l-6 2z"/><path d="M9 4v14M15 6v14"/></>,
  plus:   <><path d="M12 5v14M5 12h14"/></>,
  x:      <><path d="m6 6 12 12M18 6 6 18"/></>,
  set:    <><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.6 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82L4.21 7.12a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.6a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09c0 .67.4 1.27 1 1.51a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82c.24.6.84 1 1.51 1H21a2 2 0 0 1 0 4h-.09c-.67 0-1.27.4-1.51 1z"/></>,
  rocket: <><path d="M5 13c-1 4-2 6-2 6s2-1 6-2"/><path d="M14 6s4 0 6 2-2 6-2 6"/><path d="M9 11s2-7 9-9c0 7-2 9-2 9z"/><circle cx="14" cy="9" r="1.2"/></>,
  shield: <><path d="m12 2 8 4v6c0 5-4 9-8 10-4-1-8-5-8-10V6z"/></>,
  trend:  <><path d="m3 17 6-6 4 4 8-8"/><path d="M14 7h7v7"/></>,
  qr:     <><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><path d="M14 14h2v2h-2zM18 14h3M14 18h3M18 18v3"/></>,
  home:   <><path d="m3 12 9-9 9 9"/><path d="M5 10v10h14V10"/></>,
  flame:  <><path d="M12 2c1 4 5 5 5 11a5 5 0 1 1-10 0c0-3 2-4 2-7 2 1 3 1 3 4 1-2 0-5 0-8z"/></>,
};

function MIcon({ name, size = 20, color = 'currentColor', style = {} }) {
  const p = ICON_PATHS[name];
  if (!p) return null;
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color}
         strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" style={{ flexShrink: 0, ...style }}>
      {p}
    </svg>
  );
}

// Page background — same warm bege/dot vocabulary as web, scaled for mobile
const M_PAGE_BG = `
  radial-gradient(circle at 10% 8%, rgba(15,118,110,0.10), transparent 36%),
  radial-gradient(circle at 90% 14%, rgba(217,119,6,0.10), transparent 32%),
  radial-gradient(circle at 50% 110%, rgba(37,99,235,0.06), transparent 40%),
  #f8f4ec
`;
const M_DOT = `radial-gradient(circle at 1px 1px, rgba(15,23,42,0.05) 1px, transparent 0)`;

function MPage({ children, style = {} }) {
  return (
    <div style={{
      minHeight: '100%', width: '100%', position: 'relative',
      background: M_PAGE_BG,
      paddingBottom: 80,
      ...style,
    }}>
      <div style={{
        position: 'absolute', inset: 0,
        backgroundImage: M_DOT, backgroundSize: '20px 20px',
        WebkitMaskImage: 'linear-gradient(180deg, rgba(0,0,0,0.9), rgba(0,0,0,0))',
        maskImage: 'linear-gradient(180deg, rgba(0,0,0,0.9), rgba(0,0,0,0))',
        pointerEvents: 'none',
      }} />
      <div style={{ position: 'relative' }}>{children}</div>
    </div>
  );
}

// ── Glass card ────────────────────────────────────────────────
function MCard({ children, padding = 16, radius = 22, style = {} }) {
  return (
    <div style={{
      padding, borderRadius: radius,
      background: 'linear-gradient(180deg, rgba(255,255,255,0.94), rgba(255,252,246,0.86))',
      backdropFilter: 'blur(14px)', WebkitBackdropFilter: 'blur(14px)',
      border: '1px solid rgba(15,23,42,0.10)',
      boxShadow: '0 14px 28px rgba(15,23,42,0.06)',
      ...style,
    }}>{children}</div>
  );
}

// ── Header (large title) ─────────────────────────────────────
function MHeader({ eyebrow, title, right, sub, style = {} }) {
  return (
    <div style={{ padding: '10px 20px 14px', position: 'relative', zIndex: 2, ...style }}>
      <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 12 }}>
        <div>
          {eyebrow && <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.12em', textTransform: 'uppercase', color: '#0a5f59' }}>{eyebrow}</span>}
          <h1 style={{ margin: '6px 0 4px', fontSize: 28, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.06 }}>{title}</h1>
          {sub && <p style={{ margin: 0, fontSize: 13, color: '#5f6877', lineHeight: 1.45 }}>{sub}</p>}
        </div>
        {right}
      </div>
    </div>
  );
}

// ── Pill / tag ────────────────────────────────────────────────
function MTag({ children, tone = 'default', style = {} }) {
  const map = {
    default: { bg: 'rgba(255,255,255,0.78)', color: '#16212b', bd: 'rgba(15,23,42,0.10)' },
    primary: { bg: 'rgba(15,118,110,0.10)', color: '#0a5f59', bd: 'rgba(15,118,110,0.22)' },
    gold:    { bg: 'rgba(217,119,6,0.12)', color: '#92400e', bd: 'rgba(217,119,6,0.28)' },
    green:   { bg: 'rgba(21,128,61,0.10)', color: '#15803d', bd: 'rgba(21,128,61,0.24)' },
    red:     { bg: 'rgba(180,35,24,0.08)', color: '#991b1b', bd: 'rgba(180,35,24,0.22)' },
    dark:    { bg: 'rgba(22,33,43,0.94)', color: '#fffdf8', bd: 'transparent' },
  };
  const t = map[tone] || map.default;
  return (
    <span style={{
      display: 'inline-flex', alignItems: 'center', gap: 4,
      padding: '4px 10px', borderRadius: 999,
      fontSize: 11, fontWeight: 700,
      background: t.bg, color: t.color,
      border: `1px solid ${t.bd}`,
      whiteSpace: 'nowrap',
      ...style,
    }}>{children}</span>
  );
}

// ── Buttons ───────────────────────────────────────────────────
function MButton({ children, variant = 'primary', size = 'md', wide, style = {}, ...rest }) {
  const variants = {
    primary: { bg: 'linear-gradient(135deg,#0f766e,#0a5f59)', color: '#fffdf8', shadow: '0 10px 22px rgba(15,118,110,0.24)' },
    default: { bg: 'rgba(255,255,255,0.94)', color: '#16212b', shadow: '0 6px 14px rgba(15,23,42,0.05)' },
    dark:    { bg: '#16212b', color: '#fffdf8', shadow: '0 6px 14px rgba(15,23,42,0.18)' },
    ghost:   { bg: 'transparent', color: '#16212b', shadow: 'none' },
  };
  const v = variants[variant];
  const sizes = { sm: { h: 36, px: 14, fs: 13 }, md: { h: 46, px: 18, fs: 14 }, lg: { h: 54, px: 22, fs: 15 } };
  const s = sizes[size];
  return (
    <button {...rest} style={{
      display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 8,
      height: s.h, padding: `0 ${s.px}px`,
      width: wide ? '100%' : 'auto',
      borderRadius: 999, border: variant === 'default' ? '1px solid rgba(15,23,42,0.12)' : 'none',
      background: v.bg, color: v.color, boxShadow: v.shadow,
      fontFamily: 'inherit', fontSize: s.fs, fontWeight: 700, cursor: 'pointer',
      ...style,
    }}>{children}</button>
  );
}

// ── Photo placeholder ─────────────────────────────────────────
function MPhoto({ height = 160, radius = 18, variant = 'banya', label = '', style = {} }) {
  const variants = {
    banya:  'linear-gradient(135deg, #10313a, #38606a 50%, #9a5c30)',
    forest: 'linear-gradient(135deg, #0f3a3a, #2a6a5e 60%, #5b8a6d)',
    city:   'linear-gradient(135deg, #1f2a3a, #3a4a5a 60%, #6a7a8a)',
    spa:    'linear-gradient(135deg, #2a1f3a, #5e3a6a 50%, #b06a8a)',
    sand:   'linear-gradient(135deg, #b08a4d, #d9b577 60%, #e7d8b9)',
  };
  return (
    <div style={{
      height, borderRadius: radius,
      background: variants[variant] || variants.banya,
      position: 'relative', overflow: 'hidden',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      color: 'rgba(255,255,255,0.62)', fontSize: 10, fontWeight: 700, letterSpacing: '0.18em',
      ...style,
    }}>
      <svg width="100%" height="100%" style={{ position: 'absolute', inset: 0, opacity: 0.16 }}>
        <circle cx="22%" cy="28%" r="40" stroke="#fff" strokeWidth="1" fill="none"/>
        <circle cx="78%" cy="76%" r="60" stroke="#fff" strokeWidth="1" fill="none"/>
      </svg>
      {label ? <span style={{ position: 'relative' }}>{label}</span> : null}
    </div>
  );
}

// ── Avatar ────────────────────────────────────────────────────
function MAvatar({ name = 'А А', size = 36, tone = 'teal' }) {
  const initials = name.split(' ').map(p => p[0]).slice(0, 2).join('');
  const palettes = {
    teal:  ['#0f766e', '#0a5f59'], amber: ['#d97706', '#b45309'],
    blue:  ['#2563eb', '#1d4ed8'], plum:  ['#7c3aed', '#5b21b6'],
  };
  const [a, b] = palettes[tone] || palettes.teal;
  return (
    <div style={{
      width: size, height: size, borderRadius: 999,
      background: `linear-gradient(135deg, ${a}, ${b})`, color: '#fffdf8',
      display: 'flex', alignItems: 'center', justifyContent: 'center',
      fontSize: Math.round(size * 0.34), fontWeight: 700, flexShrink: 0,
    }}>{initials}</div>
  );
}

// ── Tab bar (bottom nav) ─────────────────────────────────────
function MTabBar({ items, active }) {
  return (
    <div style={{
      position: 'absolute', left: 12, right: 12, bottom: 14, height: 64,
      display: 'flex', alignItems: 'center', justifyContent: 'space-around',
      borderRadius: 999,
      background: 'rgba(255, 252, 247, 0.88)',
      backdropFilter: 'blur(20px)', WebkitBackdropFilter: 'blur(20px)',
      border: '1px solid rgba(15,23,42,0.08)',
      boxShadow: '0 14px 36px rgba(15,23,42,0.14)',
      zIndex: 30,
    }}>
      {items.map(([k, label, icon]) => {
        const on = k === active;
        return (
          <div key={k} style={{
            display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2,
            padding: '6px 10px', borderRadius: 999,
            background: on ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'transparent',
            color: on ? '#fffdf8' : '#5f6877',
            boxShadow: on ? '0 6px 14px rgba(15,118,110,0.24)' : 'none',
          }}>
            <MIcon name={icon} size={18} />
            <span style={{ fontSize: 10, fontWeight: 700 }}>{label}</span>
          </div>
        );
      })}
    </div>
  );
}

// ── Stars row ─────────────────────────────────────────────────
function MStars({ value = 4.9, count }) {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 11, fontWeight: 700, color: '#16212b' }}>
      <MIcon name="star" size={12} color="#faad14" style={{ fill: '#faad14' }} />
      <span>{value.toFixed(1)}</span>
      {count != null && <span style={{ color: '#5f6877', fontWeight: 500 }}>({count})</span>}
    </span>
  );
}

window.RHM = window.RHM || {};
Object.assign(window.RHM, {
  MIcon, MPage, MCard, MHeader, MTag, MButton, MPhoto, MAvatar, MTabBar, MStars,
});
