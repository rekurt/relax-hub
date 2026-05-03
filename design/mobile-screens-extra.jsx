/* global React */
// Additional mobile screens — onboarding/auth/booking-flow/QR/chat/wallet/owner-inbox
const { MIcon, MPage, MCard, MHeader, MTag, MButton, MPhoto, MAvatar, MTabBar, MStars } = window.RHM;

// ════════════════════════════════════════════════════════════════
// M · SPLASH / WELCOME
// ════════════════════════════════════════════════════════════════
function MSplash() {
  return (
    <div style={{ minHeight: '100%', position: 'relative', overflow: 'hidden', background: 'linear-gradient(170deg, #10313a 0%, #38606a 50%, #9a5c30 100%)', color: '#fffdf8', display: 'flex', flexDirection: 'column' }}>
      <div style={{ position: 'absolute', inset: 0, opacity: 0.18 }}>
        <svg width="100%" height="100%"><circle cx="20%" cy="20%" r="120" stroke="#fff" strokeWidth="1" fill="none"/><circle cx="80%" cy="78%" r="160" stroke="#fff" strokeWidth="1" fill="none"/><circle cx="60%" cy="40%" r="60" stroke="#fff" strokeWidth="1" fill="none"/></svg>
      </div>
      {/* Status bar spacer */}
      <div style={{ height: 54 }} />

      {/* Brand */}
      <div style={{ position: 'relative', padding: '28px 24px 0' }}>
        <div style={{ fontSize: 12, fontWeight: 800, letterSpacing: '0.18em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>Бронирование бань</div>
        <div style={{ fontSize: 36, fontWeight: 800, letterSpacing: '-0.05em', marginTop: 4 }}>RelaxHUB</div>
      </div>

      {/* Hero photo */}
      <div style={{ position: 'relative', flex: 1, padding: '32px 24px', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <div style={{ width: 240, height: 320, borderRadius: 32, overflow: 'hidden', position: 'relative', boxShadow: '0 30px 60px rgba(0,0,0,0.45)' }}>
          <MPhoto height="100%" radius={0} variant="banya" label="" />
        </div>
        <div style={{ position: 'absolute', top: '20%', left: '6%', width: 140, height: 200, borderRadius: 24, overflow: 'hidden', transform: 'rotate(-8deg)', boxShadow: '0 24px 50px rgba(0,0,0,0.4)' }}>
          <MPhoto height="100%" radius={0} variant="forest" label="" />
        </div>
        <div style={{ position: 'absolute', bottom: '12%', right: '4%', width: 130, height: 180, borderRadius: 24, overflow: 'hidden', transform: 'rotate(7deg)', boxShadow: '0 24px 50px rgba(0,0,0,0.4)' }}>
          <MPhoto height="100%" radius={0} variant="spa" label="" />
        </div>
      </div>

      {/* Bottom card */}
      <div style={{ position: 'relative', padding: '0 20px 32px' }}>
        <h1 style={{ fontSize: 30, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.05, margin: '0 0 10px', maxWidth: 320 }}>
          Тёплый вечер<br/>в один тап.
        </h1>
        <p style={{ fontSize: 14, color: 'rgba(255,255,255,0.78)', lineHeight: 1.5, margin: '0 0 22px', maxWidth: 320 }}>
          240 проверенных бань — с настоящими фото, ценой за час и подтверждением за 5 минут.
        </p>

        {/* Pagination dots */}
        <div style={{ display: 'flex', gap: 6, marginBottom: 20 }}>
          <span style={{ width: 22, height: 4, borderRadius: 999, background: '#fffdf8' }}/>
          <span style={{ width: 4, height: 4, borderRadius: 999, background: 'rgba(255,255,255,0.4)' }}/>
          <span style={{ width: 4, height: 4, borderRadius: 999, background: 'rgba(255,255,255,0.4)' }}/>
        </div>

        <button style={{
          width: '100%', height: 54, borderRadius: 999,
          background: '#fffdf8', color: '#16212b',
          border: 0, fontSize: 16, fontWeight: 800, letterSpacing: '-0.01em',
          boxShadow: '0 18px 36px rgba(0,0,0,0.30)',
          display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 8,
        }}>Найти баню <MIcon name="arrow" size={18}/></button>
        <button style={{ width: '100%', height: 50, marginTop: 10, borderRadius: 999, background: 'transparent', border: '1.5px solid rgba(255,255,255,0.30)', color: '#fffdf8', fontSize: 14, fontWeight: 700 }}>
          У меня уже есть аккаунт
        </button>
      </div>

      {/* Home indicator */}
      <div style={{ position: 'absolute', bottom: 8, left: '50%', transform: 'translateX(-50%)', width: 134, height: 5, borderRadius: 999, background: '#fffdf8', opacity: 0.9 }} />
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// M · OTP / VERIFICATION
// ════════════════════════════════════════════════════════════════
function MOtp() {
  return (
    <MPage>
      <div style={{ height: 54 }} />
      <div style={{ padding: '12px 20px 0' }}>
        <button style={{ width: 38, height: 38, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 6px 14px rgba(15,23,42,0.06)' }}>
          <MIcon name="arrl" size={18}/>
        </button>
      </div>

      <MHeader eyebrow="Шаг 2 из 3" title="Введите код" sub="Отправили SMS на +7 (903) 555-12-44" />

      <div style={{ padding: '8px 20px 0' }}>
        <div style={{ display: 'flex', gap: 10, justifyContent: 'space-between' }}>
          {['4', '8', '2', '·', '·', '·'].map((d, i) => (
            <div key={i} style={{
              flex: 1, height: 64, borderRadius: 18,
              background: i < 3 ? '#fff' : 'rgba(255,255,255,0.78)',
              border: i === 3 ? '2px solid #0a5f59' : '1.5px solid rgba(15,23,42,0.12)',
              boxShadow: i < 3 ? '0 6px 14px rgba(15,23,42,0.06)' : i === 3 ? '0 0 0 4px rgba(15,118,110,0.10)' : 'inset 0 1px 0 rgba(255,255,255,0.9)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              fontSize: 26, fontWeight: 800, letterSpacing: '-0.02em',
              color: i < 3 ? '#16212b' : '#5f6877',
            }}>{i === 3 ? <span style={{ width: 2, height: 28, background: '#0a5f59', borderRadius: 1 }} /> : d}</div>
          ))}
        </div>

        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginTop: 16 }}>
          <span style={{ fontSize: 13, color: '#5f6877' }}>Не пришёл? <a href="#" style={{ color: '#0a5f59', fontWeight: 700, textDecoration: 'none' }}>Звонок голосом</a></span>
          <span style={{ fontSize: 13, color: '#5f6877' }}>00:42</span>
        </div>

        <MCard style={{ marginTop: 22, padding: 14, display: 'flex', alignItems: 'flex-start', gap: 10 }}>
          <span style={{ width: 36, height: 36, borderRadius: 12, background: 'rgba(15,118,110,0.10)', color: '#0a5f59', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}><MIcon name="shield" size={18}/></span>
          <div style={{ fontSize: 13, color: '#5f6877', lineHeight: 1.5 }}>
            <b style={{ color: '#16212b' }}>Подскажу, если код придёт.</b> На iOS код подставится автоматически из SMS — нажимать ничего не надо.
          </div>
        </MCard>

        {/* Numpad */}
        <div style={{ marginTop: 24, display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
          {['1', '2', '3', '4', '5', '6', '7', '8', '9', '', '0', '⌫'].map((k, i) => (
            <button key={i} style={{
              height: 60, borderRadius: 16,
              background: k === '' ? 'transparent' : '#fff',
              border: k === '' ? 0 : '1px solid rgba(15,23,42,0.08)',
              fontSize: 24, fontWeight: 700, color: '#16212b',
              boxShadow: k === '' ? 'none' : '0 4px 10px rgba(15,23,42,0.04)',
            }}>{k}</button>
          ))}
        </div>
      </div>

      <div style={{ position: 'absolute', bottom: 8, left: '50%', transform: 'translateX(-50%)', width: 134, height: 5, borderRadius: 999, background: '#16212b', opacity: 0.85 }} />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════════
// M · BOOKING FLOW (date + slot + guests + price)
// ════════════════════════════════════════════════════════════════
function MBooking() {
  return (
    <MPage style={{ paddingBottom: 140 }}>
      <div style={{ height: 54 }} />
      <div style={{ padding: '10px 20px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <button style={{ width: 38, height: 38, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 6px 14px rgba(15,23,42,0.06)' }}><MIcon name="arrl" size={18}/></button>
        <span style={{ fontSize: 13, fontWeight: 800, color: '#5f6877', letterSpacing: '0.08em', textTransform: 'uppercase' }}>Шаг 1/2 · Слот</span>
        <span style={{ width: 38, height: 38 }} />
      </div>

      {/* Object summary */}
      <div style={{ padding: '14px 20px 0' }}>
        <MCard padding={12} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{ width: 56, height: 56, borderRadius: 16, overflow: 'hidden', flexShrink: 0 }}>
            <MPhoto height="100%" radius={0} variant="banya" label="" />
          </div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 14, fontWeight: 800 }}>Берёзовая роща</div>
            <div style={{ fontSize: 12, color: '#5f6877', marginTop: 2 }}>3 200 ₽ / час · ★ 4.9 · 8 км</div>
          </div>
          <MTag tone="green" style={{ fontSize: 10 }}>Last min</MTag>
        </MCard>
      </div>

      {/* Calendar */}
      <div style={{ padding: '18px 20px 0' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
          <h3 style={{ margin: 0, fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>Когда?</h3>
          <span style={{ fontSize: 13, fontWeight: 700, color: '#0a5f59' }}>Апрель 2026 ›</span>
        </div>
        <MCard padding={14}>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 4, fontSize: 11, color: '#5f6877', marginBottom: 6, fontWeight: 700, textAlign: 'center' }}>
            {['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'].map(d => <span key={d}>{d}</span>)}
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(7, 1fr)', gap: 4 }}>
            {Array.from({ length: 35 }, (_, i) => {
              const day = i - 1;
              const sel = day === 19;
              const isToday = day === 16;
              const past = day < 16 || day < 1;
              const busy = [3, 9, 14, 21, 28].includes(day);
              return (
                <div key={i} style={{
                  height: 38, borderRadius: 12,
                  background: sel ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : busy ? 'rgba(180,35,24,0.06)' : 'transparent',
                  color: sel ? '#fffdf8' : past ? '#cbd2da' : busy ? '#991b1b' : '#16212b',
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  fontSize: 14, fontWeight: sel ? 800 : isToday ? 800 : 600,
                  border: isToday && !sel ? '1.5px solid #0a5f59' : 'none',
                  position: 'relative',
                  textDecoration: busy ? 'line-through' : 'none',
                }}>
                  {day >= 1 && day <= 30 ? day : ''}
                </div>
              );
            })}
          </div>
        </MCard>
      </div>

      {/* Slots */}
      <div style={{ padding: '18px 20px 0' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
          <h3 style={{ margin: 0, fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>Слот · СБ 20 апр</h3>
          <span style={{ fontSize: 13, fontWeight: 700, color: '#5f6877' }}>4 ч</span>
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 8 }}>
          {[
            ['09:00', '13:00', false, false, '12 800 ₽'],
            ['13:00', '17:00', false, false, '12 800 ₽'],
            ['15:00', '19:00', false, true, '12 800 ₽'],
            ['17:00', '21:00', false, false, '14 400 ₽'],
            ['19:00', '23:00', true, false, '12 800 ₽'],
            ['21:00', '01:00', false, false, '11 200 ₽'],
          ].map(([s, e, sel, busy, p], i) => (
            <button key={i} style={{
              padding: '12px 8px', borderRadius: 16,
              background: sel ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : busy ? 'rgba(180,35,24,0.06)' : '#fff',
              color: sel ? '#fffdf8' : busy ? '#cbd2da' : '#16212b',
              border: sel ? 0 : busy ? '1px dashed rgba(180,35,24,0.30)' : '1px solid rgba(15,23,42,0.10)',
              boxShadow: sel ? '0 12px 22px rgba(15,118,110,0.30)' : busy ? 'none' : '0 4px 10px rgba(15,23,42,0.04)',
              textDecoration: busy ? 'line-through' : 'none',
            }}>
              <div style={{ fontSize: 14, fontWeight: 800, letterSpacing: '-0.01em' }}>{s} – {e}</div>
              <div style={{ fontSize: 11, opacity: sel ? 0.8 : 0.6, marginTop: 2 }}>{busy ? 'занят' : p}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Guests */}
      <div style={{ padding: '20px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>Гости</h3>
        <MCard padding={4}>
          {[
            ['Взрослые', '13+ лет', 4, 1, 6],
            ['Дети', '3 – 12 лет', 1, 0, 4],
            ['С собакой', 'до 15 кг', 0, 0, 1],
          ].map(([t, d, val, min, max], i) => (
            <div key={t} style={{
              display: 'flex', alignItems: 'center', gap: 12, padding: '14px 14px',
              borderBottom: i < 2 ? '1px solid rgba(15,23,42,0.06)' : 'none',
            }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 14, fontWeight: 700 }}>{t}</div>
                <div style={{ fontSize: 11, color: '#5f6877' }}>{d}</div>
              </div>
              <button style={{ width: 34, height: 34, borderRadius: 999, background: val > min ? '#fff' : 'rgba(15,23,42,0.04)', border: '1px solid rgba(15,23,42,0.10)', color: val > min ? '#16212b' : '#cbd2da' }}><MIcon name="x" size={12} style={{ transform: 'rotate(45deg) scale(1.5)', strokeWidth: 2.4 }}/></button>
              <span style={{ minWidth: 24, textAlign: 'center', fontSize: 17, fontWeight: 800 }}>{val}</span>
              <button style={{ width: 34, height: 34, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)' }}><MIcon name="plus" size={14}/></button>
            </div>
          ))}
        </MCard>
      </div>

      {/* Add-ons */}
      <div style={{ padding: '20px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>Добавить к вечеру</h3>
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 8 }}>
          {[
            ['Веники дубовые ×4', '600 ₽', true],
            ['Чан с травами', '1 800 ₽', false],
            ['Шашлычный сет', '2 400 ₽', false],
            ['Массаж 60 мин', '4 200 ₽', false],
          ].map(([t, p, on]) => (
            <button key={t} style={{
              padding: '12px 14px', borderRadius: 16, textAlign: 'left',
              background: on ? 'rgba(15,118,110,0.10)' : '#fff',
              border: on ? '1.5px solid rgba(15,118,110,0.32)' : '1px solid rgba(15,23,42,0.10)',
              boxShadow: '0 4px 10px rgba(15,23,42,0.04)',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                <span style={{ fontSize: 13, fontWeight: 700 }}>{t}</span>
                <span style={{ width: 22, height: 22, borderRadius: 999, background: on ? '#0a5f59' : 'rgba(15,23,42,0.06)', color: '#fff', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}>{on ? <MIcon name="check" size={12}/> : <MIcon name="plus" size={12}/>}</span>
              </div>
              <div style={{ fontSize: 12, color: '#0a5f59', fontWeight: 700, marginTop: 4 }}>+ {p}</div>
            </button>
          ))}
        </div>
      </div>

      {/* Sticky CTA */}
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        padding: '14px 20px 28px',
        background: 'linear-gradient(180deg, rgba(248,244,236,0) 0%, rgba(248,244,236,0.96) 30%, rgba(248,244,236,1) 100%)',
        display: 'flex', alignItems: 'center', gap: 12,
      }}>
        <div>
          <div style={{ fontSize: 11, color: '#5f6877', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.08em' }}>Итого</div>
          <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em' }}>13 400 ₽</div>
          <div style={{ fontSize: 11, color: '#15803d', fontWeight: 700 }}>−2 600 ₽ last min</div>
        </div>
        <MButton tone="dark" style={{ flex: 1, height: 56, fontSize: 16 }}>К оплате <MIcon name="arrow" size={16} color="#fffdf8"/></MButton>
      </div>
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════════
// M · BOOKING SUCCESS / QR TICKET
// ════════════════════════════════════════════════════════════════
function MTicket() {
  return (
    <div style={{ minHeight: '100%', position: 'relative', background: 'radial-gradient(circle at 50% 30%, rgba(15,118,110,0.20), transparent 50%), #f8f4ec', display: 'flex', flexDirection: 'column' }}>
      <div style={{ height: 54 }} />
      <div style={{ padding: '12px 20px', display: 'flex', justifyContent: 'space-between' }}>
        <span style={{ width: 38, height: 38 }} />
        <button style={{ width: 38, height: 38, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}><MIcon name="x" size={16}/></button>
      </div>

      {/* Confetti dots */}
      <svg width="100%" height="160" style={{ position: 'absolute', top: 60, left: 0, opacity: 0.7, pointerEvents: 'none' }}>
        {Array.from({ length: 24 }, (_, i) => {
          const x = (i * 17.3 + i * i) % 100;
          const y = (i * 7.7) % 140;
          const colors = ['#0a5f59', '#d97706', '#2563eb', '#15803d'];
          return <circle key={i} cx={`${x}%`} cy={y} r={2 + (i % 3)} fill={colors[i % 4]} />;
        })}
      </svg>

      <div style={{ position: 'relative', textAlign: 'center', padding: '20px 24px 0' }}>
        <div style={{ width: 80, height: 80, borderRadius: 999, background: 'linear-gradient(135deg, #15803d, #0a5f59)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', boxShadow: '0 18px 36px rgba(21,128,61,0.30)' }}>
          <MIcon name="check" size={42} color="#fffdf8" style={{ strokeWidth: 2.4 }} />
        </div>
        <h1 style={{ margin: '20px 0 6px', fontSize: 30, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.05 }}>До встречи в субботу!</h1>
        <p style={{ margin: 0, fontSize: 14, color: '#5f6877' }}>Бронь #RHB-24018 подтверждена</p>
      </div>

      {/* Ticket card */}
      <div style={{ padding: '22px 20px 0' }}>
        <div style={{
          background: '#fffdf8', borderRadius: 28, padding: 0, overflow: 'hidden',
          boxShadow: '0 30px 60px rgba(15,23,42,0.18)',
          border: '1px solid rgba(15,23,42,0.06)',
        }}>
          <div style={{ padding: 18, background: 'linear-gradient(135deg, #10313a, #38606a 60%, #9a5c30)', color: '#fffdf8' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', opacity: 0.74 }}>Электронный билет</div>
                <div style={{ fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em', marginTop: 4 }}>Берёзовая роща</div>
                <div style={{ fontSize: 12, opacity: 0.74, marginTop: 2 }}>Москва, Рублёвское ш., 12</div>
              </div>
              <span style={{ padding: '4px 10px', borderRadius: 999, background: 'rgba(255,255,255,0.16)', fontSize: 10, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase' }}>RHB-24018</span>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10, marginTop: 16 }}>
              {[['Дата', 'СБ · 20 апр'], ['Время', '19:00 – 23:00'], ['Гостей', '4 + 1 ребёнок'], ['Сумма', '10 840 ₽']].map(([k, v]) => (
                <div key={k}>
                  <div style={{ fontSize: 10, opacity: 0.74, fontWeight: 800, letterSpacing: '0.08em', textTransform: 'uppercase' }}>{k}</div>
                  <div style={{ fontSize: 14, fontWeight: 700, marginTop: 2 }}>{v}</div>
                </div>
              ))}
            </div>
          </div>

          {/* Perforation */}
          <div style={{ position: 'relative', height: 22 }}>
            <div style={{ position: 'absolute', left: -12, top: -11, width: 24, height: 24, borderRadius: 999, background: '#f8f4ec' }}/>
            <div style={{ position: 'absolute', right: -12, top: -11, width: 24, height: 24, borderRadius: 999, background: '#f8f4ec' }}/>
            <div style={{ position: 'absolute', left: 14, right: 14, top: '50%', borderTop: '2px dashed rgba(15,23,42,0.16)' }}/>
          </div>

          <div style={{ padding: 18 }}>
            <div style={{ width: 168, height: 168, margin: '0 auto', padding: 6, background: '#fffdf8', borderRadius: 16, border: '1px solid rgba(15,23,42,0.08)' }}>
              <FakeMobileQR />
            </div>
            <div style={{ textAlign: 'center', fontSize: 11, fontWeight: 800, letterSpacing: '0.12em', textTransform: 'uppercase', color: '#5f6877', marginTop: 10 }}>Покажите на ресепшн</div>
          </div>
        </div>
      </div>

      {/* Quick actions */}
      <div style={{ padding: '20px 20px 8px', display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 10 }}>
        {[
          [<MIcon name="msg" size={18}/>, 'Чат хост', 'teal'],
          [<MIcon name="cal" size={18}/>, 'В календарь', 'gold'],
          [<MIcon name="map" size={18}/>, 'Маршрут', 'blue'],
        ].map(([ic, t, tone], i) => {
          const colors = { teal: ['#0a5f59', 'rgba(15,118,110,0.10)'], gold: ['#92400e', 'rgba(217,119,6,0.10)'], blue: ['#1d4ed8', 'rgba(37,99,235,0.10)'] };
          const [c, bg] = colors[tone];
          return (
            <button key={i} style={{ padding: '14px 8px', borderRadius: 18, background: bg, border: 0, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 6, color: c, fontSize: 12, fontWeight: 700 }}>
              {ic}{t}
            </button>
          );
        })}
      </div>

      <div style={{ padding: '12px 20px 32px' }}>
        <button style={{ width: '100%', height: 50, borderRadius: 999, background: 'transparent', border: 0, color: '#16212b', fontSize: 14, fontWeight: 800, textDecoration: 'underline' }}>На главный экран</button>
      </div>

      <div style={{ position: 'absolute', bottom: 8, left: '50%', transform: 'translateX(-50%)', width: 134, height: 5, borderRadius: 999, background: '#16212b', opacity: 0.85 }} />
    </div>
  );
}

function FakeMobileQR() {
  const cells = [];
  const seed = 'RHB24018MOBILE';
  for (let r = 0; r < 19; r++) {
    for (let c = 0; c < 19; c++) {
      const corner = (r < 7 && c < 7) || (r < 7 && c > 11) || (r > 11 && c < 7);
      const innerCorner = (r >= 1 && r <= 5 && c >= 1 && c <= 5) || (r >= 1 && r <= 5 && c >= 13 && c <= 17) || (r >= 13 && r <= 17 && c >= 1 && c <= 5);
      const innerDot = (r >= 2 && r <= 4 && c >= 2 && c <= 4) || (r >= 2 && r <= 4 && c >= 14 && c <= 16) || (r >= 14 && r <= 16 && c >= 2 && c <= 4);
      let on;
      if (innerDot) on = true;
      else if (innerCorner) on = false;
      else if (corner) on = true;
      else {
        const ch = seed.charCodeAt((r * 19 + c) % seed.length);
        on = ((ch + r * 5 + c * 3) % 7) > 3;
      }
      cells.push(<rect key={`${r}-${c}`} x={c * 8} y={r * 8} width="8" height="8" fill={on ? '#16212b' : 'transparent'} />);
    }
  }
  return <svg viewBox="0 0 152 152" width="100%" height="100%">{cells}</svg>;
}

// ════════════════════════════════════════════════════════════════
// M · CHAT (with host)
// ════════════════════════════════════════════════════════════════
function MChat() {
  return (
    <MPage style={{ paddingBottom: 100 }}>
      <div style={{ height: 54 }} />
      {/* Header */}
      <div style={{ padding: '10px 16px 12px', display: 'flex', alignItems: 'center', gap: 10, background: 'rgba(255,253,247,0.92)', backdropFilter: 'blur(16px)', borderBottom: '1px solid rgba(15,23,42,0.06)', position: 'sticky', top: 54, zIndex: 2 }}>
        <button style={{ width: 36, height: 36, borderRadius: 999, background: 'transparent', border: 0 }}><MIcon name="arrl" size={20}/></button>
        <MAvatar name="Иван Соколов" size={40} tone="amber" />
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ fontSize: 14, fontWeight: 800 }}>Иван · Берёзовая роща</div>
          <div style={{ fontSize: 11, color: '#15803d', fontWeight: 700, display: 'inline-flex', alignItems: 'center', gap: 4 }}>
            <span style={{ width: 6, height: 6, borderRadius: 999, background: '#15803d' }}/>
            онлайн · отвечает за 5 мин
          </div>
        </div>
        <button style={{ width: 36, height: 36, borderRadius: 999, background: 'rgba(15,118,110,0.10)', color: '#0a5f59', border: 0 }}><MIcon name="phone" size={16}/></button>
      </div>

      {/* Booking pill */}
      <div style={{ padding: '12px 16px 0' }}>
        <MCard padding={10} radius={14} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <div style={{ width: 38, height: 38, borderRadius: 10, overflow: 'hidden', flexShrink: 0 }}><MPhoto height="100%" radius={0} variant="banya" label=""/></div>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ fontSize: 12, fontWeight: 800 }}>Бронь #RHB-24018</div>
            <div style={{ fontSize: 11, color: '#5f6877' }}>СБ 20 апр · 19–23 · 4+1 · 10 840 ₽</div>
          </div>
          <MTag tone="green" style={{ fontSize: 10 }}>Подтв.</MTag>
        </MCard>
      </div>

      {/* Day separator */}
      <div style={{ display: 'flex', justifyContent: 'center', margin: '18px 0 6px' }}>
        <span style={{ padding: '4px 10px', borderRadius: 999, background: 'rgba(15,23,42,0.06)', fontSize: 11, fontWeight: 700, color: '#5f6877' }}>Сегодня</span>
      </div>

      {/* Messages */}
      <div style={{ padding: '0 16px', display: 'flex', flexDirection: 'column', gap: 8 }}>
        <Bubble dir="in" name="Иван">Здравствуйте, Анна! Подтверждаю бронь на субботу. Чан и веники подготовлю.</Bubble>
        <Bubble dir="in" name="Иван" stack>Парковка на 4 машины перед воротами, код шлагбаума пришлю в 18:30.</Bubble>
        <Bubble dir="out">Отлично, спасибо! А с собакой можно? Маленькая, до 8 кг.</Bubble>
        <Bubble dir="out" stack>И ещё — у мужа аллергия на дуб, можно вместо дубовых берёзовые веники?</Bubble>
        <Bubble dir="in" name="Иван">Конечно! С собакой можно, в дом не пускаем — есть тёплый вольер на террасе.</Bubble>
        <Bubble dir="in" name="Иван" stack>Берёзовые веники замочу. Без доплаты, всё включено в бронь.</Bubble>

        {/* Quick action card from host */}
        <div style={{ alignSelf: 'flex-start', maxWidth: '78%' }}>
          <MCard padding={12} radius={16} style={{ borderRadius: '4px 16px 16px 16px' }}>
            <div style={{ fontSize: 11, fontWeight: 800, color: '#0a5f59', letterSpacing: '0.06em', textTransform: 'uppercase', marginBottom: 6 }}>Иван прислал маршрут</div>
            <div style={{ height: 100, borderRadius: 10, background: 'linear-gradient(135deg, #d6e4d3, #aabea3)', position: 'relative', overflow: 'hidden' }}>
              <svg width="100%" height="100%"><path d="M20 80 Q60 40 110 60 T240 30" stroke="#16212b" strokeWidth="2.5" fill="none" strokeDasharray="4 3"/></svg>
              <div style={{ position: 'absolute', left: 12, top: 60, width: 10, height: 10, borderRadius: 999, background: '#2563eb', border: '2px solid #fff' }} />
              <div style={{ position: 'absolute', right: 16, top: 18, width: 28, height: 28, borderRadius: '50% 50% 50% 0', background: '#0a5f59', transform: 'rotate(-45deg)' }} />
            </div>
            <div style={{ fontSize: 12, marginTop: 8 }}><b>21 км · 28 минут</b><br/><span style={{ color: '#5f6877' }}>М. «Молодёжная» → Рублёвское ш., 12</span></div>
          </MCard>
        </div>

        <Bubble dir="out">Супер, спасибо большое! Жду субботы 🌿</Bubble>

        {/* Typing */}
        <div style={{ alignSelf: 'flex-start', display: 'flex', alignItems: 'center', gap: 6, padding: '10px 14px', background: 'rgba(255,255,255,0.86)', borderRadius: '4px 16px 16px 16px', border: '1px solid rgba(15,23,42,0.08)' }}>
          {[0, 1, 2].map(i => <span key={i} style={{ width: 6, height: 6, borderRadius: 999, background: '#5f6877', opacity: 0.4 + i * 0.2 }}/>)}
        </div>
      </div>

      {/* Quick replies */}
      <div style={{ padding: '14px 16px 0', display: 'flex', gap: 6, overflowX: 'auto' }}>
        {['Когда заезд?', 'Где парковка?', 'Можно с детьми?', 'Сколько идти от метро?'].map(t => (
          <span key={t} style={{ padding: '7px 12px', borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)', fontSize: 12, fontWeight: 600, whiteSpace: 'nowrap', color: '#0a5f59' }}>{t}</span>
        ))}
      </div>

      {/* Composer */}
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        padding: '10px 16px 28px',
        background: 'linear-gradient(180deg, rgba(248,244,236,0) 0%, rgba(248,244,236,0.98) 30%)',
        display: 'flex', alignItems: 'center', gap: 8,
      }}>
        <button style={{ width: 40, height: 40, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)' }}><MIcon name="plus" size={18}/></button>
        <div style={{ flex: 1, height: 44, padding: '0 16px', borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)', display: 'flex', alignItems: 'center', fontSize: 14, color: '#5f6877' }}>Сообщение…</div>
        <button style={{ width: 44, height: 44, borderRadius: 999, background: 'linear-gradient(135deg,#0f766e,#0a5f59)', border: 0, color: '#fffdf8', boxShadow: '0 8px 18px rgba(15,118,110,0.30)' }}><MIcon name="arrow" size={18} color="#fffdf8"/></button>
      </div>
    </MPage>
  );
}

function Bubble({ dir, name, stack, children }) {
  const out = dir === 'out';
  return (
    <div style={{ alignSelf: out ? 'flex-end' : 'flex-start', maxWidth: '78%' }}>
      {!out && !stack && <div style={{ fontSize: 10, color: '#5f6877', fontWeight: 700, marginBottom: 2, paddingLeft: 14 }}>{name}</div>}
      <div style={{
        padding: '10px 14px',
        borderRadius: out
          ? (stack ? '18px 4px 4px 18px' : '18px 18px 4px 18px')
          : (stack ? '4px 18px 18px 4px' : '4px 18px 18px 18px'),
        background: out ? 'linear-gradient(135deg, #16212b, #2c3a48)' : '#fffdf8',
        color: out ? '#fffdf8' : '#16212b',
        border: out ? 0 : '1px solid rgba(15,23,42,0.08)',
        boxShadow: out ? '0 8px 18px rgba(15,23,42,0.18)' : '0 4px 10px rgba(15,23,42,0.04)',
        fontSize: 14, lineHeight: 1.45,
      }}>{children}</div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// M · WALLET & PROMO
// ════════════════════════════════════════════════════════════════
function MWallet() {
  return (
    <MPage>
      <div style={{ height: 54 }} />
      <MHeader eyebrow="Кошелёк" title="Бонусы и сертификаты" sub="Списывайте до 30 % суммы. Не сгорают." />

      {/* Hero balance */}
      <div style={{ padding: '8px 20px 0' }}>
        <div style={{
          padding: 22, borderRadius: 28,
          background: 'linear-gradient(135deg, #10313a 0%, #38606a 50%, #9a5c30 100%)',
          color: '#fffdf8', position: 'relative', overflow: 'hidden',
          boxShadow: '0 24px 48px rgba(15,23,42,0.20)',
        }}>
          <div style={{ position: 'absolute', inset: 0, opacity: 0.18 }}>
            <svg width="100%" height="100%"><circle cx="84%" cy="20%" r="80" stroke="#fff" strokeWidth="1" fill="none"/><circle cx="20%" cy="86%" r="100" stroke="#fff" strokeWidth="1" fill="none"/></svg>
          </div>
          <div style={{ position: 'relative' }}>
            <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', opacity: 0.74 }}>Доступно</div>
            <div style={{ fontSize: 44, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1, marginTop: 6 }}>2 480 ₽</div>
            <div style={{ fontSize: 12, opacity: 0.74, marginTop: 4 }}>+ 720 ₽ за следующие бронь и отзыв</div>
            <div style={{ display: 'flex', gap: 8, marginTop: 18 }}>
              <button style={{ flex: 1, padding: '11px 14px', borderRadius: 14, background: '#fffdf8', color: '#16212b', border: 0, fontSize: 13, fontWeight: 800 }}>Применить к брони</button>
              <button style={{ width: 44, height: 44, borderRadius: 14, background: 'rgba(255,255,255,0.12)', color: '#fffdf8', border: '1px solid rgba(255,255,255,0.22)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center' }}><MIcon name="qr" size={18}/></button>
            </div>
          </div>
        </div>
      </div>

      {/* Promo input */}
      <div style={{ padding: '16px 20px 0' }}>
        <MCard padding={10} radius={16} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span style={{ width: 36, height: 36, borderRadius: 10, background: 'rgba(217,119,6,0.10)', color: '#92400e', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}><MIcon name="gift" size={18}/></span>
          <div style={{ flex: 1, fontSize: 13, fontWeight: 600, color: '#5f6877' }}>Промокод или сертификат</div>
          <button style={{ padding: '7px 14px', borderRadius: 999, background: 'linear-gradient(135deg,#0f766e,#0a5f59)', color: '#fffdf8', border: 0, fontSize: 12, fontWeight: 800 }}>Применить</button>
        </MCard>
      </div>

      {/* Sections */}
      <div style={{ padding: '20px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>Сертификаты · 2</h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {[
            { name: 'На день рождения', amount: '5 000 ₽', from: 'Михаил Парфёнов', exp: 'до 30.06.2026', tone: 'gold' },
            { name: 'Корпоративный 8 марта', amount: '3 000 ₽', from: 'ООО «Атмосфера»', exp: 'до 31.12.2026', tone: 'plum' },
          ].map((c, i) => {
            const tones = { gold: ['#d97706', '#b45309'], plum: ['#7c3aed', '#5b21b6'] };
            const [a, b] = tones[c.tone];
            return (
              <MCard key={i} padding={0} radius={20} style={{ overflow: 'hidden', display: 'flex', minHeight: 96 }}>
                <div style={{ width: 96, background: `linear-gradient(135deg, ${a}, ${b})`, color: '#fffdf8', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', padding: 8 }}>
                  <MIcon name="gift" size={26} color="#fffdf8" />
                  <div style={{ fontSize: 18, fontWeight: 800, marginTop: 4 }}>{c.amount}</div>
                </div>
                <div style={{ flex: 1, padding: 14, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                  <div style={{ fontSize: 14, fontWeight: 800 }}>{c.name}</div>
                  <div style={{ fontSize: 11, color: '#5f6877', marginTop: 2 }}>От {c.from} · {c.exp}</div>
                  <div style={{ display: 'flex', gap: 6, marginTop: 8 }}>
                    <MTag tone="primary" style={{ fontSize: 10 }}>Действует</MTag>
                    <MTag style={{ fontSize: 10 }}>Любой объект</MTag>
                  </div>
                </div>
              </MCard>
            );
          })}
        </div>
      </div>

      <div style={{ padding: '24px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>Реферальная программа</h3>
        <MCard padding={16} style={{ position: 'relative', overflow: 'hidden' }}>
          <div style={{ position: 'absolute', right: -20, top: -10, width: 120, height: 120, borderRadius: 999, background: 'radial-gradient(circle, rgba(15,118,110,0.16), transparent 70%)' }} />
          <div style={{ position: 'relative' }}>
            <div style={{ fontSize: 13, fontWeight: 800, color: '#0a5f59', letterSpacing: '0.06em', textTransform: 'uppercase' }}>Приведи друга</div>
            <div style={{ fontSize: 22, fontWeight: 800, letterSpacing: '-0.03em', marginTop: 4, lineHeight: 1.15 }}>+ 1 000 ₽ обоим<br/><span style={{ color: '#5f6877', fontSize: 14, fontWeight: 600 }}>после первой брони друга</span></div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 14, padding: '10px 14px', background: 'rgba(15,23,42,0.06)', borderRadius: 12 }}>
              <span style={{ flex: 1, fontFamily: 'monospace', fontSize: 14, fontWeight: 700, letterSpacing: '0.05em' }}>ANNA-RH-24</span>
              <button style={{ padding: '4px 10px', borderRadius: 8, background: '#fff', border: 0, fontSize: 12, fontWeight: 700 }}>Копировать</button>
            </div>
            <div style={{ fontSize: 11, color: '#5f6877', marginTop: 8 }}>Перешло по ссылке: 8 · бронь сделали: 3</div>
          </div>
        </MCard>
      </div>

      {/* History */}
      <div style={{ padding: '24px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 17, fontWeight: 800, letterSpacing: '-0.02em' }}>История начислений</h3>
        <MCard padding={0} style={{ overflow: 'hidden' }}>
          {[
            ['Кэшбэк за бронь #RHB-23904', '+540 ₽', '12 апр', 'green'],
            ['Бонус за отзыв «Сосновый берег»', '+200 ₽', '08 апр', 'green'],
            ['Списано к брони #RHB-23890', '−800 ₽', '05 апр', 'red'],
            ['Подарок ко дню рождения', '+1 000 ₽', '01 апр', 'gold'],
          ].map(([t, v, d, tone], i) => {
            const cmap = { green: '#15803d', red: '#991b1b', gold: '#92400e' };
            return (
              <div key={i} style={{ padding: '14px 14px', display: 'flex', alignItems: 'center', gap: 12, borderBottom: i < 3 ? '1px solid rgba(15,23,42,0.06)' : 'none' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontSize: 13, fontWeight: 700 }}>{t}</div>
                  <div style={{ fontSize: 11, color: '#5f6877' }}>{d}</div>
                </div>
                <span style={{ fontSize: 14, fontWeight: 800, color: cmap[tone] }}>{v}</span>
              </div>
            );
          })}
        </MCard>
      </div>

      <MTabBar active="profile" />
    </MPage>
  );
}

// ════════════════════════════════════════════════════════════════
// M · OWNER INBOX (request detail)
// ════════════════════════════════════════════════════════════════
function MOwnerInbox() {
  return (
    <MPage style={{ paddingBottom: 120 }}>
      <div style={{ height: 54 }} />
      <div style={{ padding: '10px 16px 0', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <button style={{ width: 38, height: 38, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)' }}><MIcon name="arrl" size={18}/></button>
        <span style={{ fontSize: 12, fontWeight: 800, color: '#5f6877', letterSpacing: '0.10em', textTransform: 'uppercase' }}>Заявка · 2 мин назад</span>
        <button style={{ width: 38, height: 38, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.10)' }}><MIcon name="more" size={18}/></button>
      </div>

      <div style={{ padding: '14px 20px 0' }}>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: 8, padding: '5px 12px', borderRadius: 999, background: 'rgba(217,119,6,0.10)', color: '#92400e', border: '1px solid rgba(217,119,6,0.28)' }}>
          <span style={{ width: 6, height: 6, borderRadius: 999, background: '#d97706' }}/>
          <span style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.10em', textTransform: 'uppercase' }}>Ждёт ответа · 28 мин</span>
        </div>
        <h1 style={{ margin: '12px 0 4px', fontSize: 28, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.05 }}>Анна Поляк хочет приехать в субботу</h1>
        <p style={{ margin: 0, fontSize: 13, color: '#5f6877', lineHeight: 1.5 }}>4 взрослых + 1 ребёнок · 19:00 – 23:00 · 10 840 ₽</p>
      </div>

      {/* Guest card */}
      <div style={{ padding: '16px 20px 0' }}>
        <MCard padding={14} style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <MAvatar name="Анна Поляк" size={52} tone="teal" />
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 15, fontWeight: 800 }}>Анна Поляк</div>
            <div style={{ fontSize: 11, color: '#5f6877' }}>На RelaxHUB с 2024 · 14 броней · ★ 4.95</div>
            <div style={{ display: 'flex', gap: 6, marginTop: 6 }}>
              <MTag tone="green" style={{ fontSize: 10 }}>Проверена</MTag>
              <MTag style={{ fontSize: 10 }}>0 отмен</MTag>
              <MTag tone="primary" style={{ fontSize: 10 }}>Постоянный гость</MTag>
            </div>
          </div>
        </MCard>
      </div>

      {/* Slot vs calendar check */}
      <div style={{ padding: '16px 20px 0' }}>
        <MCard padding={14} style={{ background: 'rgba(15,118,110,0.06)', border: '1px solid rgba(15,118,110,0.20)' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <MIcon name="checkc" size={18} color="#0a5f59"/>
            <span style={{ fontSize: 13, fontWeight: 800, color: '#0a5f59' }}>Слот свободен. Конфликтов нет.</span>
          </div>
          <div style={{ fontSize: 12, color: '#0a5f59', marginTop: 6, lineHeight: 1.5 }}>
            До: ПТ 21:00 пусто. После: ВС 09:00 — следующая бронь. Между предыдущей (15:00) и этой — 4 часа на уборку.
          </div>
        </MCard>
      </div>

      {/* Breakdown */}
      <div style={{ padding: '20px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 14, fontWeight: 800, letterSpacing: '-0.01em', color: '#5f6877', textTransform: 'uppercase' }}>Детали</h3>
        <MCard padding={0}>
          {[
            ['Объект', 'Берёзовая роща'],
            ['Слот', 'СБ 20 апр · 19:00 – 23:00 (4 ч)'],
            ['Гости', '4 взрослых + 1 ребёнок (8 лет)'],
            ['Контакт', '+7 (910) 555-24-18'],
            ['Аренда (4 ч × 3 200)', '12 800 ₽'],
            ['Скидка last-min', '−2 560 ₽'],
            ['Чан с травами + веники', '+ 600 ₽'],
            ['Итого к получению', '10 840 ₽'],
          ].map(([k, v], i, arr) => (
            <div key={k} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 14px', borderBottom: i < arr.length - 1 ? '1px solid rgba(15,23,42,0.06)' : 'none' }}>
              <span style={{ flex: 1, fontSize: 13, color: '#5f6877' }}>{k}</span>
              <span style={{ fontSize: 13, fontWeight: i === arr.length - 1 ? 800 : 700, fontSize: i === arr.length - 1 ? 16 : 13 }}>{v}</span>
            </div>
          ))}
        </MCard>
      </div>

      {/* Guest message */}
      <div style={{ padding: '16px 20px 0' }}>
        <h3 style={{ margin: '0 0 10px', fontSize: 14, fontWeight: 800, letterSpacing: '-0.01em', color: '#5f6877', textTransform: 'uppercase' }}>Сообщение</h3>
        <MCard padding={14} style={{ position: 'relative' }}>
          <div style={{ fontSize: 14, lineHeight: 1.55, color: '#16212b' }}>
            «Здравствуйте! Нас 5, один ребёнок 8 лет — это нормально? И ещё — можно с маленькой собакой? Заранее спасибо 🌿»
          </div>
          <div style={{ marginTop: 12, display: 'flex', gap: 6, flexWrap: 'wrap' }}>
            {['Да, можно ✓', 'С собакой можно', 'Подготовлю чан', 'Свяжусь по телефону'].map(t => (
              <span key={t} style={{ padding: '5px 10px', borderRadius: 999, background: 'rgba(15,118,110,0.08)', color: '#0a5f59', fontSize: 11, fontWeight: 700 }}>{t}</span>
            ))}
          </div>
        </MCard>
      </div>

      {/* Sticky CTA */}
      <div style={{
        position: 'absolute', bottom: 0, left: 0, right: 0,
        padding: '14px 16px 24px',
        background: 'linear-gradient(180deg, rgba(248,244,236,0) 0%, rgba(248,244,236,0.96) 30%)',
        display: 'grid', gridTemplateColumns: '1fr 2fr', gap: 8,
      }}>
        <button style={{ height: 54, borderRadius: 999, background: '#fff', border: '1px solid rgba(15,23,42,0.14)', fontSize: 14, fontWeight: 800, color: '#16212b' }}>Отклонить</button>
        <button style={{ height: 54, borderRadius: 999, background: 'linear-gradient(135deg, #15803d, #0a5f59)', border: 0, color: '#fffdf8', fontSize: 16, fontWeight: 800, boxShadow: '0 12px 26px rgba(15,118,110,0.30)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', gap: 8 }}>
          <MIcon name="check" size={18} color="#fffdf8" style={{ strokeWidth: 2.4 }} /> Подтвердить · 10 840 ₽
        </button>
      </div>
    </MPage>
  );
}

window.RHM = window.RHM || {};
Object.assign(window.RHM, { MSplash, MOtp, MBooking, MTicket, MChat, MWallet, MOwnerInbox });
