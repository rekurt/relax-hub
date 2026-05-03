import { useMemo, useState } from 'react'
import {
  Card,
  DatePicker,
  Select,
  Table,
  Tag,
  Typography,
} from 'antd'
import { LockOutlined, WalletOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useGetMyWallet, useGetMyWalletTransactions } from '@/api/generated/wallet/wallet'
import type { InternalHandlerWalletTransactionResponse } from '@/api/generated/model'
import { formatDateTime, formatPrice } from '@/lib/format'
import EmptyState from '@/components/EmptyState'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography
const { RangePicker } = DatePicker

const TX_TYPE_OPTIONS = [
  { value: '', label: 'Все типы' },
  { value: 'income', label: 'Доход' },
  { value: 'payout', label: 'Выплата' },
  { value: 'service_fee', label: 'Комиссия' },
  { value: 'refund', label: 'Возврат' },
  { value: 'cashback', label: 'Кэшбэк' },
  { value: 'promo', label: 'Промо' },
]

const TX_TYPE_COLORS: Record<string, string> = {
  income: 'green',
  payout: 'blue',
  service_fee: 'orange',
  refund: 'red',
  cashback: 'gold',
  promo: 'purple',
  spend: 'red',
  topup: 'green',
}

const TX_TYPE_LABELS: Record<string, string> = {
  income: 'Доход',
  payout: 'Выплата',
  service_fee: 'Комиссия',
  refund: 'Возврат',
  cashback: 'Кэшбэк',
  promo: 'Промо',
  spend: 'Списание',
  topup: 'Пополнение',
}

const PERIOD_OPTIONS = [
  { value: 'week', label: 'Неделя' },
  { value: 'month', label: 'Месяц' },
  { value: 'year', label: 'Год' },
]

export default function FinanceDashboard() {
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [typeFilter, setTypeFilter] = useState('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)
  const [period, setPeriod] = useState('month')

  const { data: walletData, isLoading: walletLoading } = useGetMyWallet()
  const wallet = walletData?.data

  const { data: txData, isLoading: txLoading } = useGetMyWalletTransactions({
    page,
    page_size: pageSize,
    ...(typeFilter ? { type: typeFilter } : {}),
  })

  const transactions = txData?.data ?? []
  const txMeta = txData?.meta

  const filteredTransactions = transactions.filter((tx) => {
    if (dateRange?.[0] && dateRange?.[1] && tx.created_at) {
      const txDate = dayjs(tx.created_at)
      if (txDate.isBefore(dateRange[0], 'day') || txDate.isAfter(dateRange[1], 'day')) {
        return false
      }
    }
    return true
  })

  const incomeStats = useMemo(() => {
    return filteredTransactions.reduce(
      (acc, tx) => {
        if (tx.type === 'income' || tx.type === 'topup' || tx.type === 'cashback' || tx.type === 'promo') {
          acc.positive += tx.amount ?? 0
        } else {
          acc.negative += tx.amount ?? 0
        }
        return acc
      },
      { positive: 0, negative: 0 },
    )
  }, [filteredTransactions])

  const txColumns: ColumnsType<InternalHandlerWalletTransactionResponse> = [
    {
      title: 'Дата',
      key: 'created_at',
      render: (_, record) => record.created_at ? formatDateTime(record.created_at, 'DD.MM.YYYY HH:mm') : '\u2014',
      sorter: (a, b) => (a.created_at ?? '').localeCompare(b.created_at ?? ''),
    },
    {
      title: 'Тип',
      dataIndex: 'type',
      key: 'type',
      render: (type: string) => (
        <Tag color={TX_TYPE_COLORS[type] ?? 'default'}>
          {TX_TYPE_LABELS[type] ?? type}
        </Tag>
      ),
    },
    {
      title: 'Сумма',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number, record) => {
        const isPositive = record.type !== 'spend' && record.type !== 'service_fee' && record.type !== 'payout'
        return (
          <span style={{ color: isPositive ? '#15803d' : '#b42318', fontWeight: 500 }}>
            {isPositive ? '+' : '-'}{formatPrice(amount ?? 0)}
          </span>
        )
      },
    },
    {
      title: 'Описание',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'Баланс после',
      dataIndex: 'balance_after',
      key: 'balance_after',
      render: (value: number) => value != null ? formatPrice(value) : '\u2014',
      responsive: ['md'] as const,
    },
  ]

  return (
    <div className="bani-stack">
      <PageHeader
        eyebrow="Кошелёк"
        title="Финансы"
        description="Деньги, заморозки и история операций собраны в одном месте. Верхний блок отвечает на вопрос «сколько доступно сейчас», нижний показывает, откуда взялись изменения."
      />

      <section className="bani-hero-panel">
        <div className="bani-hero-panel__eyebrow">Кошелёк</div>
        <h2 className="bani-hero-panel__title">Операционная картина без лишних переходов</h2>
        <div className="bani-hero-panel__description">
          Владелец сначала видит доступный остаток и заморозки, а затем сразу проваливается в историю начислений, комиссий и возвратов.
        </div>
        <div className="bani-stat-grid">
          <div className="bani-stat-tile">
            <span className="bani-stat-tile__eyebrow">Баланс</span>
            <div className="bani-stat-tile__value">{walletLoading ? '...' : formatPrice(wallet?.balance ?? 0)}</div>
            <span className="bani-stat-tile__hint"><WalletOutlined /> Общий остаток в кошельке</span>
          </div>
          <div className="bani-stat-tile">
            <span className="bani-stat-tile__eyebrow">Доступно</span>
            <div className="bani-stat-tile__value">{walletLoading ? '...' : formatPrice(wallet?.available ?? 0)}</div>
            <span className="bani-stat-tile__hint">Средства доступны для вывода и следующих операций</span>
          </div>
          <div className="bani-stat-tile">
            <span className="bani-stat-tile__eyebrow">Заморожено</span>
            <div className="bani-stat-tile__value">{walletLoading ? '...' : formatPrice(wallet?.held_amount ?? 0)}</div>
            <span className="bani-stat-tile__hint"><LockOutlined /> Резерв под незавершённые брони и удержания</span>
          </div>
        </div>
      </section>

      <div className="bani-grid bani-grid--content-aside">
        <Card>
          <div className="bani-section-card">
            <div className="bani-toolbar">
              <div>
                <h2 className="bani-section-card__title">Последние операции</h2>
                <div className="bani-section-card__description">
                  Быстрые фильтры по типу и периоду помогают понять, где пришли деньги, а где ушли на комиссию, выплату или возврат.
                </div>
              </div>
              <div className="bani-filter-toolbar__controls">
                <Select
                  value={typeFilter}
                  onChange={(value) => {
                    setTypeFilter(value)
                    setPage(1)
                  }}
                  options={TX_TYPE_OPTIONS}
                  placeholder="Тип операции"
                />
                <RangePicker
                  value={dateRange}
                  onChange={(dates) => setDateRange(dates)}
                  format="DD.MM.YYYY"
                  placeholder={['С', 'По']}
                />
              </div>
            </div>

            <Table
              columns={txColumns}
              dataSource={filteredTransactions}
              rowKey="id"
              loading={txLoading}
              locale={{ emptyText: <EmptyState description="Нет операций" /> }}
              pagination={{
                current: page,
                pageSize,
                total: txMeta?.total_count ?? 0,
                showSizeChanger: true,
                showTotal: (total) => `Всего: ${total}`,
                onChange: (nextPage, nextPageSize) => {
                  setPage(nextPage)
                  setPageSize(nextPageSize)
                },
              }}
            />
          </div>
        </Card>

        <div className="bani-stack">
          <Card>
            <div className="bani-section-card">
              <div className="bani-toolbar">
                <div>
                  <h2 className="bani-section-card__title">Доходы за период</h2>
                  <div className="bani-section-card__description">Верхнеуровневый обзор для быстрого ответа на вопрос, как идёт месяц.</div>
                </div>
                <Select
                  value={period}
                  onChange={setPeriod}
                  options={PERIOD_OPTIONS}
                  style={{ minWidth: 140 }}
                />
              </div>

              <div className="bani-info-grid">
                <div className="bani-info-card">
                  <span className="bani-info-card__label">Поступления</span>
                  <div className="bani-info-card__value" style={{ color: '#15803d' }}>{formatPrice(incomeStats.positive)}</div>
                  <div className="bani-info-card__hint">Включая брони, пополнения, бонусы и промо</div>
                </div>
                <div className="bani-info-card">
                  <span className="bani-info-card__label">Списания</span>
                  <div className="bani-info-card__value" style={{ color: '#b42318' }}>{formatPrice(incomeStats.negative)}</div>
                  <div className="bani-info-card__hint">Комиссии, выплаты и возвраты за выбранный период</div>
                </div>
              </div>

              <div className="bani-section-card__surface">
                <Text type="secondary">
                  График доходов ({PERIOD_OPTIONS.find((option) => option.value === period)?.label}) будет полезен, но даже без него блок уже отвечает на основные вопросы по cashflow.
                </Text>
              </div>
            </div>
          </Card>

          <Card>
            <div className="bani-kv">
              <div className="bani-kv__row">
                <span className="bani-kv__label">Всего операций на экране</span>
                <span className="bani-kv__value">{filteredTransactions.length}</span>
              </div>
              <div className="bani-kv__row">
                <span className="bani-kv__label">Фильтр по типу</span>
                <span className="bani-kv__value">{typeFilter ? TX_TYPE_OPTIONS.find((option) => option.value === typeFilter)?.label : 'Без фильтра'}</span>
              </div>
              <div className="bani-kv__row">
                <span className="bani-kv__label">Период обзора</span>
                <span className="bani-kv__value">{PERIOD_OPTIONS.find((option) => option.value === period)?.label}</span>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  )
}
