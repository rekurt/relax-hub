/* global React */
const { Icon, Avatar, PhotoPlaceholder, Stars } = window.RH;
const { TopNav, Footer, Section } = window.RH;
const { SectionHeader } = window.RH;

// ════════════════════════════════════════════════════════════════
// PUBLIC · 03 LISTING DETAIL
// ════════════════════════════════════════════════════════════════
function PageListing() {
  return (
    <div className="rh-page" style={{ minHeight: 1900 }}>
      <TopNav active="catalogue" />

      {/* Breadcrumb */}
      <Section padding="14px 28px 0">
        <div style={{ fontSize: 12, color: '#5f6877' }}>
          <a href="#" style={crumb}>Каталог</a> · <a href="#" style={crumb}>Москва</a> · <span style={{ fontWeight: 700, color: '#16212b' }}>Берёзовая роща</span>
        </div>
      </Section>

      {/* Title + actions */}
      <Section padding="14px 28px 18px">
        <div style={{ display: 'flex', alignItems: 'flex-end', justifyContent: 'space-between', gap: 22 }}>
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 8 }}>
              <span className="rh-tag rh-tag--green">✓ Проверенный объект</span>
              <span className="rh-tag rh-tag--red">⚡ Last minute -20 %</span>
              <span className="rh-tag rh-tag--gold">Премиум</span>
            </div>
            <h1 style={{ margin: 0, fontSize: 40, fontWeight: 800, letterSpacing: '-0.04em', lineHeight: 1.05 }}>Берёзовая роща</h1>
            <div style={{ display: 'flex', gap: 18, marginTop: 10, color: '#5f6877', fontSize: 13, fontWeight: 600 }}>
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Icon name="pin" size={14} /> Москва, Рублёвское ш., 12 · 8 км от центра</span>
              <Stars value={4.9} count={128} />
              <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Icon name="user" size={14} /> до 8 гостей</span>
            </div>
          </div>
          <div style={{ display: 'flex', gap: 10 }}>
            <button className="rh-btn rh-btn--default"><Icon name="heart" size={16} /> В избранное</button>
            <button className="rh-btn rh-btn--default"><Icon name="upload" size={16} /> Поделиться</button>
            <button className="rh-btn rh-btn--default"><Icon name="msg" size={16} /></button>
          </div>
        </div>
      </Section>

      {/* Gallery */}
      <Section padding="0 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '1.6fr 1fr 1fr', gridTemplateRows: '1fr 1fr', gap: 10, height: 460 }}>
          <PhotoPlaceholder height="100%" variant="banya" radius={28} label="ФОТО · ОСНОВНОЕ" />
          <PhotoPlaceholder height="100%" variant="forest" radius={20} label="" />
          <PhotoPlaceholder height="100%" variant="city" radius={20} label="" />
          <PhotoPlaceholder height="100%" variant="sand" radius={20} label="" />
          <div style={{ position: 'relative' }}>
            <PhotoPlaceholder height="100%" variant="spa" radius={20} label="" />
            <button className="rh-btn rh-btn--default" style={{ position: 'absolute', right: 14, bottom: 14, background: 'rgba(255,255,255,0.94)' }}>
              <Icon name="grid" size={14} /> Все 28 фото
            </button>
          </div>
        </div>
      </Section>

      {/* Main content + booking sidebar */}
      <Section padding="0 28px 28px">
        <div style={{ display: 'grid', gridTemplateColumns: '1.6fr 0.8fr', gap: 28, alignItems: 'flex-start' }}>
          <div>
            {/* About */}
            <div className="rh-card" style={{ padding: 28, borderRadius: 28 }}>
              <SectionHeader eyebrow="Об объекте" title="Тёплое дерево, два бассейна и тишина за городом" />
              <p style={{ fontSize: 15, lineHeight: 1.65, color: '#16212b', marginTop: 14 }}>
                Двухэтажная баня на собственной территории с парной на берёзовых дровах,
                холодным бассейном, чаном на тёплой воде и просторной зоной отдыха
                с камином. На территории — мангал, веранда и парковка на 4 машины.
              </p>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 12, marginTop: 22 }}>
                {[
                  ['user', '8 гостей'], ['cal', 'от 2 ч'], ['car', 'парковка'], ['shield', 'депозит 5 000 ₽'],
                ].map(([i, t]) => (
                  <div key={t} style={{ padding: '12px 14px', borderRadius: 16, background: 'rgba(255,255,255,0.6)', border: '1px solid rgba(15,23,42,0.06)', display: 'flex', alignItems: 'center', gap: 10 }}>
                    <Icon name={i} size={18} style={{ color: '#0a5f59' }} />
                    <span style={{ fontSize: 13, fontWeight: 600 }}>{t}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Amenities */}
            <div className="rh-card" style={{ padding: 28, borderRadius: 28, marginTop: 18 }}>
              <SectionHeader eyebrow="Удобства" title="Что входит в стоимость" />
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 12, marginTop: 18 }}>
                {[
                  ['Парная на берёзовых дровах', true], ['Холодный бассейн 4×3 м', true],
                  ['Чан с подогревом', true], ['Душевые с косметикой', true],
                  ['Зона отдыха с камином', true], ['Мангал и шампуры', true],
                  ['Веники и шапки', true], ['Чай, вода, полотенца', true],
                  ['Wi-Fi, музыка', true], ['Караоке (по запросу)', false],
                  ['Кейтеринг (по запросу)', false], ['Массаж (по запросу)', false],
                ].map(([t, inc]) => (
                  <div key={t} style={{ display: 'flex', alignItems: 'center', gap: 10, fontSize: 13, color: inc ? '#16212b' : '#5f6877' }}>
                    <Icon name={inc ? 'check' : 'plus'} size={16} style={{ color: inc ? '#15803d' : '#5f6877' }} />
                    <span>{t}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* Schedule strip */}
            <div className="rh-card" style={{ padding: 22, borderRadius: 28, marginTop: 18 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                <span style={{ fontSize: 18, fontWeight: 800, letterSpacing: '-0.02em' }}>Свободные слоты — суббота, 20 апреля</span>
                <button style={{ background: 'none', border: 0, color: '#0a5f59', fontWeight: 700, fontSize: 13, cursor: 'pointer' }}>Все даты <Icon name="arrow" size={12} /></button>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(8, 1fr)', gap: 8 }}>
                {['12:00','14:00','16:00','18:00','19:00','20:00','21:00','22:00'].map((t, i) => (
                  <button key={t} style={{
                    padding: '12px 0', borderRadius: 16,
                    border: i === 4 ? '0' : '1px solid rgba(15,23,42,0.10)',
                    background: i === 4 ? 'linear-gradient(135deg,#0f766e,#0a5f59)' : i === 1 || i === 5 ? 'rgba(244,239,231,0.70)' : 'rgba(255,255,255,0.78)',
                    color: i === 4 ? '#fff' : i === 1 || i === 5 ? 'rgba(22,33,43,0.42)' : '#16212b',
                    fontSize: 14, fontWeight: 700, cursor: i === 1 || i === 5 ? 'not-allowed' : 'pointer',
                    boxShadow: i === 4 ? '0 14px 28px rgba(15,118,110,0.22)' : 'none',
                  }}>{t}</button>
                ))}
              </div>
              <div style={{ display: 'flex', gap: 14, marginTop: 12, fontSize: 12, color: '#5f6877' }}>
                <Legend swatch="linear-gradient(135deg,#0f766e,#0a5f59)" label="Выбрано" />
                <Legend swatch="rgba(255,255,255,0.78)" border label="Свободно" />
                <Legend swatch="rgba(244,239,231,0.70)" border label="Занято" />
              </div>
            </div>

            {/* Reviews */}
            <div className="rh-card" style={{ padding: 28, borderRadius: 28, marginTop: 18 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <SectionHeader eyebrow="Отзывы" title="Гости отмечают тишину и температуру парной" />
                <button className="rh-btn rh-btn--default rh-btn--sm">Все 128 отзывов</button>
              </div>
              <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 14, marginTop: 18 }}>
                {[
                  { n: 'Анна П.', t: 'teal', d: '12 апреля', r: 5, txt: 'Прекрасный вечер. Парная держит температуру, чан — кайф, веники свежие. Хозяин подсказал, как топить.' },
                  { n: 'Михаил К.', t: 'amber', d: '7 апреля', r: 5, txt: 'Брал на 6 человек, всем хватило места. Бассейн чистый, мангал готов к моменту приезда.' },
                  { n: 'Дарья В.', t: 'plum', d: '2 апреля', r: 4, txt: 'Хорошее место, чуть тесновато для 8 человек. Зато тишина, никаких соседей.' },
                  { n: 'Игорь С.', t: 'blue', d: '28 марта', r: 5, txt: 'Привезли в баню кейтеринг, всё прошло без сучка. Отдельное спасибо за пухлые полотенца.' },
                ].map((rv, i) => (
                  <div key={i} style={{ padding: 18, borderRadius: 20, background: 'rgba(255,255,255,0.7)', border: '1px solid rgba(15,23,42,0.06)' }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 10 }}>
                      <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
                        <Avatar name={rv.n} size={36} tone={rv.t} />
                        <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.2 }}>
                          <span style={{ fontSize: 13, fontWeight: 700 }}>{rv.n}</span>
                          <span style={{ fontSize: 11, color: '#5f6877' }}>{rv.d} · подтверждено</span>
                        </div>
                      </div>
                      <Stars value={rv.r} />
                    </div>
                    <p style={{ margin: 0, fontSize: 13, lineHeight: 1.55, color: '#16212b' }}>{rv.txt}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          {/* Booking widget */}
          <div style={{ position: 'sticky', top: 90, display: 'flex', flexDirection: 'column', gap: 14 }}>
            <div className="rh-card" style={{ padding: 22, borderRadius: 28 }}>
              <div style={{ display: 'flex', alignItems: 'baseline', gap: 8, marginBottom: 4 }}>
                <span style={{ fontSize: 26, fontWeight: 800, letterSpacing: '-0.03em' }}>3 200 ₽</span>
                <span style={{ fontSize: 14, color: '#5f6877' }}>/ час · от 2 часов</span>
              </div>
              <span className="rh-tag rh-tag--red" style={{ marginBottom: 16 }}>⚡ Last minute -20 % сегодня после 18:00</span>

              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 8, marginTop: 14 }}>
                <BookingField label="Дата" value="Сб, 20 апр" icon="cal" />
                <BookingField label="Гостей" value="4" icon="user" />
                <BookingField label="Начало" value="19:00" icon="clock" />
                <BookingField label="Конец" value="23:00" icon="clock" />
              </div>

              <div style={{ marginTop: 14, padding: 14, borderRadius: 18, background: 'rgba(255,255,255,0.62)', border: '1px solid rgba(15,23,42,0.06)' }}>
                <PriceRow label="3 200 ₽ × 4 ч" v="12 800 ₽" />
                <PriceRow label="Last minute -20 %" v="−2 560 ₽" tone="green" />
                <PriceRow label="Чан с подогревом" v="+ 1 200 ₽" />
                <PriceRow label="Депозит (возврат)" v="5 000 ₽" muted />
                <div style={{ height: 1, background: 'rgba(15,23,42,0.08)', margin: '10px 0' }} />
                <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: 16, fontWeight: 800 }}>
                  <span>Итого</span><span>11 440 ₽</span>
                </div>
              </div>

              <button className="rh-btn rh-btn--primary rh-btn--lg" style={{ width: '100%', marginTop: 16 }}>
                Забронировать <Icon name="arrow" size={16} />
              </button>
              <p style={{ fontSize: 11, color: '#5f6877', textAlign: 'center', marginTop: 10, lineHeight: 1.45 }}>
                Без оплаты — подтвердите заявку SMS-кодом. Возврат до 24 часов до визита.
              </p>
            </div>

            <div className="rh-card" style={{ padding: 18, borderRadius: 22, display: 'flex', alignItems: 'center', gap: 12 }}>
              <Avatar name="Иван Соколов" size={48} tone="amber" />
              <div style={{ display: 'flex', flexDirection: 'column', lineHeight: 1.2, flex: 1 }}>
                <span style={{ fontSize: 14, fontWeight: 700 }}>Иван Соколов</span>
                <span style={{ fontSize: 11, color: '#5f6877' }}>Хозяин · отвечает за 12 минут</span>
              </div>
              <button className="rh-btn rh-btn--default rh-btn--sm"><Icon name="msg" size={14} /> Написать</button>
            </div>
          </div>
        </div>
      </Section>

      <Footer />
    </div>
  );
}

const crumb = { color: '#5f6877', textDecoration: 'none' };

function Legend({ swatch, label, border }) {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}>
      <span style={{ width: 14, height: 14, borderRadius: 4, background: swatch, border: border ? '1px solid rgba(15,23,42,0.10)' : '0' }}/>
      {label}
    </span>
  );
}

function BookingField({ label, value, icon }) {
  return (
    <div style={{ padding: 12, borderRadius: 16, border: '1px solid rgba(15,23,42,0.10)', background: 'linear-gradient(180deg, rgba(255,255,255,0.96), rgba(251,247,240,0.86))' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6, fontSize: 11, fontWeight: 700, letterSpacing: '0.10em', textTransform: 'uppercase', color: '#5f6877' }}>
        <Icon name={icon} size={12} /> {label}
      </div>
      <div style={{ marginTop: 4, fontSize: 14, fontWeight: 700 }}>{value}</div>
    </div>
  );
}

function PriceRow({ label, v, tone, muted }) {
  const colors = { green: '#15803d' };
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', padding: '5px 0', fontSize: 13, color: muted ? '#5f6877' : tone ? colors[tone] : '#16212b', fontWeight: tone ? 700 : 500 }}>
      <span>{label}</span><span>{v}</span>
    </div>
  );
}

window.RH = window.RH || {};
Object.assign(window.RH, { PageListing });
