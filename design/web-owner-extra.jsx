/* global React */
// Owner web — inbox, finance, listing-onboarding wizard
const { Icon, Avatar, PhotoPlaceholder, Stars, OwnerNav, Footer, Section } = window.RH;

// ════════════════════════════════════════════════════════════════
// O · INBOX — заявки и подтверждения
// ════════════════════════════════════════════════════════════════
function PageOwnerInbox() {
  return (
    <div className="rh-page" style={{ minHeight: 1200 }}>
      <OwnerNav active="bookings" />
      <Section padding="28px 28px 0">
        <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', marginBottom: 20 }}>
          <div>
            <span className="rh-eyebrow">Бронирования / Заявки</span>
            <h1 className="rh-h1" style={{ marginTop: 6 }}>Inbox · 7 ждут</h1>
            <p style={{ margin: '4px 0 0', fontSize: 14, color: '#5f6877' }}>Среднее время ответа партнёров — 12 минут. Гость уходит после 30.</p>
          </div>
          <div style={{ display: 'flex', gap: 10 }}>
            <button className="rh-btn rh-btn--default"><Icon name="set" size={14}/> Авто-подтверждение</button>
            <button className="rh-btn rh-btn--primary"><Icon name="check" size={14}/> Подтвердить все совместимые</button>
          </div>
        </div>

        <div style={{ display: 'grid', gridTemplateColumns: '1fr 380px', gap: 20 }}>
          {/* Left: list of requests */}
          <div className="rh-card" style={{ padding: 0, borderRadius: 24, overflow: 'hidden' }}>
            <div style={{ display: 'flex', gap: 6, padding: 14, borderBottom: '1px solid rgba(15,23,42,0.06)' }}>
              {[['Все · 14', false], ['Заявки · 7', true], ['Подтв · 5', false], ['Отменены · 2', false]].map(([t, on]) => (
                <span key={t} style={{ padding: '7px 14px', borderRadius: 999, fontSize: 12, fontWeight: 700, background: on ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'rgba(15,23,42,0.04)', color: on ? '#fff' : '#16212b' }}>{t}</span>
              ))}
            </div>

            {[
              { name: 'Анна Поляк', avatar: 'teal', listing: 'Берёзовая роща · СБ 19:00–23:00', guests: '4 + 1 ребёнок', sum: '10 840 ₽', state: 'new', wait: '2 мин назад', sel: true, msg: 'Нас 5, 1 ребёнок — это нормально? Можно с собакой?' },
              { name: 'Михаил Парфёнов', avatar: 'amber', listing: 'Купеческая · ВС 12:00–16:00', guests: '6 гостей', sum: '8 200 ₽', state: 'new', wait: '8 мин', sel: false, msg: 'Сможете подготовить веники к 11:45?' },
              { name: 'Дарья Власова', avatar: 'plum', listing: 'Берёзовая роща · ПТ 17:00–21:00', guests: '8 гостей · корпоратив', sum: '17 600 ₽', state: 'new', wait: '23 мин', sel: false, msg: 'Нужен счёт-фактура. Юр. лицо — приложила реквизиты.' },
              { name: 'Алексей Дёмин', avatar: 'blue', listing: 'Сосновый берег · СБ 12:00–18:00', guests: '4 гостя', sum: '14 400 ₽', state: 'overdue', wait: '1 ч 02 мин', sel: false, msg: 'Доброе утро, можно подтвердить?' },
              { name: 'Мария Лисс', avatar: 'plum', listing: 'Берёзовая роща · ВС 17:00–20:00', guests: '2 гостя', sum: '9 600 ₽', state: 'new', wait: '34 мин', sel: false, msg: 'Романтический вечер, можно ли свечи и шампанское?' },
            ].map((r, i) => (
              <button key={i} style={{
                width: '100%', padding: '16px 18px', display: 'grid', gridTemplateColumns: '44px 1fr auto', gap: 14,
                background: r.sel ? 'rgba(15,118,110,0.06)' : 'transparent',
                borderLeft: r.sel ? '3px solid #0f766e' : '3px solid transparent',
                border: 0, borderBottom: '1px solid rgba(15,23,42,0.06)',
                cursor: 'pointer', textAlign: 'left',
              }}>
                <Avatar name={r.name} size={44} tone={r.avatar} />
                <div>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <span style={{ fontSize: 14, fontWeight: 800 }}>{r.name}</span>
                    {r.state === 'overdue' && <span className="rh-tag rh-tag--gold" style={{ fontSize: 10 }}>SLA: ответьте</span>}
                  </div>
                  <div style={{ fontSize: 13, color: '#5f6877', marginTop: 2 }}>{r.listing}</div>
                  <div style={{ fontSize: 12, color: '#16212b', marginTop: 4, lineHeight: 1.45 }}>«{r.msg}»</div>
                </div>
                <div style={{ textAlign: 'right' }}>
                  <div style={{ fontSize: 16, fontWeight: 800 }}>{r.sum}</div>
                  <div style={{ fontSize: 11, color: '#5f6877' }}>{r.guests}</div>
                  <div style={{ fontSize: 10, color: r.state === 'overdue' ? '#b45309' : '#5f6877', fontWeight: 700, letterSpacing: '0.06em', marginTop: 4 }}>{r.wait}</div>
                </div>
              </button>
            ))}
          </div>

          {/* Right: detail */}
          <div className="rh-card" style={{ padding: 22, borderRadius: 24, height: 'fit-content', position: 'sticky', top: 20 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <Avatar name="Анна Поляк" size={48} tone="teal" />
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 16, fontWeight: 800 }}>Анна Поляк</div>
                <div style={{ fontSize: 12, color: '#5f6877' }}>На RelaxHUB с 2024 · 14 броней · 4.95</div>
              </div>
              <button style={{ width: 32, height: 32, borderRadius: 999, background: 'rgba(15,118,110,0.10)', color: '#0a5f59', border: 0 }}><Icon name="msg" size={14}/></button>
            </div>

            <div style={{
              marginTop: 16, padding: 16, borderRadius: 18,
              background: 'rgba(255,253,247,0.86)',
              border: '1px dashed rgba(15,23,42,0.18)',
            }}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12, fontSize: 13 }}>
                <Pair k="Объект" v="Берёзовая роща" />
                <Pair k="Слот" v="20.04, СБ · 19–23" />
                <Pair k="Гости" v="4 + 1 ребёнок" />
                <Pair k="Контакт" v="+7 910 ··· 24 18" />
                <Pair k="Аренда (4 ч)" v="12 800 ₽" />
                <Pair k="Скидка last-min" v="−2 560 ₽" />
                <Pair k="Чан + веники" v="+ 600 ₽" />
                <Pair k="Итого" v={<b style={{ fontSize: 16 }}>10 840 ₽</b>} />
              </div>
            </div>

            <div style={{ marginTop: 14, padding: 12, borderRadius: 12, background: 'rgba(15,118,110,0.06)', border: '1px solid rgba(15,118,110,0.20)', fontSize: 12, color: '#0a5f59' }}>
              <Icon name="check" size={13} style={{ verticalAlign: '-2px', marginRight: 6 }} />
              <b>Слот свободен.</b> До и после — 2 ч промежуток. Конфликтов нет.
            </div>

            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginTop: 16 }}>
              <button className="rh-btn rh-btn--primary"><Icon name="check" size={14}/> Подтвердить</button>
              <button className="rh-btn rh-btn--default">Предложить другое</button>
              <button className="rh-btn rh-btn--ghost rh-btn--sm" style={{ gridColumn: '1 / -1' }}>Отклонить с причиной</button>
            </div>

            <div style={{ marginTop: 16, fontSize: 11, color: '#5f6877' }}>Шаблоны ответов:</div>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 6, marginTop: 6 }}>
              {['Подтвержу через 5 мин', 'Чан подготовим', 'Можно с собакой ✓', 'Мест нет, предложу другое'].map(t => (
                <span key={t} style={{ padding: '5px 10px', borderRadius: 999, fontSize: 11, background: 'rgba(15,23,42,0.04)', border: '1px solid rgba(15,23,42,0.08)', cursor: 'pointer' }}>{t}</span>
              ))}
            </div>
          </div>
        </div>
      </Section>
      <Footer />
    </div>
  );
}

function Pair({ k, v }) {
  return (
    <div>
      <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#5f6877' }}>{k}</div>
      <div style={{ marginTop: 2, fontSize: 13, fontWeight: 600 }}>{v}</div>
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// O · FINANCE / PAYOUTS
// ════════════════════════════════════════════════════════════════
function PageOwnerFinance() {
  return (
    <div className="rh-page" style={{ minHeight: 1300 }}>
      <OwnerNav active="payouts" />
      <Section padding="28px 28px 0">
        <span className="rh-eyebrow">Финансы / Выплаты</span>
        <h1 className="rh-h1" style={{ marginTop: 6 }}>Выручка и выплаты</h1>
        <p style={{ margin: '4px 0 22px', fontSize: 14, color: '#5f6877' }}>0 % комиссии — все деньги ваши, выплата раз в неделю по понедельникам.</p>

        <div style={{ display: 'grid', gridTemplateColumns: '1.4fr 1fr', gap: 20, marginBottom: 22 }}>
          {/* Hero card */}
          <div className="rh-card" style={{ padding: 28, borderRadius: 28, background: 'linear-gradient(135deg, #10313a 0%, #38606a 50%, #9a5c30 100%)', color: '#fffdf8', border: '1px solid rgba(255,255,255,0.12)', position: 'relative', overflow: 'hidden' }}>
            <div style={{ position: 'absolute', inset: 0, opacity: 0.18 }}>
              <svg width="100%" height="100%"><circle cx="84%" cy="20%" r="160" stroke="#fff" strokeWidth="1" fill="none"/><circle cx="14%" cy="86%" r="200" stroke="#fff" strokeWidth="1" fill="none"/></svg>
            </div>
            <div style={{ position: 'relative' }}>
              <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.14em', textTransform: 'uppercase', color: 'rgba(255,255,255,0.74)' }}>К выплате · понедельник 23 апр</div>
              <div style={{ fontSize: 56, fontWeight: 800, letterSpacing: '-0.05em', lineHeight: 1, marginTop: 10 }}>184 320 ₽</div>
              <div style={{ display: 'flex', gap: 14, marginTop: 10, fontSize: 13, color: 'rgba(255,255,255,0.84)' }}>
                <span>22 заезда · с 16 по 22 апреля</span>
                <span>·</span>
                <span>Среднее 8 374 ₽ / заезд</span>
              </div>
              <div style={{ display: 'flex', gap: 10, marginTop: 22 }}>
                <button style={{ padding: '11px 20px', borderRadius: 999, background: '#fffdf8', color: '#16212b', border: 0, fontSize: 14, fontWeight: 800, cursor: 'pointer' }}><Icon name="bolt" size={14} style={{ verticalAlign: '-2px' }}/> Вывести сейчас</button>
                <button style={{ padding: '11px 20px', borderRadius: 999, background: 'rgba(255,255,255,0.12)', color: '#fffdf8', border: '1px solid rgba(255,255,255,0.22)', fontSize: 14, fontWeight: 700, cursor: 'pointer' }}>Изменить счёт</button>
                <button style={{ padding: '11px 20px', borderRadius: 999, background: 'rgba(255,255,255,0.06)', color: '#fffdf8', border: '1px solid rgba(255,255,255,0.14)', fontSize: 14, fontWeight: 700, cursor: 'pointer' }}>Налоговая справка</button>
              </div>
            </div>
          </div>

          {/* Stats */}
          <div style={{ display: 'grid', gridTemplateRows: 'repeat(3, 1fr)', gap: 14 }}>
            {[
              ['Выручка апрель', '628 540 ₽', '+18 % к марту', '#0a5f59'],
              ['Средний чек', '8 374 ₽', '+540 ₽ к март', '#16212b'],
              ['Возвраты', '12 600 ₽ (2)', '0.7 %', '#92400e'],
            ].map(([k, v, d, c]) => (
              <div key={k} className="rh-card" style={{ padding: 18, borderRadius: 20, display: 'flex', flexDirection: 'column', justifyContent: 'center' }}>
                <span className="rh-eyebrow-muted">{k}</span>
                <div style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.03em', color: c, marginTop: 4 }}>{v}</div>
                <div style={{ fontSize: 11, color: '#5f6877', marginTop: 2 }}>{d}</div>
              </div>
            ))}
          </div>
        </div>

        {/* Chart */}
        <div className="rh-card" style={{ padding: 22, borderRadius: 24, marginBottom: 22 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 14 }}>
            <div>
              <span className="rh-eyebrow-muted">Выручка по дням</span>
              <h3 style={{ margin: '4px 0 0', fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em' }}>14 дней · 3 объекта</h3>
            </div>
            <div style={{ display: 'flex', gap: 6 }}>
              {['7 дн', '14 дн', '30 дн', '90 дн'].map((p, i) => (
                <span key={p} style={{ padding: '6px 12px', borderRadius: 999, fontSize: 12, fontWeight: 700, background: i === 1 ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'rgba(15,23,42,0.04)', color: i === 1 ? '#fff' : '#16212b' }}>{p}</span>
              ))}
            </div>
          </div>
          <RevenueChart />
          <div style={{ display: 'flex', gap: 18, fontSize: 12, color: '#5f6877', marginTop: 12 }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><span style={{ width: 12, height: 12, borderRadius: 3, background: 'linear-gradient(180deg,#0f766e,#0a5f59)' }}/>Берёзовая роща</span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><span style={{ width: 12, height: 12, borderRadius: 3, background: 'linear-gradient(180deg,#d97706,#b45309)' }}/>Купеческая</span>
            <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><span style={{ width: 12, height: 12, borderRadius: 3, background: 'linear-gradient(180deg,#2563eb,#1d4ed8)' }}/>Сосновый берег</span>
          </div>
        </div>

        {/* Payouts log */}
        <div className="rh-card" style={{ padding: 0, borderRadius: 24, overflow: 'hidden', marginBottom: 40 }}>
          <div style={{ padding: '18px 22px', display: 'flex', justifyContent: 'space-between', alignItems: 'center', borderBottom: '1px solid rgba(15,23,42,0.06)' }}>
            <h3 style={{ margin: 0, fontSize: 16, fontWeight: 800 }}>История выплат</h3>
            <a href="#" style={{ fontSize: 13, color: '#0a5f59', fontWeight: 700, textDecoration: 'none' }}>Скачать выписку →</a>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '160px 1fr 1fr 1fr 140px 80px', padding: '14px 22px', fontSize: 11, fontWeight: 800, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#5f6877', borderBottom: '1px solid rgba(15,23,42,0.06)' }}>
            <span>Дата</span><span>Период</span><span>Заездов</span><span>Счёт</span><span>Сумма</span><span>Статус</span>
          </div>
          {[
            ['16 апр 2026', '09–15 апр', '24 заезда', 'Сбер ··· 4421', '171 200 ₽', 'Готово', 'green'],
            ['09 апр 2026', '02–08 апр', '21 заезд', 'Сбер ··· 4421', '156 800 ₽', 'Готово', 'green'],
            ['02 апр 2026', '26 мар – 1 апр', '19 заездов', 'Сбер ··· 4421', '148 600 ₽', 'Готово', 'green'],
            ['26 мар 2026', '19–25 мар', '22 заезда', 'Сбер ··· 4421', '162 400 ₽', 'Возврат 12 600', 'gold'],
            ['19 мар 2026', '12–18 мар', '17 заездов', 'Тинькофф ··· 8814', '132 800 ₽', 'Готово', 'green'],
          ].map((row, i) => (
            <div key={i} style={{ display: 'grid', gridTemplateColumns: '160px 1fr 1fr 1fr 140px 80px', padding: '14px 22px', fontSize: 13, alignItems: 'center', borderBottom: i < 4 ? '1px solid rgba(15,23,42,0.06)' : 'none' }}>
              <span style={{ fontWeight: 700 }}>{row[0]}</span>
              <span style={{ color: '#5f6877' }}>{row[1]}</span>
              <span>{row[2]}</span>
              <span style={{ color: '#5f6877' }}>{row[3]}</span>
              <span style={{ fontWeight: 800, fontSize: 14 }}>{row[4]}</span>
              <span className={`rh-tag rh-tag--${row[6]}`} style={{ fontSize: 10, padding: '3px 8px' }}>{row[5]}</span>
            </div>
          ))}
        </div>
      </Section>
      <Footer />
    </div>
  );
}

function RevenueChart() {
  const days = Array.from({ length: 14 }, (_, i) => i);
  const heights = [40, 56, 36, 72, 88, 110, 78, 64, 80, 96, 72, 92, 124, 108];
  return (
    <div style={{ height: 200, display: 'flex', alignItems: 'flex-end', gap: 8, padding: '0 4px' }}>
      {days.map(d => {
        const h = heights[d];
        return (
          <div key={d} style={{ flex: 1, display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 4 }}>
            <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', height: 160, justifyContent: 'flex-end', width: '100%' }}>
              <div style={{ width: '70%', height: h, background: 'linear-gradient(180deg, #0f766e, #0a5f59)', borderRadius: '6px 6px 0 0' }} />
              <div style={{ width: '70%', height: h * 0.45, background: 'linear-gradient(180deg, #d97706, #b45309)' }} />
              <div style={{ width: '70%', height: h * 0.30, background: 'linear-gradient(180deg, #2563eb, #1d4ed8)', borderRadius: '0 0 6px 6px' }} />
            </div>
            <span style={{ fontSize: 10, color: '#5f6877' }}>{(9 + d).toString().padStart(2, '0')}.04</span>
          </div>
        );
      })}
    </div>
  );
}

// ════════════════════════════════════════════════════════════════
// O · LISTING ONBOARDING (3-step wizard)
// ════════════════════════════════════════════════════════════════
function PageOwnerOnboarding() {
  return (
    <div className="rh-page" style={{ minHeight: 1500 }}>
      <OwnerNav active="objects" />
      <Section padding="28px 28px 0">
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-end', marginBottom: 24 }}>
          <div>
            <span className="rh-eyebrow">Объекты / Новый</span>
            <h1 className="rh-h1" style={{ marginTop: 6 }}>Добавить баню — 12 минут</h1>
            <p style={{ margin: '4px 0 0', fontSize: 14, color: '#5f6877' }}>Заполняйте на сколько хватит, остальное доделаете позже. Опубликуем после модерации (до 24 часов).</p>
          </div>
          <span style={{ fontSize: 13, fontWeight: 700, color: '#5f6877' }}>Черновик · автосохранение 2 сек назад</span>
        </div>

        {/* Stepper */}
        <div className="rh-card" style={{ padding: 18, borderRadius: 20, marginBottom: 18, display: 'flex', alignItems: 'center', gap: 8 }}>
          {['Основное', 'Фото и удобства', 'Цены и календарь'].map((step, i) => {
            const done = i < 1, active = i === 1;
            return (
              <React.Fragment key={step}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span style={{
                    width: 32, height: 32, borderRadius: 999,
                    background: done ? '#0f766e' : active ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : 'rgba(15,23,42,0.06)',
                    color: done || active ? '#fffdf8' : '#5f6877',
                    display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
                    fontSize: 13, fontWeight: 800,
                  }}>{done ? <Icon name="check" size={14}/> : (i + 1)}</span>
                  <div>
                    <div style={{ fontSize: 11, fontWeight: 800, letterSpacing: '0.08em', textTransform: 'uppercase', color: '#5f6877' }}>Шаг {i + 1}</div>
                    <div style={{ fontSize: 14, fontWeight: 800 }}>{step}</div>
                  </div>
                </div>
                {i < 2 && <span style={{ flex: 1, height: 2, background: done ? '#0f766e' : 'rgba(15,23,42,0.10)', borderRadius: 999 }} />}
              </React.Fragment>
            );
          })}
        </div>

        {/* Step 2 form */}
        <div style={{ display: 'grid', gridTemplateColumns: '1.5fr 1fr', gap: 20, marginBottom: 30 }}>
          <div className="rh-card" style={{ padding: 24, borderRadius: 24 }}>
            <h2 className="rh-h3" style={{ margin: 0 }}>Фото объекта</h2>
            <p style={{ margin: '4px 0 16px', fontSize: 13, color: '#5f6877' }}>Перетащите 5 – 20 фото. Первое — на обложку. Минимум 1280 px по короткой стороне.</p>

            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12 }}>
              <div style={{ position: 'relative', gridColumn: 'span 2', gridRow: 'span 2', height: 280 }}>
                <PhotoPlaceholder height="100%" radius={16} variant="banya" label="" />
                <span style={{ position: 'absolute', top: 10, left: 10, padding: '5px 10px', borderRadius: 999, background: 'rgba(15,118,110,0.94)', color: '#fff', fontSize: 11, fontWeight: 700 }}>Обложка</span>
              </div>
              {[
                ['forest', false], ['city', false], ['sand', false], ['spa', false],
              ].map(([v], i) => (
                <PhotoPlaceholder key={i} height={132} radius={14} variant={v} label="" />
              ))}
              <button style={{
                height: 132, borderRadius: 14,
                background: 'rgba(255,255,255,0.7)',
                border: '2px dashed rgba(15,23,42,0.20)',
                display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', gap: 6,
                color: '#5f6877', fontSize: 12, fontWeight: 700, cursor: 'pointer',
              }}>
                <Icon name="upload" size={20} />
                Загрузить ещё
              </button>
              <div style={{ height: 132, borderRadius: 14, background: 'rgba(15,118,110,0.08)', border: '1px solid rgba(15,118,110,0.18)', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', color: '#0a5f59', fontSize: 11, fontWeight: 800, padding: 10, textAlign: 'center', lineHeight: 1.4 }}>
                <Icon name="bolt" size={18}/>
                AI улучшит<br/>освещение и крадж
              </div>
            </div>

            <h2 className="rh-h3" style={{ margin: '28px 0 4px' }}>Удобства</h2>
            <p style={{ margin: '0 0 14px', fontSize: 13, color: '#5f6877' }}>Гостям важны эти 18 пунктов. Чем больше — тем выше в выдаче.</p>
            <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 8 }}>
              {[
                ['Дровяная парная', true], ['Электрическая парная', false], ['Хаммам', false],
                ['Купель', true], ['Бассейн', true], ['Чан', true],
                ['Камин', true], ['Терраса', true], ['Беседка', false],
                ['Мангал', true], ['Кухня', true], ['Душевая', true],
                ['Wi-Fi', true], ['Парковка', true], ['Душевые принадлежности', false],
                ['Халаты и тапки', true], ['Можно с собакой', false], ['Можно с детьми', true],
              ].map(([t, on]) => (
                <label key={t} style={{
                  display: 'flex', alignItems: 'center', gap: 8,
                  padding: '8px 12px', borderRadius: 12,
                  background: on ? 'rgba(15,118,110,0.08)' : 'rgba(255,255,255,0.7)',
                  border: on ? '1px solid rgba(15,118,110,0.28)' : '1px solid rgba(15,23,42,0.08)',
                  fontSize: 13, fontWeight: 600, cursor: 'pointer',
                  color: on ? '#0a5f59' : '#16212b',
                }}>
                  <span style={{ width: 16, height: 16, borderRadius: 5, background: on ? '#0a5f59' : '#fff', border: on ? 0 : '1.5px solid rgba(15,23,42,0.30)', display: 'inline-flex', alignItems: 'center', justifyContent: 'center', color: '#fff' }}>{on && <Icon name="check" size={11}/>}</span>
                  {t}
                </label>
              ))}
            </div>

            <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 28, paddingTop: 18, borderTop: '1px solid rgba(15,23,42,0.06)' }}>
              <button className="rh-btn rh-btn--default"><Icon name="arrl" size={14}/> Назад</button>
              <div style={{ display: 'flex', gap: 8 }}>
                <button className="rh-btn rh-btn--ghost rh-btn--sm">Сохранить и выйти</button>
                <button className="rh-btn rh-btn--primary">Продолжить · Цены <Icon name="arrow" size={14}/></button>
              </div>
            </div>
          </div>

          {/* Live preview */}
          <div style={{ position: 'sticky', top: 20, height: 'fit-content' }}>
            <span className="rh-eyebrow-muted">Предпросмотр карточки</span>
            <div className="rh-card" style={{ padding: 0, borderRadius: 22, overflow: 'hidden', marginTop: 8 }}>
              <PhotoPlaceholder height={180} radius={0} variant="banya" label="" />
              <div style={{ padding: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <h3 style={{ margin: 0, fontSize: 16, fontWeight: 800 }}>Берёзовая роща</h3>
                  <span style={{ fontSize: 16, fontWeight: 800 }}>~3 200 ₽<span style={{ fontSize: 11, color: '#5f6877', fontWeight: 600 }}>/ч</span></span>
                </div>
                <div style={{ fontSize: 12, color: '#5f6877', marginTop: 2 }}>Москва, Рублёвское ш., 12</div>
                <div style={{ display: 'flex', gap: 5, marginTop: 10, flexWrap: 'wrap' }}>
                  {['Дровяная', 'Чан', 'Бассейн', 'Терраса', '+9'].map(t => <span key={t} className="rh-tag" style={{ fontSize: 10 }}>{t}</span>)}
                </div>
                <div style={{ marginTop: 10, fontSize: 11, color: '#5f6877' }}>
                  ▣ Обновится после публикации
                </div>
              </div>
            </div>

            <div style={{ marginTop: 14, padding: 14, borderRadius: 16, background: 'rgba(15,118,110,0.06)', border: '1px solid rgba(15,118,110,0.20)' }}>
              <div style={{ fontSize: 12, fontWeight: 800, color: '#0a5f59', letterSpacing: '0.06em', textTransform: 'uppercase' }}>Прогноз качества</div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginTop: 8 }}>
                <div style={{ flex: 1, height: 8, borderRadius: 999, background: 'rgba(15,118,110,0.16)', overflow: 'hidden' }}>
                  <div style={{ width: '72%', height: '100%', background: 'linear-gradient(90deg,#0f766e,#15803d)' }}/>
                </div>
                <span style={{ fontSize: 16, fontWeight: 800, color: '#0a5f59' }}>72%</span>
              </div>
              <div style={{ fontSize: 12, color: '#0a5f59', marginTop: 8, lineHeight: 1.5 }}>
                + 14 % если добавить ещё 4 фото<br/>
                + 8 % если описать процесс топки<br/>
                + 5 % за бесплатную отмену 24 ч
              </div>
            </div>
          </div>
        </div>
      </Section>
      <Footer />
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageOwnerInbox, PageOwnerFinance, PageOwnerOnboarding });
