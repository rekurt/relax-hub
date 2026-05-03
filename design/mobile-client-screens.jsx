/* global React */
const { MIcon, MPage, MCard, MHeader, MTag, MButton, MPhoto, MAvatar, MTabBar, MStars } = window.RHM;

const CLIENT_TABS = [
  ['home', 'Главная', 'home'],
  ['search', 'Поиск', 'search'],
  ['bookings', 'Брони', 'cal'],
  ['chat', 'Чат', 'msg'],
  ['profile', 'Профиль', 'user'],
];

// ════════════════════════════════════════════════════════════
// CLIENT · 01 HOME
// ════════════════════════════════════════════════════════════
function MClientHome() {
  return (
    <MPage>
      <MHeader
        eyebrow="14 апреля · вечер свободен"
        title={<>Привет, Анна.<br/>Куда сегодня?</>}
        sub="Каталог, понятные цены и SMS-подтверждение."
        right={<MAvatar name="Анна Поляк" size={42} tone="teal" />}
      />

      {/* Quick search */}
      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={6} radius={28} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '0 14px', flex: 1 }}>
            <MIcon name="search" size={18} color="#5f6877" />
            <span style={{ fontSize: 14, color: '#5f6877' }}>Москва · сегодня · 4 гостя</span>
          </div>
          <MButton variant="primary" size="sm" style={{ height: 40 }}><MIcon name="filter" size={14} /> Фильтр</MButton>
        </MCard>
      </div>

      {/* Scenarios */}
      <div style={{ padding: '8px 20px 14px' }}>
        <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Сценарии</span>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginTop: 10 }}>
          {[
            ['Вечер вдвоём', 'spa', 'spa'],
            ['Компания 6–10', 'banya', 'fire'],
            ['Семья', 'forest', 'shield'],
            ['Last minute', 'sand', 'bolt'],
          ].map(([title, variant, icon]) => (
            <MCard key={title} padding={0} radius={20} style={{ overflow: 'hidden' }}>
              <MPhoto height={84} variant={variant} radius={0} />
              <div style={{ padding: '10px 12px', display: 'flex', alignItems: 'center', gap: 8 }}>
                <div style={{ width: 28, height: 28, borderRadius: 10, background: 'rgba(15,118,110,0.10)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <MIcon name={icon} size={14} />
                </div>
                <span style={{ fontSize: 13, fontWeight: 700 }}>{title}</span>
              </div>
            </MCard>
          ))}
        </div>
      </div>

      {/* Featured listing */}
      <div style={{ padding: '8px 20px 14px' }}>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 10 }}>
          <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Часто бронируют</span>
          <span style={{ fontSize: 12, fontWeight: 700, color: '#0a5f59', display: 'inline-flex', alignItems: 'center', gap: 4 }}>Все 124 <MIcon name="arrow" size={12} /></span>
        </div>
        <MCard padding={0} radius={22} style={{ overflow: 'hidden' }}>
          <div style={{ position: 'relative' }}>
            <MPhoto height={160} variant="banya" radius={0} />
            <div style={{ position: 'absolute', top: 10, left: 10, display: 'flex', gap: 6 }}>
              <MTag tone="red" style={{ background: 'rgba(180,35,24,0.92)', color: '#fff', border: 'none' }}><MIcon name="bolt" size={10} /> Last minute -20%</MTag>
            </div>
            <div style={{ position: 'absolute', top: 10, right: 10, width: 36, height: 36, borderRadius: 999, background: 'rgba(255,255,255,0.94)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <MIcon name="heart" size={16} color="#16212b" />
            </div>
          </div>
          <div style={{ padding: 14 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 4 }}>
              <span style={{ fontSize: 16, fontWeight: 700 }}>Берёзовая роща</span>
              <span style={{ fontSize: 14, fontWeight: 800 }}>3 200 ₽<span style={{ fontSize: 11, color: '#5f6877', fontWeight: 500 }}>/ч</span></span>
            </div>
            <div style={{ fontSize: 12, color: '#5f6877', display: 'inline-flex', alignItems: 'center', gap: 4, marginBottom: 8 }}>
              <MIcon name="pin" size={12} /> Москва, Рублёвское ш., 12 · 18 км
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <MStars value={4.9} count={128} />
              <div style={{ display: 'flex', gap: 4 }}>
                <MTag>Сауна</MTag>
                <MTag>Парная</MTag>
                <MTag>Чан</MTag>
              </div>
            </div>
          </div>
        </MCard>
      </div>

      {/* Loyalty strip */}
      <div style={{ padding: '8px 20px 0' }}>
        <MCard padding={14} radius={20} style={{ display: 'flex', alignItems: 'center', gap: 12, background: 'linear-gradient(135deg, rgba(217,119,6,0.10), rgba(255,255,255,0.86))' }}>
          <div style={{ width: 40, height: 40, borderRadius: 12, background: 'rgba(217,119,6,0.16)', color: '#92400e', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <MIcon name="fire" size={20} />
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 13, fontWeight: 700 }}>Уровень «Завсегдатай»</div>
            <div style={{ fontSize: 11, color: '#5f6877' }}>3 брони до уровня «Хозяин пара» · +5%</div>
          </div>
          <MIcon name="arrow" size={16} color="#5f6877" />
        </MCard>
      </div>

      <MTabBar items={CLIENT_TABS} active="home" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// CLIENT · 02 SEARCH (map + cards stack)
// ════════════════════════════════════════════════════════════
function MClientSearch() {
  return (
    <MPage>
      {/* Sticky search */}
      <div style={{ padding: '6px 16px 10px', position: 'sticky', top: 0, zIndex: 5,
        background: 'linear-gradient(180deg, rgba(248,244,236,0.96), rgba(248,244,236,0))',
      }}>
        <MCard padding={6} radius={26} style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
          <button style={{ width: 40, height: 40, borderRadius: 999, border: 0, background: 'transparent', color: '#16212b' }}>
            <MIcon name="arrl" size={18} />
          </button>
          <span style={{ flex: 1, fontSize: 14, fontWeight: 600 }}>Москва · 20 апр · 4 гостя</span>
          <button style={{ width: 40, height: 40, borderRadius: 999, border: 0, background: '#16212b', color: '#fffdf8' }}>
            <MIcon name="map" size={16} color="#fffdf8" />
          </button>
        </MCard>
        <div style={{ display: 'flex', gap: 6, overflowX: 'auto', paddingTop: 10 }}>
          <MTag tone="primary"><MIcon name="filter" size={12} /> Фильтры (3)</MTag>
          <MTag>2 500–4 000 ₽</MTag>
          <MTag>Сауна</MTag>
          <MTag>Бассейн</MTag>
          <MTag>Чан</MTag>
        </div>
      </div>

      {/* Map peek */}
      <div style={{ padding: '0 16px 14px', position: 'relative' }}>
        <div style={{
          height: 180, borderRadius: 22, overflow: 'hidden',
          background: 'linear-gradient(135deg, #d7e6e3 0%, #f0e6d4 60%, #ead7c0 100%)',
          position: 'relative', border: '1px solid rgba(15,23,42,0.08)',
        }}>
          <svg width="100%" height="100%" style={{ position: 'absolute', inset: 0 }}>
            <defs>
              <pattern id="map-grid" width="40" height="40" patternUnits="userSpaceOnUse">
                <path d="M 40 0 L 0 0 0 40" fill="none" stroke="rgba(15,23,42,0.06)" strokeWidth="1"/>
              </pattern>
            </defs>
            <rect width="100%" height="100%" fill="url(#map-grid)" />
            <path d="M 40 80 Q 100 60 160 100 T 320 60" stroke="rgba(15,118,110,0.30)" strokeWidth="3" fill="none"/>
          </svg>
          {[
            { x: '25%', y: '40%', p: '3 200', on: false },
            { x: '52%', y: '30%', p: '2 500', on: true },
            { x: '70%', y: '60%', p: '4 800', on: false },
            { x: '40%', y: '70%', p: '3 600', on: false },
          ].map((m, i) => (
            <div key={i} style={{
              position: 'absolute', left: m.x, top: m.y, transform: 'translate(-50%, -100%)',
              padding: '4px 10px', borderRadius: 999,
              background: m.on ? '#16212b' : '#fffdf8',
              color: m.on ? '#fffdf8' : '#16212b',
              fontSize: 11, fontWeight: 800,
              boxShadow: '0 6px 14px rgba(15,23,42,0.16)',
              border: '1px solid rgba(15,23,42,0.10)',
            }}>{m.p} ₽</div>
          ))}
        </div>
      </div>

      {/* Result list */}
      <div style={{ padding: '0 16px 14px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
          <span style={{ fontSize: 13, fontWeight: 700 }}>124 объекта</span>
          <span style={{ fontSize: 12, color: '#5f6877', display: 'inline-flex', alignItems: 'center', gap: 4 }}>По умолчанию <MIcon name="caret" size={12} /></span>
        </div>
        {[
          { name: 'Купеческая', addr: 'Москва, Таганский', price: '2 500', rating: 4.7, reviews: 312, badge: 'Хит', variant: 'sand' },
          { name: 'Берёзовая роща', addr: 'Москва, Рублёвское ш.', price: '3 200', rating: 4.9, reviews: 128, badge: '⚡ Last min', variant: 'banya' },
          { name: 'Лофт «Пар»', addr: 'СПб, Васильевский', price: '3 600', rating: 4.9, reviews: 84, badge: 'Новое', variant: 'city' },
        ].map(l => (
          <MCard key={l.name} padding={10} radius={20} style={{ display: 'grid', gridTemplateColumns: '92px 1fr auto', gap: 12, marginBottom: 10 }}>
            <MPhoto height={92} variant={l.variant} radius={14} style={{ width: 92 }} />
            <div style={{ minWidth: 0 }}>
              <div style={{ display: 'flex', gap: 6, marginBottom: 4 }}>
                <MTag tone={l.badge.includes('⚡') ? 'red' : l.badge === 'Хит' ? 'green' : 'gold'}>{l.badge}</MTag>
              </div>
              <div style={{ fontSize: 14, fontWeight: 700 }}>{l.name}</div>
              <div style={{ fontSize: 11, color: '#5f6877', display: 'inline-flex', gap: 3, alignItems: 'center', marginTop: 2 }}>
                <MIcon name="pin" size={10} /> {l.addr}
              </div>
              <div style={{ marginTop: 6 }}><MStars value={l.rating} count={l.reviews} /></div>
            </div>
            <div style={{ textAlign: 'right', display: 'flex', flexDirection: 'column', justifyContent: 'space-between', alignItems: 'flex-end' }}>
              <MIcon name="heart" size={16} color="#5f6877" />
              <div>
                <div style={{ fontSize: 14, fontWeight: 800 }}>{l.price} ₽</div>
                <div style={{ fontSize: 10, color: '#5f6877' }}>/час</div>
              </div>
            </div>
          </MCard>
        ))}
      </div>

      <MTabBar items={CLIENT_TABS} active="search" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// CLIENT · 03 LISTING DETAIL
// ════════════════════════════════════════════════════════════
function MClientListing() {
  return (
    <MPage style={{ paddingBottom: 120 }}>
      {/* Hero photo */}
      <div style={{ position: 'relative' }}>
        <MPhoto height={300} variant="banya" radius={0} label="" />
        <div style={{ position: 'absolute', top: 56, left: 16, right: 16, display: 'flex', justifyContent: 'space-between' }}>
          <div style={{ width: 40, height: 40, borderRadius: 999, background: 'rgba(255,255,255,0.94)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <MIcon name="arrl" size={18} />
          </div>
          <div style={{ display: 'flex', gap: 8 }}>
            {['heart','msg'].map(n => (
              <div key={n} style={{ width: 40, height: 40, borderRadius: 999, background: 'rgba(255,255,255,0.94)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <MIcon name={n} size={18} />
              </div>
            ))}
          </div>
        </div>
        <div style={{ position: 'absolute', bottom: 14, left: 16, display: 'flex', gap: 6 }}>
          <MTag tone="red" style={{ background: 'rgba(180,35,24,0.92)', color: '#fff', border: 'none' }}><MIcon name="bolt" size={10} /> Last minute -20%</MTag>
          <MTag tone="dark">1 / 12</MTag>
        </div>
      </div>

      <div style={{ padding: '18px 20px 0' }}>
        <h1 style={{ margin: '0 0 6px', fontSize: 26, fontWeight: 800, letterSpacing: '-0.04em' }}>Берёзовая роща</h1>
        <div style={{ display: 'flex', gap: 12, alignItems: 'center', flexWrap: 'wrap', marginBottom: 14 }}>
          <MStars value={4.9} count={128} />
          <span style={{ fontSize: 12, color: '#5f6877', display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            <MIcon name="pin" size={12} /> Москва, Рублёвское ш., 12
          </span>
        </div>

        <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', marginBottom: 16 }}>
          <MTag tone="green"><MIcon name="checkc" size={10} /> Проверено</MTag>
          <MTag>Сауна</MTag><MTag>Парная</MTag><MTag>Чан</MTag><MTag>Бассейн</MTag><MTag>Бар</MTag>
        </div>

        <MCard padding={16} radius={22} style={{ marginBottom: 14 }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 10 }}>
            <MAvatar name="Иван Соколов" size={42} tone="amber" />
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 14, fontWeight: 700 }}>Иван Соколов</div>
              <div style={{ fontSize: 11, color: '#5f6877' }}>Хозяин · отвечает за 12 минут</div>
            </div>
            <MButton variant="default" size="sm"><MIcon name="msg" size={14} /></MButton>
          </div>
        </MCard>

        <h3 style={{ margin: '0 0 8px', fontSize: 16, fontWeight: 800 }}>Об объекте</h3>
        <p style={{ margin: '0 0 16px', fontSize: 14, color: '#5f6877', lineHeight: 1.55 }}>
          Приватная баня на дровах с двумя парными, бассейном 4×8 м и зоной отдыха на 8 человек.
          Бронирование по часам, минимум 4 часа. Чай и берёзовые веники в подарок.
        </p>

        <h3 style={{ margin: '0 0 10px', fontSize: 16, fontWeight: 800 }}>Свободно сегодня</h3>
        <div style={{ display: 'flex', gap: 8, overflowX: 'auto', marginBottom: 16, paddingBottom: 4 }}>
          {['10:00','12:00','14:00','16:00','18:00','20:00'].map((t, i) => (
            <div key={t} style={{
              padding: '10px 14px', borderRadius: 14,
              background: i === 4 ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'rgba(255,255,255,0.94)',
              color: i === 4 ? '#fffdf8' : '#16212b',
              border: '1px solid rgba(15,23,42,0.08)',
              fontSize: 13, fontWeight: 700, fontVariantNumeric: 'tabular-nums',
              boxShadow: i === 4 ? '0 8px 16px rgba(15,118,110,0.22)' : 'none',
            }}>{t}</div>
          ))}
        </div>
      </div>

      {/* Sticky CTA */}
      <div style={{
        position: 'fixed', left: 16, right: 16, bottom: 20,
        display: 'flex', gap: 10, alignItems: 'center',
        padding: 10, borderRadius: 999,
        background: 'rgba(255, 252, 247, 0.94)',
        backdropFilter: 'blur(20px)', WebkitBackdropFilter: 'blur(20px)',
        border: '1px solid rgba(15,23,42,0.08)',
        boxShadow: '0 14px 36px rgba(15,23,42,0.16)',
        zIndex: 30,
      }}>
        <div style={{ flex: 1, paddingLeft: 14 }}>
          <div style={{ fontSize: 16, fontWeight: 800 }}>3 200 ₽<span style={{ fontSize: 12, color: '#5f6877', fontWeight: 500 }}>/ч · от 4 ч</span></div>
          <div style={{ fontSize: 11, color: '#5f6877' }}>Сб, 20 апр · 18:00–22:00</div>
        </div>
        <MButton variant="primary">Забронировать <MIcon name="arrow" size={14} /></MButton>
      </div>
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// CLIENT · 04 BOOKINGS
// ════════════════════════════════════════════════════════════
function MClientBookings() {
  return (
    <MPage>
      <MHeader eyebrow="Кабинет гостя" title="Бронирования" sub="Все ваши предстоящие визиты в одном списке." />
      <div style={{ padding: '0 20px 14px' }}>
        <div style={{ display: 'inline-flex', padding: 4, borderRadius: 999, background: 'rgba(255,255,255,0.7)', border: '1px solid rgba(15,23,42,0.08)' }}>
          {[['Активные', true], ['История', false], ['Сертификаты', false]].map(([l, on]) => (
            <button key={l} style={{
              padding: '8px 14px', borderRadius: 999, border: 0, background: on ? '#fff' : 'transparent',
              boxShadow: on ? '0 4px 10px rgba(15,23,42,0.06)' : 'none',
              fontSize: 12, fontWeight: 700, color: on ? '#16212b' : '#5f6877',
            }}>{l}</button>
          ))}
        </div>
      </div>

      <div style={{ padding: '0 20px 14px', display: 'flex', flexDirection: 'column', gap: 12 }}>
        {[
          { variant: 'banya', name: 'Берёзовая роща', when: 'Сб, 20 апр · 19:00–23:00', total: '11 440 ₽', status: 'confirmed', code: 'RH-4821' },
          { variant: 'forest', name: 'Сосновый берег', when: 'Сб, 4 мая · 16:00–20:00', total: '22 200 ₽', status: 'awaiting', code: 'RH-4860' },
        ].map(b => (
          <MCard key={b.code} padding={0} radius={22} style={{ overflow: 'hidden' }}>
            <MPhoto height={120} variant={b.variant} radius={0} label="" />
            <div style={{ padding: 14 }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginBottom: 6 }}>
                <MTag tone={b.status === 'confirmed' ? 'green' : 'gold'}>
                  {b.status === 'confirmed' ? 'Подтверждено' : 'Ждёт хозяина'}
                </MTag>
                <span style={{ fontSize: 10, fontWeight: 700, letterSpacing: '0.10em', color: '#5f6877' }}>{b.code}</span>
              </div>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                <h3 style={{ margin: 0, fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em' }}>{b.name}</h3>
                <span style={{ fontSize: 16, fontWeight: 800 }}>{b.total}</span>
              </div>
              <div style={{ fontSize: 12, color: '#5f6877', marginTop: 4 }}>{b.when}</div>
              <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
                <MButton variant="default" size="sm"><MIcon name="map" size={12} /> Маршрут</MButton>
                <MButton variant="default" size="sm"><MIcon name="qr" size={12} /> QR-код</MButton>
                <MButton variant="primary" size="sm" style={{ marginLeft: 'auto' }}>Открыть <MIcon name="arrow" size={12} /></MButton>
              </div>
            </div>
          </MCard>
        ))}
      </div>

      <MTabBar items={CLIENT_TABS} active="bookings" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// CLIENT · 05 PROFILE
// ════════════════════════════════════════════════════════════
function MClientProfile() {
  return (
    <MPage>
      <MHeader eyebrow="Профиль" title="Анна Поляк" sub="anna.polyak@mail.ru · +7 999 ··· 12 34" right={<MIcon name="set" size={22} color="#5f6877" />} />

      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={20} radius={26} style={{ background: 'linear-gradient(135deg,#10313a,#38606a 60%, #9a5c30)', color: '#fffdf8', border: 'none' }}>
          <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.12em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>RH-кошелёк</div>
          <div style={{ fontSize: 30, fontWeight: 800, letterSpacing: '-0.04em', marginTop: 6 }}>4 280 ₽</div>
          <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.74)' }}>+ 540 ₽ за апрель</div>
          <div style={{ display: 'flex', gap: 8, marginTop: 14 }}>
            <MButton variant="default" size="sm" style={{ background: '#fffdf8' }}><MIcon name="plus" size={12} /> Пополнить</MButton>
            <MButton variant="ghost" size="sm" style={{ color: '#fffdf8', background: 'rgba(255,255,255,0.14)' }}>История</MButton>
          </div>
        </MCard>
      </div>

      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={14} radius={22}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
            <MIcon name="fire" size={20} color="#d97706" />
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 14, fontWeight: 800 }}>Завсегдатай</div>
              <div style={{ fontSize: 11, color: '#5f6877' }}>3 брони до уровня «Хозяин пара»</div>
            </div>
            <span style={{ fontSize: 13, fontWeight: 800, color: '#0a5f59' }}>+5%</span>
          </div>
          <div style={{ height: 8, borderRadius: 999, background: 'rgba(15,23,42,0.06)', overflow: 'hidden' }}>
            <div style={{ width: '62%', height: '100%', background: 'linear-gradient(90deg,#d97706,#b45309)' }} />
          </div>
        </MCard>
      </div>

      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={6} radius={22}>
          {[
            ['heart', 'Избранное', '12 объектов'],
            ['gift', 'Сертификаты', '2 активных'],
            ['card', 'Карты и оплата', '•• 4422'],
            ['shield', 'Безопасность', '2FA включено'],
            ['msg', 'Поддержка', 'Ежедневно 09–22'],
          ].map(([ic, l, sub], i, arr) => (
            <div key={l} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 12px', borderTop: i ? '1px solid rgba(15,23,42,0.06)' : '0' }}>
              <div style={{ width: 36, height: 36, borderRadius: 12, background: 'rgba(15,118,110,0.08)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <MIcon name={ic} size={16} />
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, fontWeight: 700 }}>{l}</div>
                <div style={{ fontSize: 11, color: '#5f6877' }}>{sub}</div>
              </div>
              <MIcon name="arrow" size={14} color="#5f6877" />
            </div>
          ))}
        </MCard>
      </div>

      <MTabBar items={CLIENT_TABS} active="profile" />
    </MPage>
  );
}

window.RHM = window.RHM || {};
Object.assign(window.RHM, { MClientHome, MClientSearch, MClientListing, MClientBookings, MClientProfile });
