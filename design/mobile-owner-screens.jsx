/* global React */
const { MIcon, MPage, MCard, MHeader, MTag, MButton, MPhoto, MAvatar, MTabBar, MStars } = window.RHM;

const OWNER_TABS = [
  ['home', 'Сводка', 'home'],
  ['cal', 'Календарь', 'cal'],
  ['inbox', 'Заявки', 'msg'],
  ['stats', 'Аналитика', 'trend'],
  ['profile', 'Кабинет', 'user'],
];

// ════════════════════════════════════════════════════════════
// OWNER · 01 TODAY (lightweight dashboard)
// ════════════════════════════════════════════════════════════
function MOwnerToday() {
  return (
    <MPage>
      <MHeader
        eyebrow="14 апреля · понедельник"
        title={<>Доброе утро,<br/>Иван</>}
        sub="5 броней сегодня. 2 ждут вас."
        right={<MAvatar name="Иван Соколов" size={42} tone="amber" />}
      />

      {/* Stat tiles */}
      <div style={{ padding: '0 20px 14px', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
        {[
          { eb: 'Сегодня', v: '5', sub: 'броней', ic: 'cal' },
          { eb: 'Загрузка', v: '74%', sub: 'неделя', ic: 'trend' },
          { eb: 'Выручка', v: '128k', sub: 'апрель ₽', ic: 'card' },
          { eb: 'Рейтинг', v: '4.9', sub: '128 отзывов', ic: 'star' },
        ].map(t => (
          <MCard key={t.eb} padding={14} radius={20}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
              <span style={{ fontSize: 10, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>{t.eb}</span>
              <div style={{ width: 26, height: 26, borderRadius: 9, background: 'rgba(15,118,110,0.08)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                <MIcon name={t.ic} size={14} />
              </div>
            </div>
            <div style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.04em', marginTop: 4 }}>{t.v}</div>
            <div style={{ fontSize: 11, color: '#5f6877' }}>{t.sub}</div>
          </MCard>
        ))}
      </div>

      {/* Awaiting actions */}
      <div style={{ padding: '6px 20px 14px' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
          <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Ждут вас</span>
          <MTag tone="gold">2 заявки</MTag>
        </div>
        {[
          { name: 'Анна П.', when: 'Сегодня, 17:00 · 6 гостей', total: '22 200 ₽', tone: 'plum' },
          { name: 'Ольга В.', when: 'Сегодня, 21:00 · 8 гостей', total: '34 800 ₽', tone: 'teal' },
        ].map((b, i) => (
          <MCard key={i} padding={14} radius={20} style={{ marginBottom: 10 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 10 }}>
              <MAvatar name={b.name} size={36} tone={b.tone} />
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, fontWeight: 700 }}>{b.name}</div>
                <div style={{ fontSize: 11, color: '#5f6877' }}>{b.when}</div>
              </div>
              <span style={{ fontSize: 14, fontWeight: 800 }}>{b.total}</span>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <MButton variant="default" size="sm" wide><MIcon name="x" size={14} /> Отклонить</MButton>
              <MButton variant="primary" size="sm" wide><MIcon name="check" size={14} /> Подтвердить</MButton>
            </div>
          </MCard>
        ))}
      </div>

      {/* Today's schedule strip */}
      <div style={{ padding: '0 20px 14px' }}>
        <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Расписание сегодня</span>
        <MCard padding={10} radius={20} style={{ marginTop: 10 }}>
          {[
            { t: '14:00', name: 'Дмитрий М.', sub: 'Парная · 4 ч', ok: true },
            { t: '17:00', name: 'Анна П.', sub: 'Бассейн · 4 ч', ok: false },
            { t: '19:00', name: 'Михаил К.', sub: 'Парная · 3 ч', ok: true },
          ].map((b, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 6px', borderTop: i ? '1px solid rgba(15,23,42,0.06)' : '0' }}>
              <span style={{ width: 50, fontSize: 16, fontWeight: 800, letterSpacing: '-0.02em', fontVariantNumeric: 'tabular-nums' }}>{b.t}</span>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 13, fontWeight: 700 }}>{b.name}</div>
                <div style={{ fontSize: 11, color: '#5f6877' }}>{b.sub}</div>
              </div>
              {b.ok ? <MTag tone="green">OK</MTag> : <MTag tone="gold">Ждёт</MTag>}
            </div>
          ))}
        </MCard>
      </div>

      <MTabBar items={OWNER_TABS} active="home" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// OWNER · 02 CALENDAR (day view)
// ════════════════════════════════════════════════════════════
function MOwnerCalendar() {
  const days = [['Пн', 14], ['Вт', 15], ['Ср', 16], ['Чт', 17, true], ['Пт', 18], ['Сб', 19], ['Вс', 20]];
  return (
    <MPage>
      <MHeader eyebrow="Календарь" title="Чт, 17 апреля" sub="Берёзовая роща · парная" right={<MButton variant="primary" size="sm"><MIcon name="plus" size={14} /></MButton>} />

      {/* Week strip */}
      <div style={{ padding: '0 20px 16px', display: 'flex', gap: 6 }}>
        {days.map(([d, n, on]) => (
          <div key={n} style={{
            flex: 1, padding: '10px 0', borderRadius: 16, textAlign: 'center',
            background: on ? 'linear-gradient(180deg,#0f766e,#0a5f59)' : 'rgba(255,255,255,0.74)',
            color: on ? '#fffdf8' : '#16212b',
            border: '1px solid rgba(15,23,42,0.08)',
            boxShadow: on ? '0 8px 16px rgba(15,118,110,0.22)' : 'none',
          }}>
            <div style={{ fontSize: 10, fontWeight: 700, opacity: on ? 0.84 : 0.6, letterSpacing: '0.10em', textTransform: 'uppercase' }}>{d}</div>
            <div style={{ fontSize: 18, fontWeight: 800, letterSpacing: '-0.03em' }}>{n}</div>
          </div>
        ))}
      </div>

      {/* Day timeline */}
      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={12} radius={22}>
          {[
            { t: '10:00', span: 'Свободно', kind: 'free' },
            { t: '12:00', span: 'Семья К. · 4', kind: 'ok', dur: '12:00–15:00', note: 'оплачено · 11 200 ₽' },
            { t: '14:00', span: '', kind: 'cont' },
            { t: '16:00', span: 'Last min · Анна П. · 6', kind: 'wait', dur: '16:00–20:00', note: 'ждёт подтверждения' },
            { t: '18:00', span: '', kind: 'cont' },
            { t: '20:00', span: 'Тех. перерыв', kind: 'block', dur: '20:00–22:00', note: 'уборка после долгой брони' },
          ].map((s, i) => (
            <div key={i} style={{ display: 'grid', gridTemplateColumns: '52px 1fr', gap: 10, alignItems: 'flex-start', padding: '8px 0', borderTop: i ? '1px solid rgba(15,23,42,0.04)' : '0' }}>
              <span style={{ fontSize: 12, fontWeight: 700, color: '#5f6877', fontVariantNumeric: 'tabular-nums' }}>{s.t}</span>
              {s.kind === 'free' && (
                <div style={{ padding: '10px 12px', borderRadius: 14, border: '1px dashed rgba(15,23,42,0.18)', color: '#5f6877', fontSize: 13 }}>Свободно — постучать клиентам?</div>
              )}
              {s.kind === 'ok' && (
                <div style={{ padding: '12px 14px', borderRadius: 14, background: 'linear-gradient(135deg,#0f766e,#0a5f59)', color: '#fffdf8', boxShadow: '0 8px 18px rgba(15,118,110,0.22)' }}>
                  <div style={{ fontSize: 13, fontWeight: 800 }}>{s.span}</div>
                  <div style={{ fontSize: 11, opacity: 0.84 }}>{s.dur} · {s.note}</div>
                </div>
              )}
              {s.kind === 'wait' && (
                <div style={{ padding: '12px 14px', borderRadius: 14, background: 'linear-gradient(135deg,#d97706,#b45309)', color: '#fffdf8', boxShadow: '0 8px 18px rgba(217,119,6,0.22)' }}>
                  <div style={{ fontSize: 13, fontWeight: 800 }}>{s.span}</div>
                  <div style={{ fontSize: 11, opacity: 0.84 }}>{s.dur} · {s.note}</div>
                </div>
              )}
              {s.kind === 'block' && (
                <div style={{ padding: '12px 14px', borderRadius: 14, background: 'repeating-linear-gradient(135deg, rgba(15,23,42,0.10) 0 6px, rgba(15,23,42,0.04) 6px 12px)', color: '#5f6877' }}>
                  <div style={{ fontSize: 13, fontWeight: 800, color: '#16212b' }}>{s.span}</div>
                  <div style={{ fontSize: 11 }}>{s.dur} · {s.note}</div>
                </div>
              )}
              {s.kind === 'cont' && <div style={{ height: 1 }} />}
            </div>
          ))}
        </MCard>
      </div>

      <MTabBar items={OWNER_TABS} active="cal" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// OWNER · 03 INBOX (chat list)
// ════════════════════════════════════════════════════════════
function MOwnerInbox() {
  return (
    <MPage>
      <MHeader eyebrow="Чаты с гостями" title="Заявки и переписка" sub="3 непрочитанных. Среднее время ответа — 12 мин." />

      <div style={{ padding: '0 20px 14px' }}>
        <div style={{ display: 'inline-flex', padding: 4, borderRadius: 999, background: 'rgba(255,255,255,0.7)', border: '1px solid rgba(15,23,42,0.08)' }}>
          {[['Все', false], ['Заявки', true], ['Активные', false]].map(([l, on]) => (
            <button key={l} style={{
              padding: '8px 14px', borderRadius: 999, border: 0,
              background: on ? '#fff' : 'transparent',
              boxShadow: on ? '0 4px 10px rgba(15,23,42,0.06)' : 'none',
              fontSize: 12, fontWeight: 700, color: on ? '#16212b' : '#5f6877',
            }}>{l}</button>
          ))}
        </div>
      </div>

      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={6} radius={22}>
          {[
            { name: 'Анна П.', last: 'А чан греется отдельно?', when: '12 мин', tone: 'plum', unread: 2, badge: 'Ждёт ответа', tag: 'gold' },
            { name: 'Ольга В.', last: 'Спасибо! Будем в 21:00.', when: '1 ч', tone: 'teal', unread: 0 },
            { name: 'Михаил К.', last: 'Отправил оплату через СБП', when: '3 ч', tone: 'amber', unread: 1, badge: '✓ Подтв.', tag: 'green' },
            { name: 'Семья К.', last: 'До завтра!', when: 'Вчера', tone: 'blue', unread: 0 },
            { name: 'Дмитрий М.', last: 'Подскажите про парковку', when: 'Вчера', tone: 'plum', unread: 0 },
          ].map((c, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 10px', borderTop: i ? '1px solid rgba(15,23,42,0.06)' : '0' }}>
              <MAvatar name={c.name} size={42} tone={c.tone} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                  <span style={{ fontSize: 14, fontWeight: 700 }}>{c.name}</span>
                  <span style={{ fontSize: 11, color: '#5f6877' }}>{c.when}</span>
                </div>
                <div style={{ fontSize: 12, color: c.unread ? '#16212b' : '#5f6877', fontWeight: c.unread ? 600 : 400, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>{c.last}</div>
                {c.badge && <div style={{ marginTop: 4 }}><MTag tone={c.tag}>{c.badge}</MTag></div>}
              </div>
              {c.unread > 0 && (
                <div style={{ width: 22, height: 22, borderRadius: 999, background: '#0a5f59', color: '#fffdf8', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 11, fontWeight: 800 }}>
                  {c.unread}
                </div>
              )}
            </div>
          ))}
        </MCard>
      </div>

      <MTabBar items={OWNER_TABS} active="inbox" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════
// OWNER · 04 ANALYTICS
// ════════════════════════════════════════════════════════════
function MOwnerStats() {
  const data = [42, 38, 64, 58, 72, 88, 96, 78, 64, 88, 104, 92, 116, 128];
  const max = 130;
  return (
    <MPage>
      <MHeader eyebrow="Аналитика" title="14 дней" sub="Динамика выручки и загрузки. Сравнение с предыдущим периодом." />

      <div style={{ padding: '0 20px 14px' }}>
        <MCard padding={18} radius={24}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 10 }}>
            <div>
              <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Выручка</div>
              <div style={{ fontSize: 30, fontWeight: 800, letterSpacing: '-0.04em' }}>1 128 480 ₽</div>
              <div style={{ fontSize: 12, color: '#15803d', fontWeight: 700 }}>+ 18 % к прошлым 14 дн.</div>
            </div>
            <MTag tone="primary">14 дней</MTag>
          </div>
          <svg width="100%" height="120" viewBox="0 0 320 120" preserveAspectRatio="none">
            <defs>
              <linearGradient id="rev2" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0" stopColor="#0f766e" stopOpacity="0.30"/>
                <stop offset="1" stopColor="#0f766e" stopOpacity="0"/>
              </linearGradient>
            </defs>
            {(() => {
              const W = 320, H = 120, step = W / (data.length - 1);
              const pts = data.map((v, i) => [i * step, H - (v / max) * H]);
              const path = pts.map((p, i) => `${i ? 'L' : 'M'}${p[0]},${p[1]}`).join(' ');
              return (
                <>
                  <path d={`${path} L${W},${H} L0,${H} Z`} fill="url(#rev2)" />
                  <path d={path} stroke="#0a5f59" strokeWidth="2.5" fill="none" strokeLinecap="round" strokeLinejoin="round"/>
                </>
              );
            })()}
          </svg>
        </MCard>
      </div>

      <div style={{ padding: '0 20px 14px', display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
        {[
          ['Загрузка', '74 %', '+6 %', 'green'],
          ['Средний чек', '14 320 ₽', '+12 %', 'green'],
          ['Отказы', '4 %', '−2 %', 'green'],
          ['Длит. брони', '4.2 ч', '0', 'default'],
        ].map(([l, v, d, t]) => (
          <MCard key={l} padding={14} radius={18}>
            <div style={{ fontSize: 10, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>{l}</div>
            <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em', marginTop: 4 }}>{v}</div>
            <MTag tone={t} style={{ marginTop: 6 }}>{d}</MTag>
          </MCard>
        ))}
      </div>

      <div style={{ padding: '0 20px 14px' }}>
        <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>Топ-объекты</span>
        <MCard padding={10} radius={22} style={{ marginTop: 10 }}>
          {[
            { name: 'Берёзовая роща · Парная', occ: 88, rev: '64 200' },
            { name: 'Сосновый берег · Бассейн', occ: 72, rev: '48 800' },
            { name: 'Купеческая · Хамам', occ: 36, rev: '15 480' },
          ].map((o, i) => (
            <div key={i} style={{ padding: '10px 6px', borderTop: i ? '1px solid rgba(15,23,42,0.06)' : '0' }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline', marginBottom: 6 }}>
                <span style={{ fontSize: 13, fontWeight: 700 }}>{o.name}</span>
                <span style={{ fontSize: 13, fontWeight: 800 }}>{o.rev} ₽</span>
              </div>
              <div style={{ height: 6, borderRadius: 999, background: 'rgba(15,23,42,0.06)', overflow: 'hidden' }}>
                <div style={{ width: `${o.occ}%`, height: '100%', background: 'linear-gradient(90deg,#0f766e,#0a5f59)' }} />
              </div>
            </div>
          ))}
        </MCard>
      </div>

      <MTabBar items={OWNER_TABS} active="stats" />
    </MPage>
  );
}

window.RHM = window.RHM || {};
Object.assign(window.RHM, { MOwnerToday, MOwnerCalendar, MOwnerInbox, MOwnerStats });
