/* global React */
const { Icon, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { OwnerNav, Footer, Section, SectionHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// OWNER · 10 CALENDAR + 11 OBJECTS + 12 PAYOUTS
// ════════════════════════════════════════════════════════════════
function PageOwnerCalendar() {
  const days = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'];
  const hours = ['10', '12', '14', '16', '18', '20', '22'];
  const slots = [
    // [dayIdx, startHour, durHours, status, label]
    [0, 14, 2, 'confirmed', 'Дмитрий М. · 4 гостя'],
    [1, 16, 4, 'confirmed', 'Корпоратив · 12'],
    [2, 18, 4, 'awaiting', 'Анна П. · 6'],
    [3, 12, 3, 'confirmed', 'Семья К.'],
    [3, 19, 3, 'block',   'Тех. перерыв'],
    [4, 16, 4, 'confirmed', 'Михаил К.'],
    [5, 14, 4, 'confirmed', 'Берёзовая роща'],
    [5, 20, 3, 'awaiting', 'Last min · 4'],
    [6, 12, 4, 'confirmed', 'Воскр. компания'],
    [6, 18, 4, 'confirmed', 'Семья В. · 5'],
  ];
  const tone = (s) => s === 'confirmed'
    ? { bg: 'linear-gradient(135deg,#0f766e,#0a5f59)', color: '#fffdf8' }
    : s === 'awaiting'
    ? { bg: 'linear-gradient(135deg,#d97706,#b45309)', color: '#fffdf8' }
    : { bg: 'repeating-linear-gradient(135deg, rgba(15,23,42,0.10) 0 6px, rgba(15,23,42,0.04) 6px 12px)', color: '#5f6877' };

  return (
    <div className="rh-page" style={{ minHeight: 1200 }}>
      <OwnerNav active="calendar" />
      <Section padding="22px 28px 0">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: 18 }}>
          <div>
            <span className="rh-eyebrow">Календарь</span>
            <h1 style={{ margin: '8px 0 6px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em' }}>14–20 апреля</h1>
            <p style={{ margin: 0, fontSize: 14, color: '#5f6877' }}>Берёзовая роща · 2 парных. Тяните брони, чтобы перенести; кликните пустую ячейку, чтобы заблокировать.</p>
          </div>
          <div style={{ display: 'flex', gap: 10 }}>
            <button className="rh-btn rh-btn--default"><Icon name="arrl" size={16} /></button>
            <button className="rh-btn rh-btn--default">Сегодня</button>
            <button className="rh-btn rh-btn--default"><Icon name="arrow" size={16} /></button>
            <button className="rh-btn rh-btn--primary"><Icon name="plus" size={16} /> Слот</button>
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
          <span className="rh-tag rh-tag--primary">Все объекты</span>
          <span className="rh-tag">Берёзовая роща · Парная</span>
          <span className="rh-tag">Сосновый берег · Бассейн</span>
          <span className="rh-tag">Купеческая · Хамам</span>
          <span className="rh-tag rh-tag--ghost" style={{ marginLeft: 'auto' }}>Неделя</span>
          <span className="rh-tag">Месяц</span>
          <span className="rh-tag">День</span>
        </div>
      </Section>

      <Section padding="22px 28px 28px">
        <div className="rh-card" style={{ padding: 22, borderRadius: 28 }}>
          {/* Calendar grid */}
          <div style={{ display: 'grid', gridTemplateColumns: '60px repeat(7, 1fr)', gap: 0 }}>
            <div />
            {days.map((d, i) => (
              <div key={d} style={{ padding: '10px 12px', borderBottom: '1px solid rgba(15,23,42,0.06)' }}>
                <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>{d}</div>
                <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em', color: i === 3 ? '#0a5f59' : '#16212b' }}>{14 + i}</div>
              </div>
            ))}
            {/* Hour rows */}
            {hours.map((h, hi) => (
              <React.Fragment key={h}>
                <div style={{ padding: '0 8px', fontSize: 11, color: '#5f6877', fontWeight: 700, fontVariantNumeric: 'tabular-nums', borderTop: '1px solid rgba(15,23,42,0.04)', minHeight: 80, paddingTop: 8 }}>{h}:00</div>
                {days.map((_, di) => (
                  <div key={di} style={{ position: 'relative', borderTop: '1px solid rgba(15,23,42,0.04)', borderLeft: '1px dashed rgba(15,23,42,0.06)', minHeight: 80 }}>
                    {hi === 0 && slots.filter(s => s[0] === di).map(([d, sh, dur, status, label], i) => {
                      const top = (sh - 10) * 40 + 6;
                      const height = dur * 40 - 8;
                      const t = tone(status);
                      return (
                        <div key={i} style={{
                          position: 'absolute', left: 6, right: 6, top, height,
                          background: t.bg, color: t.color,
                          borderRadius: 14, padding: '8px 10px', fontSize: 11, fontWeight: 700,
                          boxShadow: status !== 'block' ? '0 8px 18px rgba(15,118,110,0.18)' : 'none',
                          overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
                          display: 'flex', flexDirection: 'column', justifyContent: 'space-between',
                        }}>
                          <span>{label}</span>
                          <span style={{ fontSize: 10, opacity: 0.84, fontWeight: 600 }}>{`${sh.toString().padStart(2, '0')}:00 – ${(sh + dur).toString().padStart(2, '0')}:00`}</span>
                        </div>
                      );
                    })}
                  </div>
                ))}
              </React.Fragment>
            ))}
          </div>
          <div style={{ display: 'flex', gap: 18, marginTop: 18, paddingTop: 14, borderTop: '1px solid rgba(15,23,42,0.06)', fontSize: 12, color: '#5f6877' }}>
            <Legend tone="confirmed" label="Подтверждено" />
            <Legend tone="awaiting" label="Ждёт подтверждения" />
            <Legend tone="block" label="Заблокировано" />
            <span style={{ marginLeft: 'auto' }}>Загрузка недели · <strong style={{ color: '#16212b' }}>74 %</strong></span>
          </div>
        </div>
      </Section>
      <Footer />
    </div>
  );
}

function Legend({ tone, label }) {
  const map = {
    confirmed: 'linear-gradient(135deg,#0f766e,#0a5f59)',
    awaiting: 'linear-gradient(135deg,#d97706,#b45309)',
    block: 'repeating-linear-gradient(135deg, rgba(15,23,42,0.18) 0 4px, rgba(15,23,42,0.06) 4px 8px)',
  };
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
      <span style={{ width: 14, height: 14, borderRadius: 6, background: map[tone] }} /> {label}
    </span>
  );
}

// ── OBJECTS ──────────────────────────────────────────────────
function PageOwnerObjects() {
  const objs = [
    { name: 'Берёзовая роща', addr: 'Москва, Рублёвское ш., 12', units: 2, status: 'live', occ: 78, rev: '64 200 ₽', rating: 4.9, rev_count: 128, variant: 'banya' },
    { name: 'Сосновый берег', addr: 'Подмосковье, Истра', units: 3, status: 'live', occ: 64, rev: '48 800 ₽', rating: 4.8, rev_count: 96, variant: 'forest' },
    { name: 'Купеческая', addr: 'Москва, Таганский', units: 1, status: 'draft', occ: 0, rev: '0 ₽', rating: null, rev_count: 0, variant: 'sand' },
  ];
  return (
    <div className="rh-page" style={{ minHeight: 1100 }}>
      <OwnerNav active="objects" />
      <Section padding="22px 28px 0">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: 18 }}>
          <div>
            <span className="rh-eyebrow">Объекты</span>
            <h1 style={{ margin: '8px 0 6px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em' }}>3 объекта · 6 единиц</h1>
            <p style={{ margin: 0, fontSize: 14, color: '#5f6877' }}>Карточка объекта = карточка в каталоге. Чем полнее заполнено — тем выше CTR.</p>
          </div>
          <div style={{ display: 'flex', gap: 10 }}>
            <button className="rh-btn rh-btn--default"><Icon name="upload" size={16} /> Импорт</button>
            <button className="rh-btn rh-btn--primary"><Icon name="plus" size={16} /> Новый объект</button>
          </div>
        </div>
      </Section>

      <Section padding="22px 28px 28px">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
          {objs.map(o => (
            <div key={o.name} className="rh-card" style={{ padding: 18, borderRadius: 28, display: 'grid', gridTemplateColumns: '160px 1.4fr 1fr auto', gap: 22, alignItems: 'center' }}>
              <PhotoPlaceholder width={160} height={120} radius={20} variant={o.variant} label="" />
              <div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
                  <span className={`rh-tag ${o.status === 'live' ? 'rh-tag--green' : 'rh-tag--gold'}`}>{o.status === 'live' ? 'В каталоге' : 'Черновик'}</span>
                  {o.rating ? <Stars value={o.rating} count={o.rev_count} /> : <span style={{ fontSize: 12, color: '#5f6877' }}>Нет отзывов</span>}
                </div>
                <h3 style={{ margin: '0 0 4px', fontSize: 20, fontWeight: 800, letterSpacing: '-0.02em' }}>{o.name}</h3>
                <div style={{ fontSize: 12, color: '#5f6877', display: 'inline-flex', gap: 4, alignItems: 'center' }}><Icon name="pin" size={12} /> {o.addr}</div>
                <div style={{ marginTop: 10, display: 'flex', gap: 6, flexWrap: 'wrap' }}>
                  <span className="rh-tag">{o.units} {o.units === 1 ? 'единица' : 'единицы'}</span>
                  <span className="rh-tag">Парная</span>
                  <span className="rh-tag">Бассейн</span>
                  <span className="rh-tag">Чан</span>
                </div>
              </div>
              <div>
                <div style={{ fontSize: 11, fontWeight: 700, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Загрузка / выручка · апрель</div>
                <div style={{ display: 'flex', alignItems: 'baseline', gap: 12, marginTop: 6 }}>
                  <span style={{ fontSize: 24, fontWeight: 800, letterSpacing: '-0.03em' }}>{o.occ}%</span>
                  <span style={{ fontSize: 16, fontWeight: 700, color: '#0a5f59' }}>{o.rev}</span>
                </div>
                <div style={{ height: 8, marginTop: 10, borderRadius: 999, background: 'rgba(15,23,42,0.06)', overflow: 'hidden' }}>
                  <div style={{ width: `${o.occ}%`, height: '100%', background: 'linear-gradient(90deg,#0f766e,#0a5f59)' }} />
                </div>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 6, alignItems: 'flex-end' }}>
                <button className="rh-btn rh-btn--primary rh-btn--sm">Открыть</button>
                <button className="rh-btn rh-btn--default rh-btn--sm"><Icon name="pencil" size={14} /> Контент</button>
                <button className="rh-btn rh-btn--default rh-btn--sm"><Icon name="cal" size={14} /> Слоты</button>
              </div>
            </div>
          ))}
        </div>
      </Section>
      <Footer />
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageOwnerCalendar, PageOwnerObjects });
