/* global React */
// Additional public/client web screens — search results, booking success.
const { Icon, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;

// ════════════════════════════════════════════════════════════════
// W · SEARCH RESULTS — split list + map
// ════════════════════════════════════════════════════════════════
function PageSearch() {
  return (
    <div className="rh-page" style={{ minHeight: 1500 }}>
      <TopNav active="catalogue" authed />

      {/* Persistent filter bar */}
      <div style={{
        padding: '14px 28px', display: 'flex', alignItems: 'center', gap: 10,
        background: 'rgba(255,253,247,0.92)', backdropFilter: 'blur(16px)',
        borderBottom: '1px solid rgba(15,23,42,0.06)', position: 'sticky', top: 0, zIndex: 3,
      }}>
        <div style={{
          display: 'inline-flex', alignItems: 'center', gap: 8,
          padding: '8px 14px', borderRadius: 999,
          background: '#fff', border: '1px solid rgba(15,23,42,0.10)',
          boxShadow: '0 6px 14px rgba(15,23,42,0.04)',
          fontSize: 13, fontWeight: 600, minWidth: 320,
        }}>
          <Icon name="search" size={15} style={{ color: '#5f6877' }} />
          <span>Москва · «вечер вдвоём»</span>
          <button style={{ marginLeft: 'auto', background: 'none', border: 0, color: '#5f6877', fontSize: 14, cursor: 'pointer' }}>×</button>
        </div>
        {[
          ['Сегодня · вечер', true],
          ['2 гостя', true],
          ['до 4 000 ₽/ч', false],
          ['Дровяная', false],
          ['С чаном', false],
          ['+ ещё 7', false],
        ].map(([t, on]) => (
          <span key={t} style={{
            padding: '7px 13px', borderRadius: 999, fontSize: 12, fontWeight: 700,
            background: on ? 'rgba(15,118,110,0.10)' : 'rgba(255,255,255,0.7)',
            border: on ? '1px solid rgba(15,118,110,0.28)' : '1px solid rgba(15,23,42,0.08)',
            color: on ? '#0a5f59' : '#16212b', cursor: 'pointer',
          }}>{t}</span>
        ))}
        <span style={{ flex: 1 }} />
        <span style={{ fontSize: 12, color: '#5f6877' }}>Сортировать:</span>
        <span style={{ padding: '7px 13px', borderRadius: 999, fontSize: 12, fontWeight: 700, background: '#fff', border: '1px solid rgba(15,23,42,0.10)' }}>Релевантность <Icon name="caret" size={12} style={{ verticalAlign: 'middle' }}/></span>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1.05fr 0.95fr', minHeight: 1200 }}>
        {/* List column */}
        <div style={{ padding: '24px 24px 28px 28px', overflowY: 'auto' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 18 }}>
            <h1 style={{ margin: 0, fontSize: 24, fontWeight: 800, letterSpacing: '-0.03em' }}>
              <span style={{ color: '#5f6877', fontWeight: 600 }}>Москва · </span>
              42 объекта
            </h1>
            <div style={{ display: 'inline-flex', padding: 3, background: 'rgba(15,23,42,0.06)', borderRadius: 999, fontSize: 12, fontWeight: 700 }}>
              {['Список', 'Карта', 'Сетка'].map((l, i) => (
                <span key={l} style={{
                  padding: '6px 14px', borderRadius: 999,
                  background: i === 0 ? '#fff' : 'transparent',
                  color: i === 0 ? '#16212b' : '#5f6877',
                  boxShadow: i === 0 ? '0 4px 10px rgba(15,23,42,0.06)' : 'none', cursor: 'pointer',
                }}>{l}</span>
              ))}
            </div>
          </div>

          <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
            {SEARCH_RESULTS.map((r, i) => <SearchListing key={i} r={r} />)}
          </div>

          <div style={{ display: 'flex', justifyContent: 'center', marginTop: 24 }}>
            <button className="rh-btn rh-btn--default">Показать ещё 36 объектов</button>
          </div>
        </div>

        {/* Map column */}
        <div style={{ position: 'relative', background: 'linear-gradient(180deg, #d6e4d3 0%, #c2d6c2 60%, #aabea3 100%)', overflow: 'hidden' }}>
          <MapMock />
        </div>
      </div>

      <Footer />
    </div>
  );
}

const SEARCH_RESULTS = [
  { name: 'Берёзовая роща', district: 'Рублёвское ш.', dist: '8 км', price: 3200, rating: 4.9, rev: 128, badge: 'Last min −20 %', tone: 'gold', tags: ['Дровяная', 'Чан', 'Бассейн'], variant: 'banya' },
  { name: 'Купеческая баня', district: 'Таганский', dist: '4 км', price: 2500, rating: 4.7, rev: 312, badge: 'Хит', tone: 'green', tags: ['Парная', 'Терраса', 'Шашлык'], variant: 'sand' },
  { name: 'Сосновый берег', district: 'Истринский', dist: '24 км', price: 4800, rating: 4.95, rev: 86, badge: 'Загород', tone: 'primary', tags: ['Лес', 'Купель', 'Бассейн'], variant: 'forest' },
  { name: 'Лофт «Пар»', district: 'Хамовники', dist: '5 км', price: 5400, rating: 4.8, rev: 56, tags: ['Дизайн', 'Хаммам', 'Бар'], variant: 'city' },
];

function SearchListing({ r }) {
  return (
    <article className="rh-card" style={{ padding: 14, borderRadius: 22, display: 'grid', gridTemplateColumns: '230px 1fr auto', gap: 18 }}>
      <div style={{ position: 'relative' }}>
        <PhotoPlaceholder height={170} radius={16} variant={r.variant} label="" />
        {r.badge && (
          <span style={{
            position: 'absolute', top: 10, left: 10,
            padding: '5px 10px', borderRadius: 999, fontSize: 11, fontWeight: 700,
            background: r.tone === 'gold' ? 'rgba(217,119,6,0.94)' : r.tone === 'green' ? 'rgba(21,128,61,0.94)' : 'rgba(15,118,110,0.94)',
            color: '#fffdf8',
          }}>{r.badge}</span>
        )}
        <div style={{ position: 'absolute', bottom: 10, left: 10, display: 'flex', gap: 4 }}>
          {[0, 1, 2, 3].map(i => <span key={i} style={{ width: 22, height: 4, borderRadius: 999, background: i === 0 ? '#fff' : 'rgba(255,255,255,0.4)' }}/>)}
        </div>
        <button style={{ position: 'absolute', top: 10, right: 10, width: 32, height: 32, borderRadius: 999, background: 'rgba(255,255,255,0.92)', border: 0, cursor: 'pointer' }}>♡</button>
      </div>
      <div>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <h3 style={{ margin: 0, fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em' }}>{r.name}</h3>
            <div style={{ fontSize: 12, color: '#5f6877', marginTop: 2 }}>
              <Icon name="pin" size={12} style={{ verticalAlign: '-2px', marginRight: 4 }} />
              Москва, {r.district} · {r.dist} от центра
            </div>
          </div>
          <Stars value={r.rating} count={r.rev} />
        </div>
        <div style={{ display: 'flex', gap: 6, marginTop: 10, flexWrap: 'wrap' }}>
          {r.tags.map(t => <span key={t} className="rh-tag">{t}</span>)}
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 12, fontSize: 12, color: '#5f6877' }}>
          <Icon name="check" size={14} style={{ color: '#15803d' }} /> Бесплатная отмена за 24 ч
          <span>·</span>
          <Icon name="bolt" size={14} style={{ color: '#0a5f59' }} /> Подтверждение моментально
        </div>
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'flex-end', justifyContent: 'space-between' }}>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em' }}>{r.price.toLocaleString('ru-RU')} ₽<span style={{ fontSize: 13, color: '#5f6877', fontWeight: 600 }}>/ч</span></div>
          <div style={{ fontSize: 11, color: '#5f6877' }}>от 3 ч · до 6 гостей</div>
        </div>
        <button className="rh-btn rh-btn--primary rh-btn--sm">Выбрать слот</button>
      </div>
    </article>
  );
}

// Schematic map with clustered pins
function MapMock() {
  return (
    <>
      {/* fake roads / blocks */}
      <svg width="100%" height="100%" style={{ position: 'absolute', inset: 0 }}>
        <defs>
          <pattern id="g" width="44" height="44" patternUnits="userSpaceOnUse">
            <path d="M0 0H44M0 0V44" stroke="rgba(255,255,255,0.18)" strokeWidth="1"/>
          </pattern>
        </defs>
        <rect width="100%" height="100%" fill="url(#g)"/>
        {/* river */}
        <path d="M0 360 Q140 320 240 380 T520 420 T780 360" stroke="rgba(120,180,210,0.62)" strokeWidth="22" fill="none" strokeLinecap="round"/>
        {/* parks */}
        <circle cx="180" cy="220" r="80" fill="rgba(160,200,140,0.6)"/>
        <ellipse cx="540" cy="200" rx="100" ry="60" fill="rgba(160,200,140,0.55)"/>
        <ellipse cx="380" cy="640" rx="120" ry="80" fill="rgba(160,200,140,0.5)"/>
        {/* roads */}
        <path d="M0 240 L780 240" stroke="rgba(255,255,255,0.6)" strokeWidth="3"/>
        <path d="M0 540 L780 540" stroke="rgba(255,255,255,0.4)" strokeWidth="2"/>
        <path d="M380 0 L380 800" stroke="rgba(255,255,255,0.4)" strokeWidth="2"/>
      </svg>

      {/* You-are-here */}
      <div style={{ position: 'absolute', left: '40%', top: '52%', width: 14, height: 14, borderRadius: '50%', background: '#2563eb', border: '3px solid #fff', boxShadow: '0 0 0 6px rgba(37,99,235,0.20)' }} />

      {/* Price pins */}
      {[
        { x: '24%', y: '28%', p: '3 200', sel: false },
        { x: '52%', y: '32%', p: '2 500', sel: true },
        { x: '68%', y: '46%', p: '4 800', sel: false },
        { x: '36%', y: '56%', p: '5 400', sel: false },
        { x: '60%', y: '68%', p: '2 900', sel: false },
        { x: '20%', y: '70%', p: '3 600', sel: false },
        { x: '78%', y: '24%', p: '4 100', sel: false },
        { x: '44%', y: '78%', p: '2 200', sel: false },
      ].map((p, i) => (
        <div key={i} style={{
          position: 'absolute', left: p.x, top: p.y,
          padding: '6px 12px', borderRadius: 999,
          background: p.sel ? '#16212b' : '#fffdf8',
          color: p.sel ? '#fffdf8' : '#16212b',
          fontSize: 13, fontWeight: 800, letterSpacing: '-0.01em',
          boxShadow: p.sel ? '0 8px 18px rgba(15,23,42,0.30)' : '0 6px 14px rgba(15,23,42,0.18)',
          transform: 'translate(-50%, -100%)',
          border: p.sel ? 0 : '1px solid rgba(15,23,42,0.10)',
        }}>{p.p} ₽</div>
      ))}

      {/* Cluster */}
      <div style={{
        position: 'absolute', right: '12%', bottom: '20%',
        width: 56, height: 56, borderRadius: 999,
        background: 'rgba(15,118,110,0.94)', color: '#fffdf8',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        fontSize: 16, fontWeight: 800, boxShadow: '0 0 0 6px rgba(15,118,110,0.20), 0 12px 26px rgba(15,118,110,0.36)',
        transform: 'translate(-50%, -50%)',
      }}>+12</div>

      {/* Map controls */}
      <div style={{ position: 'absolute', top: 18, right: 18, display: 'flex', flexDirection: 'column', gap: 8 }}>
        {[<Icon name="plus" size={18}/>, <Icon name="minus" size={18}/>, <Icon name="map" size={18}/>].map((c, i) => (
          <button key={i} style={{ width: 38, height: 38, borderRadius: 12, background: '#fff', border: '1px solid rgba(15,23,42,0.10)', boxShadow: '0 6px 14px rgba(15,23,42,0.10)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', color: '#16212b', cursor: 'pointer' }}>{c}</button>
        ))}
      </div>

      {/* Bottom rebuild btn */}
      <div style={{ position: 'absolute', bottom: 20, left: '50%', transform: 'translateX(-50%)', display: 'inline-flex', alignItems: 'center', gap: 8, padding: '8px 18px', borderRadius: 999, background: '#16212b', color: '#fffdf8', fontSize: 13, fontWeight: 700, boxShadow: '0 12px 28px rgba(15,23,42,0.32)' }}>
        <Icon name="search" size={14} /> Искать в этой области
      </div>
    </>
  );
}

// ════════════════════════════════════════════════════════════════
// W · BOOKING SUCCESS (e-ticket)
// ════════════════════════════════════════════════════════════════
function PageBookingSuccess() {
  return (
    <div className="rh-page" style={{ minHeight: 1100 }}>
      <TopNav active="catalogue" authed />

      <Section padding="56px 28px 28px">
        <div className="rh-card" style={{ maxWidth: 920, margin: '0 auto', padding: 0, borderRadius: 36, overflow: 'hidden', display: 'grid', gridTemplateColumns: '1fr 360px' }}>
          <div style={{ padding: 40 }}>
            <div style={{
              display: 'inline-flex', alignItems: 'center', gap: 10,
              padding: '6px 14px', borderRadius: 999,
              background: 'rgba(21,128,61,0.10)', border: '1px solid rgba(21,128,61,0.24)',
              color: '#15803d', fontSize: 12, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase',
            }}>
              <Icon name="checkc" size={14}/> Бронь подтверждена
            </div>

            <h1 style={{ margin: '20px 0 10px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.05 }}>
              До встречи в субботу, Анна.
            </h1>
            <p style={{ margin: 0, fontSize: 15, color: '#5f6877', lineHeight: 1.55 }}>
              Бронь #RHB-24018 в «Берёзовой роще» подтверждена. Мы отправили детали на <b style={{ color: '#16212b' }}>anna@example.ru</b> и в Telegram.
            </p>

            <div style={{
              display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14,
              marginTop: 26, padding: 20, borderRadius: 22,
              background: 'linear-gradient(180deg, rgba(255,253,247,0.96), rgba(245,238,224,0.84))',
              border: '1px dashed rgba(15,23,42,0.18)',
            }}>
              {[
                ['Когда', '20 апреля, СБ\n19:00 – 23:00 (4 ч)'],
                ['Гости', '4 + 1 ребёнок\nконтакт +7 (910) 555-24-18'],
                ['Где', 'Москва, Рублёвское ш., 12\nкорпус «Сосна»'],
                ['Хозяин', 'Иван Соколов\nответит в течение 5 минут'],
              ].map(([k, v]) => (
                <div key={k}>
                  <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>{k}</div>
                  <div style={{ marginTop: 4, fontSize: 14, fontWeight: 600, whiteSpace: 'pre-line', lineHeight: 1.4 }}>{v}</div>
                </div>
              ))}
            </div>

            <div style={{ marginTop: 22, padding: '14px 16px', borderRadius: 14, background: 'rgba(217,119,6,0.06)', border: '1px solid rgba(217,119,6,0.20)', fontSize: 13, color: '#92400e', display: 'flex', alignItems: 'flex-start', gap: 10 }}>
              <Icon name="bolt" size={16} style={{ marginTop: 1, color: '#d97706' }} />
              <span><b>Что важно знать:</b> отмена бесплатно до 18 апреля, 19:00. Дальше — 50 % стоимости. Полный регламент на странице брони.</span>
            </div>

            <div style={{ display: 'flex', gap: 10, marginTop: 24, flexWrap: 'wrap' }}>
              <button className="rh-btn rh-btn--primary"><Icon name="msg" size={14}/> Чат с хозяином</button>
              <button className="rh-btn rh-btn--default"><Icon name="cal" size={14}/> В календарь</button>
              <button className="rh-btn rh-btn--default"><Icon name="download" size={14}/> Скачать билет</button>
              <button className="rh-btn rh-btn--ghost rh-btn--sm">Поделиться</button>
            </div>
          </div>

          {/* QR side */}
          <aside style={{
            padding: 32, color: '#fffdf8',
            background: 'linear-gradient(160deg, #10313a 0%, #38606a 60%, #9a5c30 100%)',
            display: 'flex', flexDirection: 'column', gap: 16, position: 'relative', overflow: 'hidden',
          }}>
            <div style={{ position: 'absolute', inset: 0, opacity: 0.20 }}>
              <svg width="100%" height="100%"><circle cx="80%" cy="20%" r="120" stroke="#fff" strokeWidth="1" fill="none"/><circle cx="20%" cy="86%" r="160" stroke="#fff" strokeWidth="1" fill="none"/></svg>
            </div>
            <div style={{ position: 'relative' }}>
              <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>Электронный билет</div>
              <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em', marginTop: 4 }}>RHB-24018</div>
            </div>
            <div style={{ position: 'relative', background: '#fffdf8', padding: 16, borderRadius: 20, color: '#16212b' }}>
              <FakeQR />
              <div style={{ marginTop: 10, fontSize: 11, color: '#5f6877', textAlign: 'center', letterSpacing: '0.12em', fontWeight: 700 }}>Покажите на ресепшн</div>
            </div>
            <div style={{ position: 'relative', display: 'flex', justifyContent: 'space-between', fontSize: 12, color: 'rgba(255,255,255,0.74)' }}>
              <div>
                <div style={{ fontWeight: 700, color: '#fffdf8' }}>Заезд</div>
                <div>СБ · 18:45</div>
              </div>
              <div style={{ textAlign: 'right' }}>
                <div style={{ fontWeight: 700, color: '#fffdf8' }}>Сумма</div>
                <div>10 840 ₽</div>
              </div>
            </div>
          </aside>
        </div>

        {/* Suggested */}
        <div style={{ maxWidth: 920, margin: '36px auto 0' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <h2 className="rh-h2">Пока ждёте — добавьте к вечеру</h2>
            <a href="#" style={{ color: '#0a5f59', fontWeight: 700, fontSize: 13, textDecoration: 'none' }}>Все add-ons →</a>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 14, marginTop: 14 }}>
            {[
              ['Шашлычный сет', 'Свинина, овощи, лаваш на 5 чел.', '2 400 ₽', 'sand'],
              ['Чан с травами', 'Тёплый, 6 чел., 2 ч', '1 800 ₽', 'forest'],
              ['Массаж 60 мин', 'На месте, лицензированный мастер', '4 200 ₽', 'spa'],
            ].map(([t, d, p, v]) => (
              <article key={t} className="rh-card" style={{ padding: 0, borderRadius: 20, overflow: 'hidden' }}>
                <PhotoPlaceholder height={120} radius={0} variant={v} label="" />
                <div style={{ padding: 14 }}>
                  <div style={{ fontSize: 14, fontWeight: 800 }}>{t}</div>
                  <div style={{ fontSize: 12, color: '#5f6877', marginTop: 2 }}>{d}</div>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 10 }}>
                    <span style={{ fontWeight: 800 }}>{p}</span>
                    <button className="rh-btn rh-btn--default rh-btn--sm"><Icon name="plus" size={12}/> Добавить</button>
                  </div>
                </div>
              </article>
            ))}
          </div>
        </div>
      </Section>

      <Footer />
    </div>
  );
}

function FakeQR() {
  // Schematic 21x21 QR — looks like a real QR without encoding anything
  const cells = [];
  const seed = 'RHB24018RELAXHUB2026SOKOLOV';
  for (let r = 0; r < 21; r++) {
    for (let c = 0; c < 21; c++) {
      const corner = (r < 7 && c < 7) || (r < 7 && c > 13) || (r > 13 && c < 7);
      const innerCorner = (r >= 1 && r <= 5 && c >= 1 && c <= 5) || (r >= 1 && r <= 5 && c >= 15 && c <= 19) || (r >= 15 && r <= 19 && c >= 1 && c <= 5);
      const innerDot = (r >= 2 && r <= 4 && c >= 2 && c <= 4) || (r >= 2 && r <= 4 && c >= 16 && c <= 18) || (r >= 16 && r <= 18 && c >= 2 && c <= 4);
      let on;
      if (innerDot) on = true;
      else if (innerCorner) on = false;
      else if (corner) on = true;
      else {
        const ch = seed.charCodeAt((r * 21 + c) % seed.length);
        on = ((ch + r * 3 + c * 5) % 7) > 3;
      }
      cells.push(<rect key={`${r}-${c}`} x={c * 8} y={r * 8} width="8" height="8" fill={on ? '#16212b' : 'transparent'} />);
    }
  }
  return (
    <svg viewBox="0 0 168 168" width="100%" style={{ display: 'block' }}>
      {cells}
    </svg>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageSearch, PageBookingSuccess });
