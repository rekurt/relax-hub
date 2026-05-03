import type { ReactNode } from 'react'
import { useMemo, useState } from 'react'
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  GiftOutlined,
  MailOutlined,
  SearchOutlined,
  StopOutlined,
  WalletOutlined,
} from '@/components/design/icons'
import { App, Button, Input, Pagination, Spin, Tag } from '@/components/design/system'
import { Link } from 'react-router-dom'
import {
  useGetCertificatesCodeBalance,
  useGetMyCertificates,
  usePostCertificatesRedeem,
} from '@/api/generated/certificates/certificates'
import PageHeader from '@/components/PageHeader'
import { formatDateTime, formatPrice } from '@/lib/format'

const STATUS_MAP: Record<string, { color: string; label: string; icon: ReactNode }> = {
  active: { color: 'green', label: 'Активен', icon: <CheckCircleOutlined /> },
  used: { color: 'default', label: 'Использован', icon: <StopOutlined /> },
  expired: { color: 'red', label: 'Истёк', icon: <ClockCircleOutlined /> },
}

export default function CertificateList() {
  const { message } = App.useApp()
  const [page, setPage] = useState(1)
  const [redeemCode, setRedeemCode] = useState('')
  const [checkCode, setCheckCode] = useState('')

  const { data: certificatesData, isLoading } = useGetMyCertificates({ page, page_size: 10 })
  const redeemMutation = usePostCertificatesRedeem()
  const { data: balanceData, isLoading: balanceLoading } = useGetCertificatesCodeBalance(
    checkCode,
    { query: { enabled: checkCode.trim().length >= 4 } },
  )

  const certificates = useMemo(() => certificatesData?.data ?? [], [certificatesData?.data])
  const meta = certificatesData?.meta

  const stats = useMemo(() => {
    const activeCertificates = certificates.filter((certificate) => certificate.status === 'active')
    const activeBalance = activeCertificates.reduce((total, certificate) => total + (certificate.balance ?? 0), 0)

    return {
      total: certificates.length,
      activeCount: activeCertificates.length,
      activeBalance,
    }
  }, [certificates])

  const handleRedeem = () => {
    const code = redeemCode.trim()
    if (code.length < 4) return

    redeemMutation.mutate(
      { data: { code } },
      {
        onSuccess: () => {
          message.success('Сертификат успешно активирован')
          setRedeemCode('')
          setCheckCode('')
        },
        onError: () => {
          message.error('Не удалось активировать сертификат')
        },
      },
    )
  }

  const handleCheckBalance = () => {
    const code = redeemCode.trim()
    if (code.length < 4) return
    setCheckCode(code)
  }

  if (isLoading) {
    return (
      <div className="rh-certificates-list-loading">
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div className="rh-stack rh-certificates-list-page">
      <PageHeader
        eyebrow="Личный кабинет"
        title="Мои сертификаты"
        description="Здесь видны активные остатки, статусы кодов и быстрые действия для проверки или активации нового сертификата."
        extra={(
          <Link to="/certificates">
            <Button type="primary" size="large" icon={<GiftOutlined />}>Купить сертификат</Button>
          </Link>
        )}
      />

      <section className="rh-hero-panel rh-certificates-list-hero">
        <div className="rh-hero-panel__eyebrow">Overview</div>
        <h2 className="rh-hero-panel__title">Сертификаты собраны как нормальный рабочий кабинет, а не как разрозненный список кодов</h2>
        <div className="rh-hero-panel__description">
          Можно быстро увидеть живой баланс, проверить новый код, активировать подарок и вернуться к покупке без визуального разрыва со страницей checkout.
        </div>

        <div className="rh-stat-grid">
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Активный баланс</span>
            <div className="rh-stat-tile__value">{formatPrice(stats.activeBalance)}</div>
            <span className="rh-stat-tile__hint"><WalletOutlined /> Сумма, доступная для следующих бронирований</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Активных сертификатов</span>
            <div className="rh-stat-tile__value">{stats.activeCount}</div>
            <span className="rh-stat-tile__hint">Коды с ненулевым остатком и актуальным сроком действия</span>
          </div>
          <div className="rh-stat-tile">
            <span className="rh-stat-tile__eyebrow">Всего сертификатов</span>
            <div className="rh-stat-tile__value">{stats.total}</div>
            <span className="rh-stat-tile__hint">Полная история всех выпущенных и привязанных сертификатов</span>
          </div>
        </div>
      </section>

      <div className="rh-grid rh-grid--content-aside rh-certificates-list-grid">
        <section className="rh-section-card">
          <div className="rh-section-card__surface rh-certificates-activation">
            <div className="rh-toolbar">
              <div>
                <div className="rh-certificates-section-eyebrow">Activation</div>
                <h2 className="rh-section-card__title">Активация сертификата</h2>
                <div className="rh-section-card__description">
                  Один и тот же блок отвечает и за ручную активацию кода, и за быструю проверку остатка перед бронированием.
                </div>
              </div>
            </div>

            <div className="rh-certificates-activation__controls">
              <Input
                placeholder="BANI-XXXX-XXXX"
                value={redeemCode}
                onChange={(event) => setRedeemCode(event.target.value)}
                onPressEnter={handleCheckBalance}
                prefix={<GiftOutlined />}
              />
              <Button
                size="large"
                icon={<SearchOutlined />}
                disabled={redeemCode.trim().length < 4}
                onClick={handleCheckBalance}
              >
                Проверить баланс
              </Button>
              <Button
                type="primary"
                size="large"
                disabled={redeemCode.trim().length < 4}
                loading={redeemMutation.isPending}
                onClick={handleRedeem}
              >
                Активировать сертификат
              </Button>
            </div>

            {checkCode && (
              <div className="rh-certificates-balance-panel">
                {balanceLoading ? (
                  <Spin size="small" />
                ) : balanceData?.data ? (
                  <>
                    <div className="rh-certificates-balance-panel__head">
                      <div>
                        <div className="rh-certificates-balance-panel__eyebrow">Balance check</div>
                        <strong>{balanceData.data.code}</strong>
                      </div>
                      {(() => {
                        const status = STATUS_MAP[balanceData.data.status ?? ''] ?? {
                          color: 'default',
                          label: balanceData.data.status ?? 'Неизвестно',
                          icon: null,
                        }
                        return <Tag color={status.color} icon={status.icon}>{status.label}</Tag>
                      })()}
                    </div>

                    <div className="rh-info-grid">
                      <div className="rh-info-card">
                        <span className="rh-info-card__label">Номинал</span>
                        <div className="rh-info-card__value">{formatPrice(balanceData.data.amount ?? 0)}</div>
                      </div>
                      <div className="rh-info-card">
                        <span className="rh-info-card__label">Остаток</span>
                        <div className="rh-info-card__value">{formatPrice(balanceData.data.balance ?? 0)}</div>
                      </div>
                      <div className="rh-info-card">
                        <span className="rh-info-card__label">Действителен до</span>
                        <div className="rh-info-card__value">
                          {balanceData.data.valid_until ? formatDateTime(balanceData.data.valid_until, 'DD.MM.YYYY') : '—'}
                        </div>
                      </div>
                    </div>
                  </>
                ) : null}
              </div>
            )}
          </div>
        </section>

        <section className="rh-section-card">
          <div className="rh-section-card__surface">
            <div className="rh-certificates-section-eyebrow">Usage</div>
            <h2 className="rh-section-card__title">Как использовать сертификат</h2>
            <div className="rh-feature-list">
              <div className="rh-feature-item">
                <div className="rh-feature-item__icon"><GiftOutlined /></div>
                <div className="rh-feature-item__copy">
                  <div className="rh-feature-item__title">Укажите код в checkout бронирования</div>
                  <div className="rh-feature-item__description">Система автоматически применит доступный остаток к итоговой стоимости брони.</div>
                </div>
              </div>
              <div className="rh-feature-item">
                <div className="rh-feature-item__icon"><MailOutlined /></div>
                <div className="rh-feature-item__copy">
                  <div className="rh-feature-item__title">Проверяйте баланс до оплаты</div>
                  <div className="rh-feature-item__description">Если сертификат частично использован, здесь всегда видно, сколько ещё доступно.</div>
                </div>
              </div>
              <div className="rh-feature-item">
                <div className="rh-feature-item__icon"><WalletOutlined /></div>
                <div className="rh-feature-item__copy">
                  <div className="rh-feature-item__title">Остаток не теряется</div>
                  <div className="rh-feature-item__description">Неиспользованная сумма сохраняется на сертификате до окончания срока действия.</div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>

      <section className="rh-section-card">
        <div className="rh-toolbar">
          <div>
            <div className="rh-certificates-section-eyebrow">Collection</div>
            <h2 className="rh-section-card__title">Все сертификаты</h2>
            <div className="rh-section-card__description">
              Карточки показывают статус, остаток, сроки и контакты, чтобы не приходилось разбираться по одному коду за раз.
            </div>
          </div>
          <Link to="/certificates">
            <Button icon={<GiftOutlined />}>Открыть checkout</Button>
          </Link>
        </div>

        {certificates.length === 0 ? (
          <div className="rh-certificates-empty">
            <div className="rh-admin-empty-state">
              <div className="rh-admin-empty-state__title">У вас пока нет сертификатов</div>
              <p className="rh-admin-empty-state__text">
                Купите первый сертификат или активируйте подарочный код.
              </p>
              <Link to="/certificates">
                <Button type="primary" size="large" icon={<GiftOutlined />}>Перейти к покупке</Button>
              </Link>
            </div>
          </div>
        ) : (
          <>
            <div className="rh-certificates-collection">
              {certificates.map((certificate) => {
                const status = STATUS_MAP[certificate.status ?? ''] ?? {
                  color: 'default',
                  label: certificate.status ?? 'Неизвестно',
                  icon: null,
                }

                return (
                  <article key={certificate.id} className="rh-certificates-card">
                    <div className="rh-certificates-card__header">
                      <div>
                        <div className="rh-certificates-card__eyebrow">Certificate</div>
                        <h3 className="rh-certificates-card__code">{certificate.code}</h3>
                      </div>
                      <Tag color={status.color} icon={status.icon}>{status.label}</Tag>
                    </div>

                    <div className="rh-certificates-card__stats">
                      <div>
                        <span>Номинал</span>
                        <strong>{formatPrice(certificate.amount ?? 0)}</strong>
                      </div>
                      <div>
                        <span>Остаток</span>
                        <strong>{formatPrice(certificate.balance ?? 0)}</strong>
                      </div>
                      <div>
                        <span>Действителен до</span>
                        <strong>{certificate.valid_until ? formatDateTime(certificate.valid_until, 'DD.MM.YYYY') : '—'}</strong>
                      </div>
                    </div>

                    <div className="rh-certificates-card__meta">
                      <div className="rh-kv__row">
                        <span className="rh-kv__label">Создан</span>
                        <span className="rh-kv__value">{certificate.created_at ? formatDateTime(certificate.created_at, 'DD.MM.YYYY') : '—'}</span>
                      </div>
                      <div className="rh-kv__row">
                        <span className="rh-kv__label">Покупатель</span>
                        <span className="rh-kv__value">{certificate.purchaser_email ?? '—'}</span>
                      </div>
                      <div className="rh-kv__row">
                        <span className="rh-kv__label">Получатель</span>
                        <span className="rh-kv__value">{certificate.recipient_name || certificate.recipient_email || 'Себе'}</span>
                      </div>
                    </div>
                  </article>
                )
              })}
            </div>

            {(meta?.total_pages ?? 1) > 1 && (
              <div className="rh-certificates-pagination">
                <Pagination
                  current={page}
                  pageSize={meta?.page_size ?? 10}
                  total={meta?.total_count}
                  onChange={setPage}
                  showSizeChanger={false}
                />
              </div>
            )}
          </>
        )}
      </section>
    </div>
  )
}
