/* global React */
const { Icon, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;
const { ListingCard, ClientHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// CLIENT · 07 FAVORITES + 08 WALLET (combined screen)
// ════════════════════════════════════════════════════════════════
function PageClientFavorites() {
  return (
    <div className="rh-page" style={{ minHeight: 1100 }}>
      <TopNav active="cabinet" authed />
      <Section padding="22px 28px 0"><ClientHeader active="favorites" /></Section>
      <Section padding="14px 28px 28px">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 18, flexWrap: 'wrap' }}>
          <span className="rh-eyebrow-muted">12 объектов в избранном</span>
          <span className="rh-tag rh-tag--primary">Все</span>
          <span className="rh-tag">Москва (8)</span>
          <span className="rh-tag">СПб (3)</span>
          <span className="rh-tag">Подмосковье (1)</span>
          <button className="rh-btn rh-btn--default rh-btn--sm" style={{ marginLeft: 'auto' }}><Icon name="filter" size={14} /> Фильтр</button>
          <button className="rh-btn rh-btn--default rh-btn--sm">Сравнить (2)</button>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 18 }}>
          {[
            { name: 'Берёзовая роща', addr: 'Москва, Рублёвское ш.', price: '3 200', rating: 4.9, reviews: 128, tags: ['Сауна','Чан'], badge: '⚡ Last minute', variant: 'banya' },
            { name: 'Сосновый берег', addr: 'Истра', price: '4 800', rating: 4.8, reviews: 96, tags: ['Бассейн','Терраса'], badge: 'Премиум', variant: 'forest' },
            { name: 'Купеческая', addr: 'Москва, Таганский', price: '2 500', rating: 4.7, reviews: 312, tags: ['Парная','Хамам'], badge: 'Хит', variant: 'sand' },
            { name: 'Лофт «Пар»', addr: 'СПб, Васильевский', price: '3 600', rating: 4.9, reviews: 84, tags: ['Сауна','Чан'], badge: 'Новое', variant: 'city' },
            { name: 'Дача на Истре', addr: 'Истра, ДНП «Лесное»', price: '5 200', rating: 5.0, reviews: 41, tags: ['Бассейн','Мангал'], badge: 'Премиум', variant: 'forest' },
            { name: 'Тёплый камень', addr: 'Москва, Сокольники', price: '2 800', rating: 4.6, reviews: 188, tags: ['Парная','Бар'], badge: '⚡ Last minute', variant: 'banya' },
            { name: 'Чан и звёзды', addr: 'Подмосковье, Истра', price: '6 200', rating: 4.9, reviews: 28, tags: ['Чан','Терраса'], badge: 'Новое', variant: 'spa' },
            { name: 'Городская парилка', addr: 'СПб, Петроградский', price: '2 200', rating: 4.5, reviews: 220, tags: ['Парная'], badge: 'Хит', variant: 'city' },
          ].map(l => <ListingCard key={l.name} l={l} />)}
        </div>
      </Section>
      <Footer />
    </div>
  );
}

function PageClientWallet() {
  return (
    <div className="rh-page" style={{ minHeight: 1000 }}>
      <TopNav active="cabinet" authed />
      <Section padding="22px 28px 0"><ClientHeader active="wallet" /></Section>
      <Section padding="14px 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr 1fr', gap: 18 }}>
          {/* Wallet card */}
          <div className="rh-card" style={{ padding: 26, borderRadius: 28, background: 'linear-gradient(135deg,#10313a,#38606a 60%, #9a5c30)', color: '#fffdf8', border: 'none' }}>
            <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>RH-кошелёк</div>
            <div style={{ fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em', marginTop: 8 }}>4 280 ₽</div>
            <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.74)' }}>+ 540 ₽ за апрель</div>
            <div style={{ display: 'flex', gap: 8, marginTop: 22 }}>
              <button className="rh-btn" style={{ background: '#fffdf8', color: '#16212b', flex: 1 }}><Icon name="plus" size={14} /> Пополнить</button>
              <button className="rh-btn" style={{ background: 'rgba(255,255,255,0.14)', color: '#fffdf8', border: '1px solid rgba(255,255,255,0.20)' }}><Icon name="upload" size={14} /></button>
            </div>
          </div>
          {/* Loyalty */}
          <div className="rh-card" style={{ padding: 26, borderRadius: 28 }}>
            <span className="rh-eyebrow-muted">Уровень лояльности</span>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 6 }}>
              <Icon name="fire" size={22} style={{ color: '#d97706' }} />
              <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em' }}>Завсегдатай</div>
            </div>
            <div style={{ fontSize: 12, color: '#5f6877', marginTop: 4 }}>3 брони до уровня «Хозяин пара»</div>
            <div style={{ height: 10, marginTop: 16, borderRadius: 999, background: 'rgba(15,23,42,0.08)', overflow: 'hidden' }}>
              <div style={{ width: '62%', height: '100%', background: 'linear-gradient(90deg,#d97706,#b45309)' }} />
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, color: '#5f6877', marginTop: 8 }}>
              <span>5 / 8 броней</span><span>+5 % бонусов</span>
            </div>
          </div>
          {/* Promo */}
          <div className="rh-card" style={{ padding: 26, borderRadius: 28, background: 'linear-gradient(180deg, rgba(217,119,6,0.10), rgba(255,255,255,0.86))' }}>
            <span className="rh-eyebrow" style={{ color: '#92400e' }}>Промокод</span>
            <h3 style={{ margin: '6px 0 4px', fontSize: 22, fontWeight: 800, letterSpacing: '-0.02em' }}>WEEKEND25</h3>
            <p style={{ margin: 0, fontSize: 13, color: '#5f6877', lineHeight: 1.5 }}>−25 % на любую бронь до 31 мая. Один раз на гостя.</p>
            <button className="rh-btn rh-btn--default rh-btn--sm" style={{ marginTop: 14 }}><Icon name="check" size={14} /> Скопировать</button>
          </div>
        </div>

        {/* Transactions */}
        <div className="rh-card" style={{ marginTop: 22, padding: 24, borderRadius: 28 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <h3 style={{ margin: 0, fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em' }}>История</h3>
            <div style={{ display: 'flex', gap: 8 }}>
              <span className="rh-tag rh-tag--primary">Все</span>
              <span className="rh-tag">Списания</span>
              <span className="rh-tag">Бонусы</span>
              <span className="rh-tag">Возвраты</span>
            </div>
          </div>
          {[
            ['Бронь · Берёзовая роща', '20 апр', '−11 440 ₽', '−'],
            ['Бонус за визит', '15 апр', '+ 270 ₽', '+'],
            ['Возврат · Лофт «Пар»', '10 апр', '+ 3 600 ₽', '+'],
            ['Бронь · Сосновый берег', '3 апр', '−22 200 ₽', '−'],
            ['Пополнение картой', '28 марта', '+ 5 000 ₽', '+'],
          ].map(([label, date, amount, sign]) => (
            <div key={label + date} style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', padding: '14px 0', borderTop: '1px solid rgba(15,23,42,0.06)' }}>
              <div style={{ display: 'flex', gap: 12, alignItems: 'center' }}>
                <div style={{ width: 36, height: 36, borderRadius: 12, background: sign === '+' ? 'rgba(21,128,61,0.10)' : 'rgba(15,23,42,0.06)', color: sign === '+' ? '#15803d' : '#16212b', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                  <Icon name={sign === '+' ? 'plus' : 'card'} size={16} />
                </div>
                <div style={{ lineHeight: 1.3 }}>
                  <div style={{ fontSize: 14, fontWeight: 700 }}>{label}</div>
                  <div style={{ fontSize: 12, color: '#5f6877' }}>{date}</div>
                </div>
              </div>
              <div style={{ fontSize: 15, fontWeight: 700, color: sign === '+' ? '#15803d' : '#16212b' }}>{amount}</div>
            </div>
          ))}
        </div>
      </Section>
      <Footer />
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageClientFavorites, PageClientWallet });
