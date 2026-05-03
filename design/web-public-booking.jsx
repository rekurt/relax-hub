/* global React */
const { Icon, BrandLockup, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;
const { SectionHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// PUBLIC · 04 BOOKING CHECKOUT
// ════════════════════════════════════════════════════════════════
function PageBooking() {
  return (
    <div className="rh-page" style={{ minHeight: 1500 }}>
      <TopNav active="catalogue" authed />

      <Section padding="22px 28px 0">
        <div style={{ fontSize: 12, color: '#5f6877' }}>
          <a href="#" style={{ color: '#5f6877', textDecoration: 'none' }}>Берёзовая роща</a> · <span style={{ fontWeight: 700, color: '#16212b' }}>Бронирование</span>
        </div>
        <h1 style={{ margin: '12px 0 6px', fontSize: 36, fontWeight: 800, letterSpacing: '-0.04em' }}>Подтвердите бронирование</h1>
        <p style={{ margin: 0, fontSize: 14, color: '#5f6877', maxWidth: 560 }}>4 шага: контакты, оплата, дополнительно и подтверждение SMS-кодом. Без скрытых платежей.</p>
      </Section>

      {/* Stepper */}
      <Section padding="22px 28px 14px">
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          {[
            ['Контакты', true, true],
            ['Оплата', true, false],
            ['Дополнительно', false, false],
            ['Подтверждение', false, false],
          ].map(([label, active, done], i) => (
            <React.Fragment key={i}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span style={{
                  width: 30, height: 30, borderRadius: 999,
                  background: done ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : active ? '#fff' : 'rgba(255,255,255,0.6)',
                  border: active && !done ? '2px solid #0a5f59' : '1px solid rgba(15,23,42,0.10)',
                  color: done ? '#fff' : active ? '#0a5f59' : '#5f6877',
                  fontWeight: 800, fontSize: 13,
                  display: 'flex', alignItems: 'center', justifyContent: 'center',
                  boxShadow: active && !done ? '0 6px 14px rgba(15,118,110,0.20)' : 'none',
                }}>
                  {done ? <Icon name="check" size={14} /> : i + 1}
                </span>
                <span style={{ fontSize: 13, fontWeight: active ? 700 : 500, color: active ? '#16212b' : '#5f6877' }}>{label}</span>
              </div>
              {i < 3 ? <div style={{ flex: 1, height: 2, background: i < 1 ? 'linear-gradient(90deg, #0a5f59, rgba(15,118,110,0.16))' : 'rgba(15,23,42,0.08)', borderRadius: 999, maxWidth: 80 }} /> : null}
            </React.Fragment>
          ))}
        </div>
      </Section>

      <Section padding="14px 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 0.9fr', gap: 24, alignItems: 'flex-start' }}>
          {/* Left: form */}
          <div style={{ display: 'flex', flexDirection: 'column', gap: 18 }}>
            <div className="rh-card" style={{ padding: 28, borderRadius: 28 }}>
              <SectionHeader eyebrow="01 · Гость" title="Контакты для подтверждения" />
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 14, marginTop: 18 }}>
                <Field label="Имя" value="Анна" />
                <Field label="Фамилия" value="Поляк" />
                <Field label="Телефон" value="+7 (903) 555-12-44" focus />
                <Field label="Email" value="anna@relaxhub.ru" />
              </div>
              <div style={{ display: 'flex', gap: 10, marginTop: 14, padding: 12, borderRadius: 16, background: 'rgba(15,118,110,0.06)', border: '1px solid rgba(15,118,110,0.16)' }}>
                <Icon name="shield" size={16} style={{ color: '#0a5f59', marginTop: 2 }} />
                <span style={{ fontSize: 12, color: '#0a5f59', lineHeight: 1.5, fontWeight: 600 }}>SMS-код придёт на номер +7 (903) 555-12-44 после нажатия «Забронировать». Без оплаты сейчас — счёт после подтверждения хозяином.</span>
              </div>
            </div>

            <div className="rh-card" style={{ padding: 28, borderRadius: 28 }}>
              <SectionHeader eyebrow="02 · Оплата" title="Способ оплаты" />
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12, marginTop: 18 }}>
                {[
                  ['Карта', 'card', '•• 4422', true],
                  ['СБП', 'qr', 'Сбер · быстро', false],
                  ['RH-кошелёк', 'gift', '4 280 ₽', false],
                ].map(([t, ic, sub, on]) => (
                  <button key={t} style={{
                    padding: 16, borderRadius: 18, cursor: 'pointer', textAlign: 'left',
                    background: on ? 'linear-gradient(180deg, rgba(15,118,110,0.10), rgba(255,255,255,0.6))' : 'rgba(255,255,255,0.6)',
                    border: on ? '1.5px solid rgba(15,118,110,0.40)' : '1px solid rgba(15,23,42,0.08)',
                    display: 'flex', flexDirection: 'column', gap: 8,
                  }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Icon name={ic} size={20} style={{ color: '#0a5f59' }} />
                      {on ? <span style={{ width: 18, height: 18, borderRadius: 999, background: 'linear-gradient(135deg,#0f766e,#0a5f59)', color: '#fff', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icon name="check" size={11} /></span> : null}
                    </div>
                    <div>
                      <div style={{ fontSize: 14, fontWeight: 700 }}>{t}</div>
                      <div style={{ fontSize: 12, color: '#5f6877' }}>{sub}</div>
                    </div>
                  </button>
                ))}
              </div>
              <label style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 16, fontSize: 13, color: '#16212b', cursor: 'pointer' }}>
                <span style={{ width: 18, height: 18, borderRadius: 5, background: 'linear-gradient(135deg,#0f766e,#0a5f59)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icon name="check" size={11} style={{ color: '#fff' }} /></span>
                <span>Сохранить карту для будущих бронирований</span>
              </label>
            </div>

            <div className="rh-card" style={{ padding: 28, borderRadius: 28 }}>
              <SectionHeader eyebrow="03 · Дополнительно" title="Усильте программу — необязательно" />
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, marginTop: 18 }}>
                {[
                  ['Чан с подогревом', '+ 1 200 ₽', 'fire', true],
                  ['Веник дубовый × 4', '+ 600 ₽', 'plus', false],
                  ['Кейтеринг «Мангал»', 'от 1 800 ₽', 'plus', false],
                  ['Массаж 60 мин', 'от 3 500 ₽', 'plus', false],
                ].map(([t, p, ic, on]) => (
                  <div key={t} style={{
                    padding: 14, borderRadius: 18,
                    background: on ? 'linear-gradient(180deg, rgba(15,118,110,0.08), rgba(255,255,255,0.6))' : 'rgba(255,255,255,0.6)',
                    border: on ? '1.5px solid rgba(15,118,110,0.30)' : '1px solid rgba(15,23,42,0.08)',
                    display: 'flex', alignItems: 'center', gap: 12,
                  }}>
                    <div style={{ width: 38, height: 38, borderRadius: 12, background: 'rgba(15,118,110,0.10)', color: '#0a5f59', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                      <Icon name={ic} size={18} />
                    </div>
                    <div style={{ flex: 1, lineHeight: 1.25 }}>
                      <div style={{ fontSize: 14, fontWeight: 700 }}>{t}</div>
                      <div style={{ fontSize: 12, color: '#5f6877' }}>{p}</div>
                    </div>
                    <button className={on ? 'rh-btn rh-btn--primary rh-btn--sm' : 'rh-btn rh-btn--default rh-btn--sm'}>{on ? 'Добавлено' : 'Добавить'}</button>
                  </div>
                ))}
              </div>

              <Field label="Промокод" value="" placeholder="Введите промокод" suffix="Применить" style={{ marginTop: 18 }} />

              <Field label="Комментарий хозяину" value="" placeholder="Например: приедем на двух машинах, нужна детская скамеечка" textarea style={{ marginTop: 14 }} />
            </div>
          </div>

          {/* Right: order summary */}
          <div style={{ position: 'sticky', top: 90, display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div className="rh-card" style={{ padding: 22, borderRadius: 28 }}>
              <div style={{ display: 'flex', gap: 12, alignItems: 'center', marginBottom: 18 }}>
                <PhotoPlaceholder width={64} height={64} radius={16} variant="banya" label="" />
                <div>
                  <div style={{ fontSize: 15, fontWeight: 800 }}>Берёзовая роща</div>
                  <Stars value={4.9} count={128} />
                </div>
              </div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                <SummaryRow icon="cal" label="Сб, 20 апр 2026" />
                <SummaryRow icon="clock" label="19:00 – 23:00 · 4 часа" />
                <SummaryRow icon="user" label="4 гостя" />
                <SummaryRow icon="pin" label="Москва, Рублёвское ш., 12" />
              </div>

              <div style={{ height: 1, background: 'rgba(15,23,42,0.08)', margin: '16px 0' }} />

              <div style={{ display: 'flex', flexDirection: 'column', gap: 6, fontSize: 13 }}>
                <PriceLine l="3 200 ₽ × 4 ч" v="12 800 ₽" />
                <PriceLine l="Last minute -20 %" v="−2 560 ₽" tone="green" />
                <PriceLine l="Чан с подогревом" v="+1 200 ₽" />
                <PriceLine l="Депозит (возврат)" v="5 000 ₽" muted />
              </div>

              <div style={{ height: 1, background: 'rgba(15,23,42,0.08)', margin: '14px 0' }} />
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
                <span style={{ fontSize: 14, fontWeight: 700 }}>К оплате</span>
                <span style={{ fontSize: 28, fontWeight: 800, letterSpacing: '-0.03em' }}>11 440 ₽</span>
              </div>

              <button className="rh-btn rh-btn--primary rh-btn--lg" style={{ width: '100%', marginTop: 16 }}>
                Забронировать <Icon name="arrow" size={16} />
              </button>
              <p style={{ fontSize: 11, color: '#5f6877', textAlign: 'center', marginTop: 10, lineHeight: 1.45 }}>
                Нажимая «Забронировать», вы принимаете <a href="#" style={{ color: '#0a5f59' }}>правила площадки</a> и политику отмены.
              </p>
            </div>

            <div className="rh-card rh-card--flat" style={{ padding: 16, borderRadius: 20, display: 'flex', gap: 10 }}>
              <Icon name="shield" size={18} style={{ color: '#0a5f59', marginTop: 2 }} />
              <div>
                <div style={{ fontSize: 13, fontWeight: 700 }}>Возврат до 24 часов</div>
                <div style={{ fontSize: 12, color: '#5f6877', lineHeight: 1.5 }}>Полный возврат при отмене за сутки. Меньше — 50 % депозита.</div>
              </div>
            </div>
          </div>
        </div>
      </Section>

      <Footer />
    </div>
  );
}

function Field({ label, value, placeholder, focus, suffix, textarea, style = {} }) {
  return (
    <label style={{ display: 'flex', flexDirection: 'column', gap: 6, ...style }}>
      <span style={{ fontSize: 12, fontWeight: 700, color: '#16212b' }}>{label}</span>
      <div style={{
        padding: textarea ? '12px 14px' : '0 14px',
        height: textarea ? 'auto' : 50, minHeight: textarea ? 96 : 50,
        borderRadius: 16,
        border: focus ? '1.5px solid rgba(15,118,110,0.52)' : '1px solid rgba(15,23,42,0.14)',
        background: focus ? '#fff' : 'linear-gradient(180deg, rgba(255,255,255,0.98), rgba(251,247,240,0.96))',
        boxShadow: focus ? '0 0 0 4px rgba(15,118,110,0.10), 0 18px 32px rgba(15,23,42,0.08)' : 'inset 0 1px 0 rgba(255,255,255,0.88), 0 10px 24px rgba(15,23,42,0.04)',
        display: 'flex', alignItems: textarea ? 'flex-start' : 'center', gap: 10,
        fontSize: 15, fontWeight: 500,
        transform: focus ? 'translateY(-1px)' : 'none',
        transition: 'all 0.2s ease',
      }}>
        <span style={{ color: value ? '#16212b' : '#7a8390', flex: 1 }}>{value || placeholder}</span>
        {suffix ? <button style={{ background: 'none', border: 0, color: '#0a5f59', fontWeight: 700, fontSize: 13, cursor: 'pointer' }}>{suffix}</button> : null}
      </div>
    </label>
  );
}

function SummaryRow({ icon, label }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 13, color: '#16212b' }}>
      <Icon name={icon} size={14} style={{ color: '#5f6877' }} />
      <span>{label}</span>
    </div>
  );
}

function PriceLine({ l, v, tone, muted }) {
  const colors = { green: '#15803d' };
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', color: muted ? '#5f6877' : tone ? colors[tone] : '#16212b', fontWeight: tone ? 700 : 500 }}>
      <span>{l}</span><span>{v}</span>
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageBooking });
