/* global React, ReactDOM, DesignCanvas, DCSection, DCArtboard, DCPostIt,
   IOSDevice, IOSStatusBar, IOSNavBar,
   useTweaks, TweaksPanel, TweakSection, TweakRadio, TweakToggle, TweakSelect, TweakButton */

const { useState, useEffect, useRef } = React;
const RH = window.RH;
const RHM = window.RHM;

// ── Helpers ──────────────────────────────────────────────────────

// Web artboard frame: shows a 1280px page at design size.
// We render at native size but cap height; the design canvas shrinks
// the artboard to fit. Inside, content can scroll if it overflows.
function WebFrame({ width = 1280, height = 1700, scrollable = true, children }) {
  return (
    <div className="rh-art-frame" style={{ width, height }}>
      <div
        className="rh-art-scroll"
        style={{
          overflowY: scrollable ? 'auto' : 'hidden',
          overflowX: 'hidden',
        }}
      >
        {children}
      </div>
    </div>
  );
}

// Mobile artboard frame: a single iOS device, centered.
function MobileFrame({ children, dark = false }) {
  return (
    <div
      className="rh-art-frame rh-art-frame--mobile"
      style={{ width: 402, height: 874, padding: 0, background: 'transparent', boxShadow: 'none' }}
    >
      <IOSDevice width={402} height={874} dark={dark}>
        {children}
      </IOSDevice>
    </div>
  );
}

// Section banner — used inside DCSection for richer storytelling.
function SectionBanner({ num, title, sub }) {
  return (
    <div className="rh-banner">
      <div className="rh-banner__num">{num}</div>
      <div>
        <div className="rh-banner__t">{title}</div>
        <div className="rh-banner__s">{sub}</div>
      </div>
    </div>
  );
}

// ── Component showcase row (last section) ───────────────────────
function ShowcaseTile({ title, children }) {
  return (
    <div className="ks-tile">
      <h4>{title}</h4>
      {children}
    </div>
  );
}

function ComponentShowcase() {
  const { Icon, BrandLockup, Avatar, PhotoPlaceholder, Stars } = RH;
  return (
    <div style={{ width: 1280, padding: 28, background: '#fffdf8', borderRadius: 24, border: '1px solid rgba(15,23,42,0.08)' }}>
      <div style={{ display: 'flex', alignItems: 'baseline', gap: 14, marginBottom: 18 }}>
        <BrandLockup size={26} sub />
        <span className="rh-eyebrow-muted">Атомы и молекулы — оригинал в design-system, тут — выжимка</span>
      </div>

      {/* Row 1: Colors / Type / Shadows / Radii */}
      <div className="ks-row" style={{ marginBottom: 14 }}>
        <ShowcaseTile title="Brand · 5 / 5">
          <div className="ks-swatch">
            {[
              ['#0f766e', 'Teal'],
              ['#0a5f59', 'Strong'],
              ['#d97706', 'Amber'],
              ['#16212b', 'Ink'],
              ['#f8f4ec', 'Sand'],
            ].map(([c, l]) => (
              <div key={l} title={l} className="ks-chip" style={{ background: c }} />
            ))}
          </div>
          <div className="ks-stack" style={{ marginTop: 12 }}>
            <span style={{ fontSize: 12, color: '#5f6877' }}>Semantic: green / gold / red / cyan</span>
            <div className="ks-swatch">
              <span className="rh-tag rh-tag--green">Подтверждено</span>
              <span className="rh-tag rh-tag--gold">Ждёт</span>
              <span className="rh-tag rh-tag--red">Отменено</span>
              <span className="rh-tag rh-tag--cyan">Бассейн</span>
            </div>
          </div>
        </ShowcaseTile>

        <ShowcaseTile title="Type · Manrope">
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4 }}>
            <span className="rh-eyebrow">Eyebrow · 12 / 800</span>
            <h2 className="rh-h2" style={{ marginTop: 4 }}>H2 · Берёзовая роща</h2>
            <h3 className="rh-h3">H3 · Свободно сегодня</h3>
            <p className="rh-body" style={{ margin: 0 }}>Body — частная баня на дровах с двумя парными.</p>
            <span className="rh-meta">Meta · Москва, Рублёвское ш., 12</span>
          </div>
        </ShowcaseTile>

        <ShowcaseTile title="Buttons">
          <div className="ks-stack" style={{ gap: 10 }}>
            <button className="rh-btn rh-btn--primary"><Icon name="plus" size={14} /> Primary CTA</button>
            <button className="rh-btn rh-btn--default"><Icon name="cal" size={14} /> Default</button>
            <button className="rh-btn rh-btn--ghost rh-btn--sm">Ghost · Sm</button>
            <button className="rh-btn rh-btn--primary rh-btn--lg" style={{ width: '100%' }}>Large wide</button>
          </div>
        </ShowcaseTile>

        <ShowcaseTile title="Avatars · Stars">
          <div className="ks-swatch" style={{ gap: 10, marginBottom: 14 }}>
            <Avatar name="Иван Соколов" tone="amber" />
            <Avatar name="Анна Поляк" tone="teal" />
            <Avatar name="Мария Лисс" tone="plum" />
            <Avatar name="Дмитрий М" tone="blue" />
          </div>
          <div className="ks-stack" style={{ gap: 6 }}>
            <Stars value={4.9} count={128} />
            <Stars value={4.8} count={96} />
            <Stars value={4.6} count={312} />
          </div>
        </ShowcaseTile>
      </div>

      {/* Row 2: Tags / Card / Photo / Icons */}
      <div className="ks-row">
        <ShowcaseTile title="Tags / Pills">
          <div className="ks-swatch">
            <span className="rh-tag rh-tag--primary">Сауна</span>
            <span className="rh-tag">Бассейн</span>
            <span className="rh-tag">Чан</span>
            <span className="rh-tag rh-tag--gold">Last min −20%</span>
            <span className="rh-tag rh-tag--green">Проверено</span>
            <span className="rh-tag rh-tag--ghost">Мангал</span>
          </div>
        </ShowcaseTile>

        <ShowcaseTile title="Glass card">
          <div className="rh-card" style={{ padding: 14, borderRadius: 18 }}>
            <span className="rh-eyebrow-muted">Кошелёк</span>
            <div style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.04em' }}>4 280 ₽</div>
            <div style={{ fontSize: 11, color: '#5f6877' }}>+ 540 ₽ за апрель</div>
          </div>
        </ShowcaseTile>

        <ShowcaseTile title="Photo placeholders">
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8 }}>
            <PhotoPlaceholder height={66} radius={12} variant="banya" label="" />
            <PhotoPlaceholder height={66} radius={12} variant="forest" label="" />
            <PhotoPlaceholder height={66} radius={12} variant="city" label="" />
            <PhotoPlaceholder height={66} radius={12} variant="sand" label="" />
          </div>
        </ShowcaseTile>

        <ShowcaseTile title="Icons · 1.75 stroke">
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(6, 1fr)', gap: 10, alignItems: 'center', justifyItems: 'center', color: '#16212b' }}>
            {['search','pin','cal','heart','msg','user','star','fire','bolt','gift','shield','set'].map(n => (
              <Icon key={n} name={n} size={20} />
            ))}
          </div>
        </ShowcaseTile>
      </div>
    </div>
  );
}

// ── Definition: every screen as a record ────────────────────────
//
// kind: 'web' | 'mobile' | 'showcase'
// surface: 'public' | 'client' | 'owner' | 'system'
// platform: 'web' | 'ios'
// height: artboard height for DCArtboard
// render: function returning React element
const SCREENS = [
  // Public web — 5
  { id: 'w-home',      kind: 'web',    surface: 'public', platform: 'web', label: '01 · Home',           h: 1700, render: () => <RH.PageHome /> },
  { id: 'w-cat',       kind: 'web',    surface: 'public', platform: 'web', label: '02 · Catalogue',      h: 1500, render: () => <RH.PageCatalogue /> },
  { id: 'w-list',      kind: 'web',    surface: 'public', platform: 'web', label: '03 · Listing',        h: 1900, render: () => <RH.PageListing /> },
  { id: 'w-book',      kind: 'web',    surface: 'public', platform: 'web', label: '04 · Checkout',       h: 1500, render: () => <RH.PageBooking /> },
  { id: 'w-auth',      kind: 'web',    surface: 'public', platform: 'web', label: '05 · Sign-in / Up',   h:  920, render: () => <RH.PageAuth /> },

  // Client web cabinet — 3
  { id: 'w-cl-book',   kind: 'web',    surface: 'client', platform: 'web', label: '06 · My bookings',    h: 1100, render: () => <RH.PageClientBookings /> },
  { id: 'w-cl-fav',    kind: 'web',    surface: 'client', platform: 'web', label: '07 · Favorites',      h: 1100, render: () => <RH.PageClientFavorites /> },
  { id: 'w-cl-wal',    kind: 'web',    surface: 'client', platform: 'web', label: '08 · Wallet & promo', h: 1000, render: () => <RH.PageClientWallet /> },

  // Owner web cabinet — 3
  { id: 'w-ow-dash',   kind: 'web',    surface: 'owner',  platform: 'web', label: '09 · Dashboard',      h: 1500, render: () => <RH.PageOwnerDashboard /> },
  { id: 'w-ow-cal',    kind: 'web',    surface: 'owner',  platform: 'web', label: '10 · Calendar',       h: 1200, render: () => <RH.PageOwnerCalendar /> },
  { id: 'w-ow-obj',    kind: 'web',    surface: 'owner',  platform: 'web', label: '11 · Objects',        h: 1100, render: () => <RH.PageOwnerObjects /> },

  // Client mobile (iOS) — 5
  { id: 'm-cl-home',   kind: 'mobile', surface: 'client', platform: 'ios', label: '12 · Home',           render: () => <RHM.MClientHome /> },
  { id: 'm-cl-srch',   kind: 'mobile', surface: 'client', platform: 'ios', label: '13 · Search',         render: () => <RHM.MClientSearch /> },
  { id: 'm-cl-list',   kind: 'mobile', surface: 'client', platform: 'ios', label: '14 · Listing',        render: () => <RHM.MClientListing /> },
  { id: 'm-cl-book',   kind: 'mobile', surface: 'client', platform: 'ios', label: '15 · Bookings',       render: () => <RHM.MClientBookings /> },
  { id: 'm-cl-prof',   kind: 'mobile', surface: 'client', platform: 'ios', label: '16 · Profile',        render: () => <RHM.MClientProfile /> },

  // Owner mobile (iOS) — 4
  { id: 'm-ow-tod',    kind: 'mobile', surface: 'owner',  platform: 'ios', label: '17 · Today',          render: () => <RHM.MOwnerToday /> },
  { id: 'm-ow-cal',    kind: 'mobile', surface: 'owner',  platform: 'ios', label: '18 · Calendar',       render: () => <RHM.MOwnerCalendar /> },
  { id: 'm-ow-inb',    kind: 'mobile', surface: 'owner',  platform: 'ios', label: '19 · Inbox',          render: () => <RHM.MOwnerInbox /> },
  { id: 'm-ow-stat',   kind: 'mobile', surface: 'owner',  platform: 'ios', label: '20 · Analytics',      render: () => <RHM.MOwnerStats /> },
];

// ── Tweak defaults (persisted via the host bridge) ──────────────
const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "surfaces": "all",
  "platforms": "all",
  "density": "comfortable",
  "darkMobile": false,
  "showShowcase": true,
  "language": "ru"
}/*EDITMODE-END*/;

// ── Filter logic ────────────────────────────────────────────────
function filterScreens(t) {
  return SCREENS.filter(s => {
    if (t.surfaces !== 'all' && s.surface !== t.surfaces) return false;
    if (t.platforms !== 'all' && s.platform !== t.platforms) return false;
    return true;
  });
}

// ── Main app ────────────────────────────────────────────────────
function App() {
  const [tweaks, setTweak] = useTweaks(TWEAK_DEFAULTS);
  const filtered = filterScreens(tweaks);

  // group screens by surface for sectioning
  const bySurface = (surface) => filtered.filter(s => s.surface === surface);

  const publicWeb   = bySurface('public');
  const clientWeb   = filtered.filter(s => s.surface === 'client' && s.platform === 'web');
  const ownerWeb    = filtered.filter(s => s.surface === 'owner'  && s.platform === 'web');
  const clientIOS   = filtered.filter(s => s.surface === 'client' && s.platform === 'ios');
  const ownerIOS    = filtered.filter(s => s.surface === 'owner'  && s.platform === 'ios');

  // Helper: render web artboards within a section
  const renderArtboard = (s) => {
    if (s.kind === 'web') {
      return (
        <DCArtboard key={s.id} id={s.id} label={s.label} width={1280} height={s.h}>
          <WebFrame width={1280} height={s.h}>
            <div className={tweaks.density === 'compact' ? 'density-compact' : ''}>
              {s.render()}
            </div>
          </WebFrame>
        </DCArtboard>
      );
    }
    if (s.kind === 'mobile') {
      return (
        <DCArtboard key={s.id} id={s.id} label={s.label} width={402} height={874}>
          <MobileFrame dark={tweaks.darkMobile}>
            {s.render()}
          </MobileFrame>
        </DCArtboard>
      );
    }
    return null;
  };

  return (
    <>
      <DesignCanvas>
        {publicWeb.length > 0 && (
          <DCSection id="public-web" title="Public · Web" subtitle="Маркетинг + поиск + бронирование. 1280 px.">
            {publicWeb.map(renderArtboard)}
            <DCPostIt x={20} y={40}>Цельный flow от главной до оплаты — клиент знает цену, дату, имя хозяина и сумму к оплате до клика «Перейти к оплате».</DCPostIt>
          </DCSection>
        )}

        {clientWeb.length > 0 && (
          <DCSection id="client-web" title="Client · Web cabinet" subtitle="Кабинет гостя · мои брони, избранное, кошелёк и бонусы.">
            {clientWeb.map(renderArtboard)}
            <DCPostIt x={20} y={40} color="amber">Tab-навигация по кабинету одинаковая на всех 3 экранах — общая &laquo;ClientHeader&raquo;.</DCPostIt>
          </DCSection>
        )}

        {ownerWeb.length > 0 && (
          <DCSection id="owner-web" title="Owner · Web cabinet" subtitle="Кабинет владельца — операционка: брони, расписание, объекты, выручка.">
            {ownerWeb.map(renderArtboard)}
            <DCPostIt x={20} y={40} color="teal">У владельца другая «температура» интерфейса — больше плотности, меньше воздуха. Тот же дизайн-язык, но 14-day chart и календарь — это рабочая поверхность.</DCPostIt>
          </DCSection>
        )}

        {clientIOS.length > 0 && (
          <DCSection id="client-ios" title="Client · iOS app" subtitle="Мобильное приложение гостя · iOS 26 · 402×874 px (iPhone 15 Pro).">
            {clientIOS.map(renderArtboard)}
            <DCPostIt x={20} y={40}>Один поиск + сценарии = 80% сценариев попадания в бронь без листания каталога.</DCPostIt>
          </DCSection>
        )}

        {ownerIOS.length > 0 && (
          <DCSection id="owner-ios" title="Owner · iOS app" subtitle="Мобильное приложение владельца · быстрые подтверждения, расписание, метрики.">
            {ownerIOS.map(renderArtboard)}
            <DCPostIt x={20} y={40} color="amber">Главный workflow — &laquo;Подтвердить / Отклонить&raquo; в один тап. Всё остальное — справочно.</DCPostIt>
          </DCSection>
        )}

        {tweaks.showShowcase && (
          <DCSection id="components" title="Components & tokens" subtitle="Атомы дизайн-системы. Manrope · #0f766e · 4-px spacing · 12/16/22/28 радиусы.">
            <DCArtboard id="kit-1" label="Atoms · Tokens" width={1280} height={520}>
              <ComponentShowcase />
            </DCArtboard>
            <DCPostIt x={20} y={40} color="violet">Если что-то нельзя собрать из этих 12 атомов — это исключение, не правило. Сначала пробуем собрать.</DCPostIt>
          </DCSection>
        )}
      </DesignCanvas>

      <TweaksPanel title="Tweaks">
        <TweakSection label="Surface">
          <TweakRadio
            label="Roles"
            value={tweaks.surfaces}
            options={[
              { value: 'all', label: 'All' },
              { value: 'public', label: 'Public' },
              { value: 'client', label: 'Client' },
              { value: 'owner', label: 'Owner' },
            ]}
            onChange={(v) => setTweak('surfaces', v)}
          />
          <TweakRadio
            label="Platform"
            value={tweaks.platforms}
            options={[
              { value: 'all', label: 'All' },
              { value: 'web', label: 'Web' },
              { value: 'ios', label: 'iOS' },
            ]}
            onChange={(v) => setTweak('platforms', v)}
          />
        </TweakSection>

        <TweakSection label="Display">
          <TweakRadio
            label="Density"
            value={tweaks.density}
            options={[
              { value: 'comfortable', label: 'Comfortable' },
              { value: 'compact', label: 'Compact' },
            ]}
            onChange={(v) => setTweak('density', v)}
          />
          <TweakToggle
            label="iOS dark mode"
            value={tweaks.darkMobile}
            onChange={(v) => setTweak('darkMobile', v)}
          />
          <TweakToggle
            label="Show component kit"
            value={tweaks.showShowcase}
            onChange={(v) => setTweak('showShowcase', v)}
          />
        </TweakSection>

        <TweakSection label="Locale">
          <TweakSelect
            label="Language"
            value={tweaks.language}
            options={[
              { value: 'ru', label: 'Русский' },
              { value: 'en', label: 'English (placeholder)' },
            ]}
            onChange={(v) => setTweak('language', v)}
          />
        </TweakSection>

        <TweakSection label="Jump to">
          {[
            ['public-web', '↑ Public · Web'],
            ['client-web', '↑ Client · Web'],
            ['owner-web',  '↑ Owner · Web'],
            ['client-ios', '↑ Client · iOS'],
            ['owner-ios',  '↑ Owner · iOS'],
          ].map(([sid, label]) => (
            <TweakButton
              key={sid}
              label={label}
              secondary
              onClick={() => {
                const el = document.querySelector(`[data-dc-section="${sid}"]`);
                if (!el) return;
                const top = el.getBoundingClientRect().top + window.scrollY - 60;
                window.scrollTo({ top, behavior: 'smooth' });
              }}
            />
          ))}
        </TweakSection>
      </TweaksPanel>
    </>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />);
