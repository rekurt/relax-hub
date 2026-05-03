/* global React */
const { Icon, BrandLockup, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;
const { ListingCard, SectionHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// PUBLIC · 02 CATALOGUE — Filters + grid + map peek
// ════════════════════════════════════════════════════════════════
function PageCatalogue() {
  return (
    <div className="rh-page" style={{ minHeight: 1500 }}>
      <TopNav active="catalogue" />

      {/* Filter bar */}
      <Section padding="20px 28px 8px">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
          <span className="rh-eyebrow-muted" style={{ marginRight: 8 }}>Найдено · 124 объекта</span>
          <FilterChip label="Москва" icon="pin" active />
          <FilterChip label="20 апреля · 19:00" icon="cal" />
          <FilterChip label="4 гостя" icon="user" />
          <FilterChip label="до 4 000 ₽/ч" />
          <FilterChip label="Парная + Чан" />
          <FilterChip label="+ Фильтры" icon="filter" muted />
          <div style={{ marginLeft: 'auto', display: 'flex', gap: 8, alignItems: 'center' }}>
            <span style={{ fontSize: 12, color: '#5f6877', fontWeight: 600 }}>Сортировка</span>
            <button className="rh-btn rh-btn--default rh-btn--sm">Популярные <Icon name="caret" size={14} /></button>
            <div style={{ display: 'flex', background: 'rgba(255,255,255,0.7)', border: '1px solid rgba(15,23,42,0.08)', borderRadius: 999, padding: 4, gap: 2 }}>
              {[['grid', true], ['list', false], ['map', false]].map(([n, on]) => (
                <button key={n} style={{
                  width: 32, height: 32, borderRadius: 999, border: 0, cursor: 'pointer',
                  background: on ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'transparent',
                  color: on ? '#fffdf8' : '#16212b',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}><Icon name={n} size={15} /></button>
              ))}
            </div>
          </div>
        </div>
      </Section>

      <Section padding="14px 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '300px 1fr', gap: 22, alignItems: 'flex-start' }}>
          {/* Left filter rail */}
          <aside className="rh-card" style={{ padding: 22, borderRadius: 24, position: 'sticky', top: 90 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
              <span style={{ fontSize: 16, fontWeight: 800, letterSpacing: '-0.02em' }}>Фильтры</span>
              <button style={linkBtn}>Сбросить</button>
            </div>

            <FilterGroup title="Цена за час">
              <PriceRange />
            </FilterGroup>

            <FilterGroup title="Удобства">
              <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6 }}>
                {['Сауна', 'Парная', 'Хамам', 'Бассейн', 'Чан', 'Купель', 'Мангал', 'Караоке', 'Бар', 'Терраса'].map((a, i) => (
                  <span key={a} className={i < 3 ? 'rh-tag rh-tag--primary' : 'rh-tag'}>{a}</span>
                ))}
              </div>
            </FilterGroup>

            <FilterGroup title="Вместимость">
              <div style={{ display: 'flex', gap: 6 }}>
                {['1–2', '3–4', '5–8', '8+'].map((c, i) => (
                  <span key={c} className={i === 1 ? 'rh-tag rh-tag--primary' : 'rh-tag'} style={{ flex: 1, justifyContent: 'center' }}>{c}</span>
                ))}
              </div>
            </FilterGroup>

            <FilterGroup title="Длительность">
              <div style={{ display: 'flex', gap: 6 }}>
                {['2 ч', '3 ч', '4 ч', '6 ч+'].map(c => <span key={c} className="rh-tag" style={{ flex: 1, justifyContent: 'center' }}>{c}</span>)}
              </div>
            </FilterGroup>

            <FilterGroup title="Тег объекта">
              <CheckRow label="Last minute -20 %" count={18} active />
              <CheckRow label="Подходит для детей" count={42} />
              <CheckRow label="Премиум" count={26} />
              <CheckRow label="Корпоратив" count={31} />
              <CheckRow label="Можно с собакой" count={9} />
            </FilterGroup>

            <button className="rh-btn rh-btn--primary" style={{ width: '100%', marginTop: 6 }}>
              Показать 124 объекта
            </button>
          </aside>

          {/* Listing grid */}
          <div>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 18 }}>
              {[
                { name: 'Берёзовая роща', addr: 'Москва, Рублёвское ш., 12', price: '3 200', rating: 4.9, reviews: 128, tags: ['Сауна', 'Чан', 'Парная'], badge: '⚡ Last minute', variant: 'banya' },
                { name: 'Сосновый берег', addr: 'Истра, дер. Луцино', price: '4 800', rating: 4.8, reviews: 96, tags: ['Бассейн', 'Терраса'], badge: 'Премиум', variant: 'forest' },
                { name: 'Купеческая', addr: 'Москва, Таганский', price: '2 500', rating: 4.7, reviews: 312, tags: ['Парная', 'Хамам'], badge: 'Хит', variant: 'sand' },
                { name: 'Лофт «Пар»', addr: 'СПб, Васильевский', price: '3 600', rating: 4.9, reviews: 84, tags: ['Сауна', 'Чан'], badge: 'Новое', variant: 'city' },
                { name: 'Дача на Истре', addr: 'Истра, ДНП «Лесное»', price: '5 200', rating: 5.0, reviews: 41, tags: ['Бассейн', 'Мангал', 'Терраса'], badge: 'Премиум', variant: 'forest' },
                { name: 'Тёплый камень', addr: 'Москва, Сокольники', price: '2 800', rating: 4.6, reviews: 188, tags: ['Парная', 'Бар'], badge: '⚡ Last minute', variant: 'banya' },
              ].map(l => <ListingCard key={l.name} l={l} />)}
            </div>

            {/* Compare strip */}
            <div className="rh-card" style={{ marginTop: 22, padding: 18, borderRadius: 22, display: 'flex', alignItems: 'center', gap: 14 }}>
              <span style={{ fontSize: 13, fontWeight: 700, color: '#0a5f59' }}>Сравнить (2)</span>
              <div style={{ display: 'flex', gap: 8 }}>
                <CompareSlot name="Берёзовая роща" />
                <CompareSlot name="Сосновый берег" />
                <CompareSlot empty />
              </div>
              <button className="rh-btn rh-btn--primary rh-btn--sm" style={{ marginLeft: 'auto' }}>Открыть сравнение <Icon name="arrow" size={14} /></button>
            </div>

            <div style={{ display: 'flex', justifyContent: 'center', marginTop: 26, gap: 8 }}>
              {['‹','1','2','3','4','…','21','›'].map((p, i) => (
                <button key={i} className={p === '2' ? 'rh-btn rh-btn--primary rh-btn--sm' : 'rh-btn rh-btn--default rh-btn--sm'} style={{ width: 38, padding: 0, justifyContent: 'center' }}>{p}</button>
              ))}
            </div>
          </div>
        </div>
      </Section>

      <Footer />
    </div>
  );
}

// ── helpers ──────────────────────────────────────────────────────
const linkBtn = { background: 'none', border: 0, cursor: 'pointer', color: '#0a5f59', fontWeight: 700, fontSize: 12 };

function FilterChip({ label, icon, active, muted }) {
  return (
    <button className={active ? 'rh-tag rh-tag--primary' : muted ? 'rh-tag rh-tag--ghost' : 'rh-tag'} style={{ height: 36, padding: '0 14px', cursor: 'pointer', fontSize: 13, fontWeight: 600 }}>
      {icon ? <Icon name={icon} size={14} /> : null}{label}
      {!muted ? <Icon name="caret" size={12} /> : null}
    </button>
  );
}

function FilterGroup({ title, children }) {
  return (
    <div style={{ paddingBottom: 16, marginBottom: 16, borderBottom: '1px solid rgba(15,23,42,0.06)' }}>
      <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.12em', textTransform: 'uppercase', color: '#5f6877', marginBottom: 10 }}>{title}</div>
      {children}
    </div>
  );
}

function PriceRange() {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 12, fontWeight: 700, color: '#16212b', marginBottom: 6 }}>
        <span>1 200 ₽</span><span>4 000 ₽</span>
      </div>
      <div style={{ position: 'relative', height: 8, borderRadius: 999, background: 'rgba(15,23,42,0.08)' }}>
        <div style={{ position: 'absolute', left: '12%', right: '32%', top: 0, bottom: 0, borderRadius: 999, background: 'linear-gradient(90deg,#0f766e,#0a5f59)' }}/>
        <div style={knob(12)} /><div style={knob(68)} />
      </div>
      <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 11, color: '#5f6877', marginTop: 8 }}>
        <span>0 ₽</span><span>10 000+ ₽</span>
      </div>
    </div>
  );
}
const knob = (left) => ({ position: 'absolute', left: `calc(${left}% - 8px)`, top: -4, width: 16, height: 16, borderRadius: 999, background: '#fff', border: '2px solid #0a5f59', boxShadow: '0 4px 10px rgba(15,23,42,0.16)' });

function CheckRow({ label, count, active }) {
  return (
    <label style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '6px 0', fontSize: 13, fontWeight: 500, color: '#16212b', cursor: 'pointer' }}>
      <span style={{
        width: 18, height: 18, borderRadius: 6,
        background: active ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : '#fff',
        border: active ? '0' : '1px solid rgba(15,23,42,0.16)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        color: '#fff',
      }}>{active ? <Icon name="check" size={12} /> : null}</span>
      <span style={{ flex: 1 }}>{label}</span>
      <span style={{ fontSize: 12, color: '#5f6877' }}>{count}</span>
    </label>
  );
}

function CompareSlot({ name, empty }) {
  if (empty) return <div style={{ width: 130, height: 56, border: '1.5px dashed rgba(15,23,42,0.18)', borderRadius: 16, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#5f6877', fontSize: 12, fontWeight: 600 }}>+ Добавить</div>;
  return (
    <div style={{ width: 160, padding: '8px 12px', borderRadius: 16, background: 'rgba(15,118,110,0.08)', border: '1px solid rgba(15,118,110,0.22)', display: 'flex', alignItems: 'center', gap: 8, fontSize: 12, fontWeight: 700 }}>
      <PhotoPlaceholder width={36} height={36} radius={10} label="" />
      <span style={{ minWidth: 0, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{name}</span>
      <button style={{ marginLeft: 'auto', background: 'none', border: 0, cursor: 'pointer', color: '#5f6877' }}><Icon name="x" size={14} /></button>
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageCatalogue });
