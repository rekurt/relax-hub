/* global React */
const { Icon, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { OwnerNav, Footer, Section } = window.RH;
const { SectionHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// OWNER · 09 DASHBOARD
// ════════════════════════════════════════════════════════════════
function PageOwnerDashboard() {
  return (
    <div className="rh-page" style={{ minHeight: 1500 }}>
      <OwnerNav active="dashboard" />
      <Section padding="22px 28px 0">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', gap: 18 }}>
          <div>
            <span className="rh-eyebrow">Сводка · 14 апреля</span>
            <h1 style={{ margin: '8px 0 6px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em' }}>Доброе утро, Иван</h1>
            <p style={{ margin: 0, fontSize: 14, color: '#5f6877' }}>Сегодня 5 бронирований и 2 заявки ждут подтверждения. Без звонков.</p>
          </div>
          <div style={{ display: 'flex', gap: 10 }}>
            <button className="rh-btn rh-btn--default"><Icon name="download" size={16} /> Экспорт</button>
            <button className="rh-btn rh-btn--primary"><Icon name="plus" size={16} /> Создать бронь</button>
          </div>
        </div>
      </Section>

      <Section padding="22px 28px 0">
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 14 }}>
          {[
            { eb: 'Бронирований сегодня', v: '12', sub: '+3 к прошлой неделе', tone: 'teal', icon: 'cal' },
            { eb: 'Выручка за месяц', v: '128 480 ₽', sub: 'Выплата 25 апреля', tone: 'amber', icon: 'card' },
            { eb: 'Загрузка слотов', v: '74 %', sub: 'Этот четверг', tone: 'teal', icon: 'trend' },
            { eb: 'Средний рейтинг', v: '4.9 ★', sub: '128 отзывов', tone: 'gold', icon: 'star' },
          ].map(t => (
            <div key={t.eb} className="rh-card" style={{ padding: 20, borderRadius: 22 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <span className="rh-eyebrow-muted">{t.eb}</span>
                <div style={{ width: 32, height: 32, borderRadius: 12, background: 'rgba(15,118,110,0.08)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icon name={t.icon} size={16} /></div>
              </div>
              <div style={{ fontSize: 32, fontWeight: 800, letterSpacing: '-0.04em', marginTop: 8 }}>{t.v}</div>
              <div style={{ fontSize: 12, color: '#5f6877' }}>{t.sub}</div>
            </div>
          ))}
        </div>
      </Section>

      <Section padding="22px 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: 22, alignItems: 'flex-start' }}>
          {/* Today's bookings */}
          <div className="rh-card" style={{ padding: 24, borderRadius: 28 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
              <SectionHeader eyebrow="Сегодня" title="Бронирования на 14 апреля" />
              <button style={{ background: 'none', border: 0, color: '#0a5f59', fontWeight: 700, fontSize: 13, cursor: 'pointer' }}>Все 12 <Icon name="arrow" size={12} /></button>
            </div>
            {[
              { time: '14:00', name: 'Дмитрий М.', guests: 4, obj: 'Берёзовая роща · Парная', total: '12 800 ₽', s: 'confirmed' },
              { time: '17:00', name: 'Анна П.', guests: 6, obj: 'Сосновый берег · Бассейн', total: '22 200 ₽', s: 'awaiting' },
              { time: '19:00', name: 'Михаил К.', guests: 4, obj: 'Берёзовая роща · Парная', total: '11 440 ₽', s: 'confirmed' },
              { time: '21:00', name: 'Ольга В.', guests: 8, obj: 'Сосновый берег · Корпоратив', total: '34 800 ₽', s: 'awaiting' },
            ].map((b, i) => (
              <div key={i} style={{ display: 'grid', gridTemplateColumns: 'auto 1fr auto auto', gap: 14, alignItems: 'center', padding: '12px 0', borderTop: i ? '1px solid rgba(15,23,42,0.06)' : '0' }}>
                <div style={{ width: 60, fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em', fontVariantNumeric: 'tabular-nums' }}>{b.time}</div>
                <div>
                  <div style={{ fontSize: 14, fontWeight: 700 }}>{b.name} <span style={{ fontSize: 12, color: '#5f6877', fontWeight: 500 }}>· {b.guests} гостей</span></div>
                  <div style={{ fontSize: 12, color: '#5f6877' }}>{b.obj}</div>
                </div>
                <div style={{ fontSize: 14, fontWeight: 700 }}>{b.total}</div>
                {b.s === 'awaiting' ? (
                  <div style={{ display: 'flex', gap: 6 }}>
                    <button className="rh-btn rh-btn--primary rh-btn--sm" style={{ padding: '0 12px' }}><Icon name="check" size={14} /></button>
                    <button className="rh-btn rh-btn--default rh-btn--sm" style={{ padding: '0 12px' }}><Icon name="x" size={14} /></button>
                  </div>
                ) : <span className="rh-tag rh-tag--green">Подтверждено</span>}
              </div>
            ))}
          </div>

          {/* Activity feed */}
          <div className="rh-card" style={{ padding: 24, borderRadius: 28 }}>
            <SectionHeader eyebrow="Лента" title="Активность" />
            <div style={{ marginTop: 14, display: 'flex', flexDirection: 'column', gap: 10 }}>
              {[
                { t: 'Бронирование подтверждено', s: 'Берёзовая роща · 20 апр, 19:00', tone: 'teal', dot: true },
                { t: 'Новый отзыв на «Сосновый берег»', s: '★ 5.0 · «Тишина и температура — топ»', tone: 'amber' },
                { t: 'Возврат за бронь #4821 проведён', s: '2 100 ₽ → карта •• 4422', tone: 'plum' },
                { t: 'Запрос на бронь', s: 'Ольга В. · 21:00 · 8 гостей', tone: 'teal', dot: true },
              ].map((it, i) => (
                <div key={i} style={{
                  display: 'flex', gap: 12, padding: 14, borderRadius: 18,
                  background: it.dot ? 'linear-gradient(180deg, rgba(15,118,110,0.07), rgba(255,255,255,0.74))' : 'rgba(255,255,255,0.6)',
                  border: '1px solid rgba(15,23,42,0.06)',
                }}>
                  {it.dot ? <span style={{ width: 8, height: 8, borderRadius: 999, background: '#0a5f59', boxShadow: '0 0 0 5px rgba(15,118,110,0.10)', marginTop: 7, flexShrink: 0 }} /> : <span style={{ width: 8, height: 8, marginTop: 7 }} />}
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontSize: 13, fontWeight: 700 }}>{it.t}</div>
                    <div style={{ fontSize: 12, color: '#5f6877', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{it.s}</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Revenue chart */}
        <div className="rh-card" style={{ marginTop: 22, padding: 28, borderRadius: 28 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: 18 }}>
            <SectionHeader eyebrow="Аналитика" title="Выручка за 14 дней" />
            <div style={{ display: 'flex', gap: 6 }}>
              <span className="rh-tag rh-tag--primary">14 дней</span>
              <span className="rh-tag">Месяц</span>
              <span className="rh-tag">Квартал</span>
            </div>
          </div>
          <RevenueChart />
        </div>
      </Section>
      <Footer />
    </div>
  );
}

function RevenueChart() {
  const data = [42, 38, 64, 58, 72, 88, 96, 78, 64, 88, 104, 92, 116, 128];
  const max = 130;
  const W = 1180, H = 220;
  const step = W / (data.length - 1);
  const points = data.map((v, i) => [i * step, H - (v / max) * H]);
  const path = points.map((p, i) => `${i ? 'L' : 'M'}${p[0]},${p[1]}`).join(' ');
  const area = `${path} L${W},${H} L0,${H} Z`;
  return (
    <div style={{ position: 'relative' }}>
      <svg width="100%" height={H + 30} viewBox={`0 0 ${W} ${H + 30}`}>
        <defs>
          <linearGradient id="rev" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0" stopColor="#0f766e" stopOpacity="0.30"/>
            <stop offset="1" stopColor="#0f766e" stopOpacity="0"/>
          </linearGradient>
        </defs>
        {[0.25, 0.5, 0.75].map(p => <line key={p} x1="0" x2={W} y1={H * p} y2={H * p} stroke="rgba(15,23,42,0.06)" strokeDasharray="4 6"/>)}
        <path d={area} fill="url(#rev)"/>
        <path d={path} stroke="#0a5f59" strokeWidth="2.5" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
        {points.map(([x, y], i) => <circle key={i} cx={x} cy={y} r={i === 12 ? 6 : 3} fill="#fff" stroke="#0a5f59" strokeWidth={i === 12 ? 3 : 2}/>)}
        {data.map((v, i) => i % 2 === 0 ? <text key={i} x={i * step} y={H + 22} fontSize="11" fill="#5f6877" textAnchor="middle">{1 + i} апр</text> : null)}
      </svg>
      <div style={{ position: 'absolute', left: '85%', top: 14, padding: '8px 12px', borderRadius: 14, background: '#16212b', color: '#fff', fontSize: 12, fontWeight: 700, boxShadow: '0 14px 28px rgba(15,23,42,0.20)' }}>
        13 апр · 116 800 ₽
      </div>
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageOwnerDashboard });
