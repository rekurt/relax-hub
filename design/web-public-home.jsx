/* global React */
const { Icon, BrandLockup, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;

// ════════════════════════════════════════════════════════════════
// PUBLIC · 01 HOME — Hero, scenarios, featured cards
// ════════════════════════════════════════════════════════════════
function PageHome() {
  return (
    <div className="rh-page" style={{ minHeight: 1700 }}>
      <TopNav active="home" />

      {/* HERO */}
      <Section padding="56px 28px 24px">
        <div style={{ display: 'grid', gridTemplateColumns: '1.18fr 0.82fr', gap: 36, alignItems: 'center' }}>
          <div>
            <span className="rh-eyebrow" style={{ fontSize: 12 }}>Discovery без шума</span>
            <h1 style={{
              margin: '14px 0 18px',
              fontSize: 56, lineHeight: 1.02, fontWeight: 800,
              letterSpacing: '-0.05em',
            }}>
              Бани для вечера вдвоём,<br/>компании и выходных<br/>за городом.
            </h1>
            <p style={{ fontSize: 17, lineHeight: 1.6, color: '#5f6877', maxWidth: 560, fontWeight: 500 }}>
              Каталог проверенных мест, понятные цены и SMS-подтверждение собраны
              в один короткий маршрут до бронирования.
            </p>

            {/* Search bar */}
            <div className="rh-card" style={{
              marginTop: 30, padding: 14,
              display: 'grid',
              gridTemplateColumns: '1.4fr 1fr 1fr 1fr auto',
              gap: 10, alignItems: 'center',
              borderRadius: 28,
            }}>
              <SearchField label="Город" value="Москва" icon="pin" />
              <SearchField label="Дата" value="Сб, 20 апреля" icon="cal" />
              <SearchField label="Время" value="19:00 – 23:00" icon="clock" />
              <SearchField label="Гостей" value="4 гостя" icon="user" />
              <button className="rh-btn rh-btn--primary rh-btn--lg" style={{ padding: '0 22px' }}>
                <Icon name="search" size={18} /> Подобрать
              </button>
            </div>

            <div style={{ display: 'flex', gap: 18, marginTop: 22, color: '#5f6877', fontSize: 13, fontWeight: 600 }}>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Icon name="checkc" size={16} style={{ color: '#15803d' }} /> 0 % комиссии для владельцев</span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Icon name="shield" size={16} style={{ color: '#0f766e' }} /> Возврат до 24 часов</span>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Icon name="bolt" size={16} style={{ color: '#d97706' }} /> Last minute -20 %</span>
            </div>
          </div>

          {/* Hero card stack */}
          <div style={{ position: 'relative', height: 460 }}>
            <PhotoPlaceholder height={460} radius={32} variant="banya" label="ФОТО · ПЕРВЫЙ ЭКРАН" />
            <div className="rh-card" style={{
              position: 'absolute', left: -30, bottom: 28, width: 280,
              padding: 16, borderRadius: 24,
            }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span style={{ fontSize: 15, fontWeight: 700 }}>Берёзовая роща</span>
                <Stars value={4.9} count={128} />
              </div>
              <p style={{ margin: '6px 0 12px', fontSize: 12, color: '#5f6877' }}>Москва, Рублёвское ш., 12</p>
              <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                <span className="rh-tag rh-tag--gold">⚡ Last minute -20%</span>
                <span className="rh-tag">Сауна</span><span className="rh-tag">Бассейн</span>
              </div>
            </div>
            <div className="rh-card" style={{
              position: 'absolute', right: -22, top: 28, padding: '14px 18px',
              borderRadius: 22, display: 'flex', gap: 12, alignItems: 'center',
            }}>
              <Avatar name="М П" size={40} tone="amber" />
              <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.2 }}>
                <span style={{ fontSize: 13, fontWeight: 700 }}>Михаил П.</span>
                <span style={{ fontSize: 11, color: '#5f6877' }}>забронировал · 3 мин назад</span>
              </div>
            </div>
          </div>
        </div>
      </Section>

      {/* SCENARIOS */}
      <Section padding="44px 28px">
        <SectionHeader eyebrow="Сценарии" title="Не «найди баню» — «найди вечер»" />
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 18, marginTop: 28 }}>
          {[
            ['Вечер вдвоём', 'spa', 'spa', 'Тихие приватные парные с зоной отдыха'],
            ['Компания 6–10', 'banya', 'fire', 'Большие чаны, мангал и караоке'],
            ['Семья с детьми', 'forest', 'shield', 'Бассейн, безопасность, родительский режим'],
            ['Корпоратив', 'city', 'rocket', 'До 30 человек, кейтеринг и счёт по реквизитам'],
          ].map(([title, variant, icon, desc]) => (
            <div key={title} className="rh-card" style={{ padding: 18, borderRadius: 24 }}>
              <PhotoPlaceholder height={140} variant={variant} radius={18} label="" />
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 14 }}>
                <div style={{ width: 32, height: 32, borderRadius: 999, background: 'rgba(15,118,110,0.10)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Icon name={icon} size={16} />
                </div>
                <span style={{ fontSize: 16, fontWeight: 700 }}>{title}</span>
              </div>
              <p style={{ margin: '8px 0 0', fontSize: 13, color: '#5f6877', lineHeight: 1.5 }}>{desc}</p>
            </div>
          ))}
        </div>
      </Section>

      {/* HOW IT WORKS */}
      <Section padding="36px 28px">
        <div className="rh-card" style={{ padding: 36, borderRadius: 32, background: 'linear-gradient(180deg, rgba(255,255,255,0.96), rgba(250,245,238,0.96))' }}>
          <SectionHeader eyebrow="Как устроен выбор" title="Маршрут до бронирования — 3 шага" />
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 22, marginTop: 28 }}>
            {[
              ['01', 'Выберите сценарий или сразу откройте каталог.', 'Public-маршрут начинается с того, что вам нужно от вечера, а не от названия объекта.'],
              ['02', 'Уточните город, гостей и дату без длинной формы.', 'Поиск экономит клики — фильтры включаются по мере необходимости.'],
              ['03', 'Перейдите к объекту и завершите бронь ближе к финалу.', 'Решение строится на живой выдаче, а не на рекламных обещаниях.'],
            ].map(([n, t, d]) => (
              <div key={n} style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
                <span style={{
                  fontSize: 28, fontWeight: 800, letterSpacing: '-0.04em',
                  color: '#0f766e',
                }}>{n}</span>
                <div>
                  <div style={{ fontSize: 16, fontWeight: 700, marginBottom: 6 }}>{t}</div>
                  <div style={{ fontSize: 13, color: '#5f6877', lineHeight: 1.55 }}>{d}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </Section>

      {/* FEATURED LISTINGS */}
      <Section padding="36px 28px">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end' }}>
          <SectionHeader eyebrow="Подборка" title="Часто бронируют в эти выходные" />
          <a href="#" style={{ color: '#0a5f59', fontWeight: 700, fontSize: 13, textDecoration: 'none', display: 'inline-flex', alignItems: 'center', gap: 6 }}>
            Все 124 объекта <Icon name="arrow" size={14} />
          </a>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 18, marginTop: 22 }}>
          {[
            { name: 'Берёзовая роща', addr: 'Москва, Рублёвское ш., 12', price: '3 200', rating: 4.9, reviews: 128, tags: ['Сауна', 'Парная', 'Чан'], badge: '⚡ Last minute', variant: 'banya' },
            { name: 'Сосновый берег', addr: 'Подмосковье, Истра', price: '4 800', rating: 4.8, reviews: 96, tags: ['Бассейн', 'Купель', 'Мангал'], badge: 'Премиум', variant: 'forest' },
            { name: 'Купеческая', addr: 'Москва, Таганский', price: '2 500', rating: 4.7, reviews: 312, tags: ['Парная', 'Хамам', 'Бар'], badge: 'Хит', variant: 'sand' },
            { name: 'Лофт «Пар»', addr: 'СПб, Васильевский', price: '3 600', rating: 4.9, reviews: 84, tags: ['Сауна', 'Терраса', 'Чан'], badge: 'Новое', variant: 'city' },
          ].map(l => <ListingCard key={l.name} l={l} />)}
        </div>
      </Section>

      {/* OWNER PROMO STRIP */}
      <Section padding="36px 28px">
        <div style={{
          padding: '44px 40px',
          borderRadius: 32,
          background: 'linear-gradient(135deg, #10313a 0%, #38606a 44%, #9a5c30 100%)',
          color: '#fffdf8',
          display: 'grid', gridTemplateColumns: '1.3fr 0.7fr', gap: 28, alignItems: 'center',
        }}>
          <div>
            <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>Партнёрам</span>
            <h2 style={{ margin: '10px 0 12px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.1 }}>0 % комиссии — монетизация только через подписку и продвижение.</h2>
            <p style={{ margin: 0, fontSize: 15, lineHeight: 1.55, color: 'rgba(255,255,255,0.84)', maxWidth: 540 }}>
              Бесплатный CRM-кабинет, календарь, чат с гостями, аналитика. Подключение за один день, без интеграции с телефонией.
            </p>
            <div style={{ display: 'flex', gap: 12, marginTop: 20 }}>
              <button className="rh-btn rh-btn--primary">Стать партнёром <Icon name="arrow" size={16} /></button>
              <button className="rh-btn" style={{ background: 'rgba(255,255,255,0.10)', border: '1px solid rgba(255,255,255,0.20)', color: '#fffdf8' }}>Кабинет владельца</button>
            </div>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            {[['1 240', 'объектов'], ['18 %', 'к выручке партнёров'], ['72 ч', 'возврат средств'], ['4.8 ★', 'средний рейтинг']].map(([v, l]) => (
              <div key={l} style={{ padding: 16, borderRadius: 20, background: 'rgba(255,255,255,0.08)', border: '1px solid rgba(255,255,255,0.14)' }}>
                <div style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.03em' }}>{v}</div>
                <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.74)' }}>{l}</div>
              </div>
            ))}
          </div>
        </div>
      </Section>

      <Footer />
    </div>
  );
}

// ── Helpers ──────────────────────────────────────────────────────
function SectionHeader({ eyebrow, title, subtitle }) {
  return (
    <div>
      {eyebrow ? <span className="rh-eyebrow" style={{ fontSize: 12 }}>{eyebrow}</span> : null}
      {title ? <h2 style={{ margin: '10px 0 0', fontSize: 32, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.1 }}>{title}</h2> : null}
      {subtitle ? <p style={{ margin: '8px 0 0', fontSize: 15, color: '#5f6877', maxWidth: 640, lineHeight: 1.55 }}>{subtitle}</p> : null}
    </div>
  );
}

function SearchField({ label, value, icon }) {
  return (
    <div style={{
      padding: '8px 14px', borderRadius: 18,
      background: 'linear-gradient(180deg, rgba(255,255,255,0.98), rgba(251,247,240,0.86))',
      border: '1px solid rgba(15,23,42,0.08)',
      display: 'flex', alignItems: 'center', gap: 10,
    }}>
      <div style={{ width: 32, height: 32, borderRadius: 12, background: 'rgba(15,118,110,0.08)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Icon name={icon} size={16} />
      </div>
      <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.2, minWidth: 0 }}>
        <span style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>{label}</span>
        <span style={{ fontSize: 14, fontWeight: 600, color: '#16212b', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{value}</span>
      </div>
    </div>
  );
}

function ListingCard({ l }) {
  return (
    <div className="rh-card" style={{ borderRadius: 24, overflow: 'hidden' }}>
      <div style={{ position: 'relative' }}>
        <PhotoPlaceholder height={170} variant={l.variant} radius={0} label="" />
        <span className={`rh-tag ${l.badge.includes('Last') ? 'rh-tag--red' : l.badge.includes('Премиум') ? 'rh-tag--gold' : l.badge.includes('Новое') ? 'rh-tag--cyan' : 'rh-tag--green'}`} style={{ position: 'absolute', top: 12, left: 12 }}>
          {l.badge}
        </span>
        <button style={{
          position: 'absolute', top: 10, right: 10,
          width: 34, height: 34, borderRadius: 999,
          background: 'rgba(255,255,255,0.92)', border: '1px solid rgba(15,23,42,0.08)',
          color: '#16212b', cursor: 'pointer',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}>
          <Icon name="heart" size={16} />
        </button>
      </div>
      <div style={{ padding: 16, display: 'flex', flexDirection: 'column', gap: 8 }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', gap: 8 }}>
          <span style={{ fontSize: 16, fontWeight: 700, lineHeight: 1.25 }}>{l.name} <Icon name="checkc" size={14} style={{ color: '#15803d', marginLeft: 2, verticalAlign: '-2px' }} /></span>
          <span style={{ fontSize: 15, fontWeight: 700, whiteSpace: 'nowrap' }}>{l.price} ₽<span style={{ fontSize: 12, color: '#5f6877' }}>/ч</span></span>
        </div>
        <span style={{ fontSize: 12, color: '#5f6877', display: 'inline-flex', alignItems: 'center', gap: 4 }}>
          <Icon name="pin" size={12} /> {l.addr}
        </span>
        <Stars value={l.rating} count={l.reviews} />
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 4 }}>
          {l.tags.map(t => <span key={t} className="rh-tag">{t}</span>)}
        </div>
      </div>
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageHome, SectionHeader, SearchField, ListingCard });
