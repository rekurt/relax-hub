/* global React */
const { Icon, BrandLockup, Avatar } = window.RH;

// ── Top nav (public + client) ─────────────────────────────────────
function TopNav({ active = 'catalogue', authed = false }) {
  const links = [
    ['home', 'Главная'],
    ['catalogue', 'Каталог'],
    ['scenarios', 'Сценарии'],
    ['certificates', 'Сертификаты'],
    ['help', 'Помощь'],
  ];
  return (
    <header style={{
      display: 'flex', alignItems: 'center', gap: 24,
      padding: '14px 28px',
      background: 'rgba(255, 252, 247, 0.84)',
      backdropFilter: 'blur(20px)', WebkitBackdropFilter: 'blur(20px)',
      borderBottom: '1px solid rgba(15,23,42,0.08)',
      boxShadow: '0 4px 14px rgba(15, 23, 42, 0.04)',
      position: 'relative', zIndex: 4,
    }}>
      <BrandLockup size={22} sub />
      <nav style={{ display: 'flex', gap: 4, marginLeft: 28, flex: 1 }}>
        {links.map(([k, label]) => (
          <a key={k} href="#" style={{
            padding: '8px 14px', borderRadius: 999, textDecoration: 'none',
            fontSize: 14, fontWeight: 600,
            color: k === active ? '#0a5f59' : '#16212b',
            background: k === active ? 'rgba(15, 118, 110, 0.10)' : 'transparent',
          }}>{label}</a>
        ))}
      </nav>
      <button style={navIconBtn}><Icon name="globe" size={18} /></button>
      {authed ? (
        <>
          <button style={navIconBtn}><Icon name="bell" size={18} /></button>
          <button style={navIconBtn}><Icon name="heart" size={18} /></button>
          <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, padding: '4px 14px 4px 4px', borderRadius: 999, background: 'rgba(255,255,255,0.7)', border: '1px solid rgba(15,23,42,0.1)' }}>
            <Avatar name="Анна Поляк" size={32} tone="teal" />
            <span style={{ fontSize: 13, fontWeight: 600 }}>Анна</span>
          </div>
        </>
      ) : (
        <div style={{ display: 'flex', gap: 8 }}>
          <button className="rh-btn rh-btn--default rh-btn--sm">Войти</button>
          <button className="rh-btn rh-btn--primary rh-btn--sm">Стать партнёром</button>
        </div>
      )}
    </header>
  );
}

const navIconBtn = {
  width: 38, height: 38, borderRadius: 999,
  background: 'rgba(255,255,255,0.62)', border: '1px solid rgba(15,23,42,0.08)',
  display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
  cursor: 'pointer', color: '#16212b',
};

// ── Owner cabinet horizontal nav ─────────────────────────────────
function OwnerNav({ active = 'bookings' }) {
  const links = [
    ['dashboard', 'Сводка'],
    ['bookings', 'Бронирования'],
    ['calendar', 'Календарь'],
    ['objects', 'Объекты'],
    ['pricing', 'Цены и акции'],
    ['reviews', 'Отзывы'],
    ['payouts', 'Выплаты'],
    ['settings', 'Настройки'],
  ];
  return (
    <header style={{
      display: 'flex', alignItems: 'center', gap: 18,
      padding: '14px 28px',
      background: 'rgba(255, 252, 247, 0.86)',
      backdropFilter: 'blur(20px)', WebkitBackdropFilter: 'blur(20px)',
      borderBottom: '1px solid rgba(15,23,42,0.08)',
      boxShadow: '0 4px 14px rgba(15, 23, 42, 0.04)',
    }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
        <BrandLockup size={20} />
        <span style={{
          padding: '4px 10px', borderRadius: 999,
          background: 'rgba(217,119,6,0.12)', border: '1px solid rgba(217,119,6,0.28)',
          color: '#92400e', fontSize: 11, fontWeight: 700,
          letterSpacing: '0.08em', textTransform: 'uppercase',
        }}>Кабинет владельца</span>
      </div>
      <nav style={{ display: 'flex', gap: 2, marginLeft: 18, flex: 1, flexWrap: 'wrap' }}>
        {links.map(([k, label]) => (
          <a key={k} href="#" style={{
            padding: '8px 12px', borderRadius: 999, textDecoration: 'none',
            fontSize: 13, fontWeight: 600,
            color: k === active ? '#0a5f59' : '#16212b',
            background: k === active ? 'rgba(15, 118, 110, 0.10)' : 'transparent',
          }}>{label}</a>
        ))}
      </nav>
      <button style={navIconBtn}><Icon name="msg" size={18} /></button>
      <button style={navIconBtn}><Icon name="bell" size={18} /></button>
      <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, padding: '4px 14px 4px 4px', borderRadius: 999, background: 'rgba(255,255,255,0.7)', border: '1px solid rgba(15,23,42,0.1)' }}>
        <Avatar name="Иван Соколов" size={32} tone="amber" />
        <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.1 }}>
          <span style={{ fontSize: 13, fontWeight: 700 }}>Иван С.</span>
          <span style={{ fontSize: 11, color: '#5f6877' }}>3 объекта</span>
        </div>
      </div>
    </header>
  );
}

// ── Footer ───────────────────────────────────────────────────────
function Footer() {
  const cols = [
    ['Каталог', ['Все бани', 'Москва', 'Санкт-Петербург', 'Сочи', 'Казань']],
    ['Сценарии', ['Вечер вдвоём', 'Компания друзей', 'Семья с детьми', 'Корпоратив', 'Last minute']],
    ['Партнёрам', ['Стать партнёром', 'Кабинет владельца', '0 % комиссии', 'API', 'Маркетинг']],
    ['Поддержка', ['Помощь', 'Контакты', 'Возвраты', 'Сертификаты', 'Безопасность']],
  ];
  return (
    <footer style={{
      marginTop: 60,
      padding: '52px 32px 28px',
      borderRadius: '32px 32px 0 0',
      background: 'radial-gradient(circle at top left, rgba(15, 118, 110, 0.32), transparent 28%), linear-gradient(180deg, rgba(17, 27, 33, 0.98), rgba(12, 18, 24, 0.98))',
      color: '#f7f4eb',
      boxShadow: '0 -30px 80px rgba(15, 23, 42, 0.18)',
    }}>
      <div style={{ display: 'grid', gridTemplateColumns: '1.4fr repeat(4, 1fr)', gap: 36, marginBottom: 36 }}>
        <div>
          <BrandLockup size={28} dark sub />
          <p style={{ marginTop: 18, fontSize: 13, lineHeight: 1.6, color: 'rgba(247,244,235,0.66)', maxWidth: 280 }}>
            Платформа для аккуратного выбора и бронирования приватных бань без лишнего шума в интерфейсе.
          </p>
          <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
            {['Telegram', 'VK', 'Дзен'].map(s => (
              <span key={s} style={{ padding: '6px 12px', borderRadius: 999, fontSize: 11, fontWeight: 700, letterSpacing: '0.08em', textTransform: 'uppercase', background: 'rgba(255,255,255,0.06)', border: '1px solid rgba(255,255,255,0.10)' }}>{s}</span>
            ))}
          </div>
        </div>
        {cols.map(([title, items]) => (
          <div key={title}>
            <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', color: 'rgba(247,244,235,0.66)', marginBottom: 14 }}>{title}</div>
            {items.map(it => (
              <a key={it} href="#" style={{ display: 'block', padding: '5px 0', textDecoration: 'none', color: '#f7f4eb', fontSize: 13, fontWeight: 500 }}>{it}</a>
            ))}
          </div>
        ))}
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', paddingTop: 22, borderTop: '1px solid rgba(255,255,255,0.10)', fontSize: 12, color: 'rgba(247,244,235,0.66)' }}>
        <span>© 2026 RelaxHUB · Все права защищены</span>
        <span>support@relaxhub.ru · Ежедневно с 09:00 до 22:00 МСК</span>
        <span>relaxhub.ru</span>
      </div>
    </footer>
  );
}

// ── Section wrapper used inside pages ────────────────────────────
function Section({ children, padding = '36px 28px', background, style = {} }) {
  return (
    <section style={{ padding, background, ...style }}>{children}</section>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { TopNav, OwnerNav, Footer, Section });
