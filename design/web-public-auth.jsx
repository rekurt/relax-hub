/* global React */
const { Icon, BrandLockup, Avatar, PhotoPlaceholder } = window.RH;

// ════════════════════════════════════════════════════════════════
// PUBLIC · 05 AUTH (split)
// ════════════════════════════════════════════════════════════════
function PageAuth() {
  return (
    <div className="rh-page" style={{ minHeight: 880, padding: 28 }}>
      <div className="rh-card" style={{
        display: 'grid', gridTemplateColumns: '1.05fr 0.95fr',
        borderRadius: 36, overflow: 'hidden', minHeight: 824, padding: 0,
      }}>
        {/* Brand aside */}
        <aside style={{
          padding: 40, color: '#fffdf8',
          background: 'linear-gradient(135deg, #10313a 0%, #38606a 44%, #9a5c30 100%)',
          display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
          position: 'relative', overflow: 'hidden',
        }}>
          <div style={{ position: 'absolute', inset: 0, opacity: 0.18 }}>
            <svg width="100%" height="100%"><circle cx="20%" cy="20%" r="120" stroke="#fff" strokeWidth="1" fill="none"/><circle cx="80%" cy="78%" r="160" stroke="#fff" strokeWidth="1" fill="none"/></svg>
          </div>
          <div style={{ position: 'relative' }}>
            <BrandLockup size={28} dark sub />
          </div>
          <div style={{ position: 'relative' }}>
            <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>Кабинет владельца</span>
            <h1 style={{ margin: '14px 0 14px', fontSize: 44, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.05 }}>
              Бронирования без звонков и Excel.
            </h1>
            <p style={{ margin: 0, fontSize: 15, lineHeight: 1.6, color: 'rgba(255,255,255,0.84)', maxWidth: 420 }}>
              Календарь, чат с гостями, динамические цены и аналитика — в одной вкладке.
              Подключение за день, 0 % комиссии с бронирований.
            </p>
            <div style={{ display: 'flex', gap: 10, marginTop: 22, flexWrap: 'wrap' }}>
              {['0 % комиссии', 'Календарь', 'Динамические цены', 'CRM'].map(t => (
                <span key={t} style={{ padding: '6px 12px', borderRadius: 999, fontSize: 12, fontWeight: 700, background: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.14)' }}>{t}</span>
              ))}
            </div>
          </div>
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center', gap: 12, padding: 16, borderRadius: 20, background: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.14)' }}>
            <Avatar name="Иван Соколов" size={40} tone="amber" />
            <div style={{ lineHeight: 1.3 }}>
              <div style={{ fontSize: 13, fontWeight: 700 }}>«За месяц снял 70 % звонков. Гости сами выбирают слот.»</div>
              <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.72)' }}>Иван Соколов · «Берёзовая роща»</div>
            </div>
          </div>
        </aside>

        {/* Form */}
        <main style={{ padding: 40, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
          <div style={{ maxWidth: 380, margin: '0 auto', width: '100%' }}>
            <div style={{ display: 'inline-flex', padding: 4, borderRadius: 999, background: 'rgba(15,23,42,0.06)', marginBottom: 22 }}>
              {['Войти', 'Создать аккаунт'].map((l, i) => (
                <button key={l} style={{
                  padding: '8px 18px', borderRadius: 999, border: 0, cursor: 'pointer',
                  fontSize: 13, fontWeight: 700,
                  background: i === 0 ? '#fff' : 'transparent',
                  color: i === 0 ? '#16212b' : '#5f6877',
                  boxShadow: i === 0 ? '0 6px 14px rgba(15,23,42,0.08)' : 'none',
                }}>{l}</button>
              ))}
            </div>

            <h2 style={{ margin: '0 0 6px', fontSize: 30, fontWeight: 800, letterSpacing: '-0.04em' }}>С возвращением</h2>
            <p style={{ margin: '0 0 22px', fontSize: 14, color: '#5f6877' }}>Войдите по телефону — отправим SMS-код. Без паролей.</p>

            <Field label="Телефон" value="+7 (903) 555-12-44" focus />

            <button className="rh-btn rh-btn--primary rh-btn--lg" style={{ width: '100%', marginTop: 18 }}>
              Получить код <Icon name="arrow" size={16} />
            </button>

            <div style={{ display: 'flex', alignItems: 'center', gap: 12, margin: '22px 0' }}>
              <div style={{ flex: 1, height: 1, background: 'rgba(15,23,42,0.10)' }} />
              <span style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.14em', textTransform: 'uppercase', color: '#5f6877' }}>или</span>
              <div style={{ flex: 1, height: 1, background: 'rgba(15,23,42,0.10)' }} />
            </div>

            <button className="rh-btn rh-btn--default" style={{ width: '100%', marginBottom: 8 }}>
              <Icon name="qr" size={16} /> Войти через Telegram
            </button>
            <button className="rh-btn rh-btn--default" style={{ width: '100%' }}>
              <Icon name="globe" size={16} /> Войти через Apple ID
            </button>

            <p style={{ fontSize: 12, color: '#5f6877', textAlign: 'center', marginTop: 22, lineHeight: 1.55 }}>
              Создавая аккаунт, вы принимаете <a href="#" style={{ color: '#0a5f59' }}>условия</a> и <a href="#" style={{ color: '#0a5f59' }}>политику конфиденциальности</a> RelaxHUB.
            </p>
          </div>
        </main>
      </div>
    </div>
  );
}

function Field({ label, value, focus }) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
      <span style={{ fontSize: 12, fontWeight: 700 }}>{label}</span>
      <div style={{
        height: 52, padding: '0 14px', borderRadius: 16,
        border: focus ? '1.5px solid rgba(15,118,110,0.52)' : '1px solid rgba(15,23,42,0.14)',
        background: focus ? '#fff' : 'linear-gradient(180deg, rgba(255,255,255,0.98), rgba(251,247,240,0.96))',
        boxShadow: focus ? '0 0 0 4px rgba(15,118,110,0.10), 0 18px 32px rgba(15,23,42,0.08)' : 'inset 0 1px 0 rgba(255,255,255,0.88), 0 10px 24px rgba(15,23,42,0.04)',
        display: 'flex', alignItems: 'center',
        fontSize: 15, fontWeight: 600,
        transform: focus ? 'translateY(-1px)' : 'none',
      }}>{value}</div>
    </label>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageAuth });
