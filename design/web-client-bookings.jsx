/* global React */
const { Icon, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;
const { ListingCard, SectionHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// CLIENT · 06 MY BOOKINGS
// ════════════════════════════════════════════════════════════════
function PageClientBookings() {
  return (
    <div className="rh-page" style={{ minHeight: 1100 }}>
      <TopNav active="cabinet" authed />

      <Section padding="22px 28px 0">
        <ClientHeader active="bookings" />
      </Section>

      <Section padding="14px 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 320px', gap: 24, alignItems: 'flex-start' }}>
          <div>
            {/* Tabs */}
            <div style={{ display: 'inline-flex', padding: 4, borderRadius: 999, background: 'rgba(255,255,255,0.62)', border: '1px solid rgba(15,23,42,0.06)', marginBottom: 18 }}>
              {[['Активные', 2, true], ['История', 14, false], ['Отменённые', 1, false]].map(([l, n, on]) => (
                <button key={l} style={{
                  padding: '8px 18px', borderRadius: 999, border: 0, cursor: 'pointer',
                  background: on ? '#fff' : 'transparent', boxShadow: on ? '0 6px 14px rgba(15,23,42,0.08)' : 'none',
                  fontSize: 13, fontWeight: 700, color: on ? '#16212b' : '#5f6877',
                  display: 'inline-flex', alignItems: 'center', gap: 8,
                }}>{l}<span style={{ padding: '1px 8px', fontSize: 11, borderRadius: 999, background: on ? 'rgba(15,118,110,0.10)' : 'rgba(15,23,42,0.06)', color: on ? '#0a5f59' : '#5f6877' }}>{n}</span></button>
              ))}
            </div>

            {/* Upcoming */}
            <BookingRow
              variant="banya" name="Берёзовая роща" addr="Москва, Рублёвское ш., 12"
              when="Сб, 20 апр · 19:00–23:00" guests="4 гостя · 4 часа"
              total="11 440 ₽" status="confirmed" code="RH-4821"
              host="Иван Соколов" hostTone="amber"
            />
            <BookingRow
              variant="forest" name="Сосновый берег" addr="Истра, дер. Луцино"
              when="Сб, 4 мая · 16:00–20:00" guests="6 гостей · 4 часа"
              total="22 200 ₽" status="awaiting" code="RH-4860"
              host="Мария Лисс" hostTone="plum"
            />

            {/* Empty state spec */}
            <div className="rh-card rh-card--flat" style={{ marginTop: 18, padding: 22, borderRadius: 22, display: 'flex', alignItems: 'center', gap: 14 }}>
              <Icon name="cal" size={20} style={{ color: '#5f6877' }} />
              <div style={{ flex: 1, fontSize: 13, color: '#5f6877' }}>
                <div style={{ fontWeight: 700, color: '#16212b', marginBottom: 2 }}>Запланируйте следующий вечер</div>
                Подборка из ваших избранных мест ждёт ниже — большая часть свободна на эти выходные.
              </div>
              <button className="rh-btn rh-btn--primary rh-btn--sm">Открыть каталог <Icon name="arrow" size={14} /></button>
            </div>
          </div>

          {/* Right rail */}
          <aside style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div className="rh-card" style={{ padding: 22, borderRadius: 24 }}>
              <span className="rh-eyebrow-muted">Кошелёк</span>
              <div style={{ fontSize: 30, fontWeight: 800, letterSpacing: '-0.04em', marginTop: 4 }}>4 280 ₽</div>
              <div style={{ fontSize: 12, color: '#5f6877', marginBottom: 14 }}>+ 3 % бонусов с каждой брони</div>
              <button className="rh-btn rh-btn--primary rh-btn--sm" style={{ width: '100%' }}><Icon name="plus" size={14} /> Пополнить</button>
            </div>

            <div className="rh-card" style={{ padding: 20, borderRadius: 24 }}>
              <span className="rh-eyebrow-muted">Уровень</span>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 6 }}>
                <Icon name="fire" size={20} style={{ color: '#d97706' }} />
                <div>
                  <div style={{ fontSize: 16, fontWeight: 800 }}>Завсегдатай</div>
                  <div style={{ fontSize: 11, color: '#5f6877' }}>3 брони до уровня «Хозяин пара»</div>
                </div>
              </div>
              <div style={{ height: 8, marginTop: 12, borderRadius: 999, background: 'rgba(15,23,42,0.08)', overflow: 'hidden' }}>
                <div style={{ width: '62%', height: '100%', background: 'linear-gradient(90deg,#d97706,#b45309)' }} />
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: '#5f6877', marginTop: 6 }}>
                <span>5 / 8</span><span>+5 % бонусов</span>
              </div>
            </div>

            <div className="rh-card" style={{ padding: 18, borderRadius: 24 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <div style={{ width: 36, height: 36, borderRadius: 12, background: 'rgba(217,119,6,0.10)', color: '#92400e', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icon name="gift" size={18} /></div>
                <div style={{ lineHeight: 1.3 }}>
                  <div style={{ fontSize: 13, fontWeight: 700 }}>Сертификат «На двоих»</div>
                  <div style={{ fontSize: 11, color: '#5f6877' }}>5 000 ₽ · до 31 мая</div>
                </div>
              </div>
            </div>
          </aside>
        </div>
      </Section>

      <Footer />
    </div>
  );
}

// ── Client cabinet shared header ───────────────────────────────
function ClientHeader({ active }) {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: 18 }}>
        <div>
          <span className="rh-eyebrow">Кабинет гостя</span>
          <h1 style={{ margin: '8px 0 6px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em' }}>Привет, Анна</h1>
          <p style={{ margin: 0, fontSize: 14, color: '#5f6877' }}>Ваши бронирования, избранное и сертификаты — без переключения вкладок.</p>
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <button className="rh-btn rh-btn--default"><Icon name="set" size={16} /> Настройки</button>
          <button className="rh-btn rh-btn--primary"><Icon name="plus" size={16} /> Новое бронирование</button>
        </div>
      </div>
      <nav style={{ display: 'flex', gap: 6, marginTop: 22, padding: 6, borderRadius: 999, background: 'rgba(255,255,255,0.62)', border: '1px solid rgba(15,23,42,0.06)', width: 'fit-content' }}>
        {[
          ['bookings', 'Бронирования', 'cal'],
          ['favorites', 'Избранное', 'heart'],
          ['wallet', 'Кошелёк и бонусы', 'gift'],
          ['certs', 'Сертификаты', 'card'],
          ['chat', 'Чат с хозяевами', 'msg'],
        ].map(([k, label, ic]) => (
          <button key={k} style={{
            padding: '8px 16px', borderRadius: 999, border: 0, cursor: 'pointer',
            background: k === active ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'transparent',
            color: k === active ? '#fffdf8' : '#16212b',
            fontSize: 13, fontWeight: 700,
            display: 'inline-flex', alignItems: 'center', gap: 8,
            boxShadow: k === active ? '0 10px 22px rgba(15,118,110,0.22)' : 'none',
          }}><Icon name={ic} size={14} /> {label}</button>
        ))}
      </nav>
    </div>
  );
}

function BookingRow({ variant, name, addr, when, guests, total, status, code, host, hostTone }) {
  const statusMap = {
    confirmed: { color: '#15803d', bg: 'rgba(21,128,61,0.10)', border: 'rgba(21,128,61,0.24)', label: 'Подтверждено' },
    awaiting:  { color: '#b45309', bg: 'rgba(217,119,6,0.12)', border: 'rgba(217,119,6,0.28)', label: 'Ждёт хозяина' },
    cancelled: { color: '#991b1b', bg: 'rgba(180,35,24,0.08)', border: 'rgba(180,35,24,0.22)', label: 'Отменено' },
  };
  const s = statusMap[status];
  return (
    <div className="rh-card" style={{ padding: 22, borderRadius: 24, marginBottom: 14, display: 'grid', gridTemplateColumns: '160px 1fr auto', gap: 22, alignItems: 'center' }}>
      <PhotoPlaceholder width={160} height={120} radius={18} variant={variant} label="" />
      <div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
          <span style={{ padding: '4px 10px', borderRadius: 999, fontSize: 11, fontWeight: 700, color: s.color, background: s.bg, border: `1px solid ${s.border}` }}>{s.label}</span>
          <span style={{ fontSize: 11, color: '#5f6877', fontWeight: 700, letterSpacing: '0.10em', textTransform: 'uppercase' }}>{code}</span>
        </div>
        <h3 style={{ margin: '0 0 6px', fontSize: 20, fontWeight: 800, letterSpacing: '-0.02em' }}>{name}</h3>
        <div style={{ display: 'flex', gap: 16, color: '#5f6877', fontSize: 12, fontWeight: 600, marginBottom: 10 }}>
          <span style={{ display: 'inline-flex', gap: 4, alignItems: 'center' }}><Icon name="pin" size={12} /> {addr}</span>
          <span style={{ display: 'inline-flex', gap: 4, alignItems: 'center' }}><Icon name="cal" size={12} /> {when}</span>
          <span style={{ display: 'inline-flex', gap: 4, alignItems: 'center' }}><Icon name="user" size={12} /> {guests}</span>
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <Avatar name={host} size={26} tone={hostTone} />
          <span style={{ fontSize: 12, color: '#16212b', fontWeight: 600 }}>{host} · хозяин</span>
          <button style={{ background: 'none', border: 0, cursor: 'pointer', color: '#0a5f59', fontWeight: 700, fontSize: 12, display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            <Icon name="msg" size={12} /> Написать
          </button>
        </div>
      </div>
      <div style={{ textAlign: 'right' }}>
        <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em' }}>{total}</div>
        <div style={{ fontSize: 11, color: '#5f6877', marginBottom: 14 }}>оплачено картой •• 4422</div>
        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
          <button className="rh-btn rh-btn--default rh-btn--sm">Маршрут</button>
          <button className="rh-btn rh-btn--primary rh-btn--sm">Открыть бронь</button>
        </div>
      </div>
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageClientBookings, ClientHeader });
