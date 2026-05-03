import { useState } from 'react'
import { Spin, Table, Tag, Button, App, Space } from 'antd'
import type { ColumnsType } from 'antd/es/table'
import {
  WalletOutlined,
  SafetyOutlined,
  DollarOutlined,
  SyncOutlined,
  WarningOutlined,
  CameraOutlined,
} from '@ant-design/icons'
import {
  useGetApiV1AdminReconciliationSummary,
  useGetApiV1AdminReconciliationReports,
  useGetApiV1AdminReconciliationSnapshots,
  usePostApiV1AdminReconciliationSnapshot,
  usePostApiV1AdminReconciliationReconcile,
} from '@/api/generated/admin/admin'
import type {
  InternalHandlerReconciliationReportResponse,
  InternalHandlerFloatSnapshotResponse,
} from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

function reconciliationStatusTag(status?: string) {
  switch (status) {
    case 'ok': return <Tag color="green">OK</Tag>
    case 'mismatch': return <Tag color="red">Расхождение</Tag>
    case 'error': return <Tag color="orange">Ошибка</Tag>
    default: return <Tag>{status ?? '—'}</Tag>
  }
}

export default function AdminFinanceDashboard() {
  const { message } = App.useApp()
  const [reportsPage, setReportsPage] = useState(1)
  const [snapshotsPage, setSnapshotsPage] = useState(1)

  const { data: summaryData, isLoading: summaryLoading } = useGetApiV1AdminReconciliationSummary()
  const { data: reportsData, isLoading: reportsLoading } = useGetApiV1AdminReconciliationReports({
    page: reportsPage,
    page_size: 10,
  })
  const { data: snapshotsData, isLoading: snapshotsLoading } = useGetApiV1AdminReconciliationSnapshots({
    page: snapshotsPage,
    page_size: 10,
  })

  const snapshotMutation = usePostApiV1AdminReconciliationSnapshot()
  const reconcileMutation = usePostApiV1AdminReconciliationReconcile()

  const summary = summaryData?.data
  const reports = reportsData?.data ?? []
  const reportsMeta = reportsData?.meta
  const snapshots = snapshotsData?.data ?? []
  const snapshotsMeta = snapshotsData?.meta

  const summaryMetricTiles = [
    {
      label: 'Кошельки клиентов',
      value: formatPrice(summary?.client_wallets_total ?? 0),
      hint: `${summary?.client_wallets_count ?? 0} кошельков`,
      icon: <WalletOutlined />,
    },
    {
      label: 'Кошельки владельцев',
      value: formatPrice(summary?.owner_wallets_total ?? 0),
      hint: `${summary?.owner_wallets_count ?? 0} кошельков`,
      icon: <DollarOutlined />,
    },
    {
      label: 'Эскроу',
      value: formatPrice(summary?.escrow_held_total ?? 0),
      hint: `${summary?.escrow_count ?? 0} записей`,
      icon: <SafetyOutlined />,
    },
    {
      label: 'Холды',
      value: formatPrice(summary?.wallet_holds_total ?? 0),
      hint: `Итого на платформе: ${formatPrice(summary?.platform_total ?? 0)}`,
      icon: <WarningOutlined />,
    },
  ]

  const handleSnapshot = () => {
    snapshotMutation.mutate(undefined, {
      onSuccess: () => message.success('Снимок создан'),
      onError: () => message.error('Ошибка создания снимка'),
    })
  }

  const handleReconcile = () => {
    reconcileMutation.mutate({ params: {} }, {
      onSuccess: () => message.success('Сверка запущена'),
      onError: () => message.error('Ошибка запуска сверки'),
    })
  }

  const reportColumns: ColumnsType<InternalHandlerReconciliationReportResponse> = [
    {
      title: 'Дата',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (v: string) => v ? formatDateTime(v) : '—',
    },
    {
      title: 'Период',
      key: 'period',
      render: (_: unknown, r) => r.period_start && r.period_end
        ? `${formatDateTime(r.period_start, 'DD.MM.YYYY')} — ${formatDateTime(r.period_end, 'DD.MM.YYYY')}`
        : '—',
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: reconciliationStatusTag,
    },
    {
      title: 'Платежи (внутр. / провайдер)',
      key: 'payments',
      render: (_: unknown, r) => `${r.internal_payments_count ?? 0} / ${r.provider_payments_count ?? 0}`,
    },
    {
      title: 'Расхождение платежей',
      dataIndex: 'payment_discrepancy',
      key: 'payment_discrepancy',
      render: (v: number) => v ? formatPrice(v) : '0 ₽',
    },
    {
      title: 'Расхождение возвратов',
      dataIndex: 'refund_discrepancy',
      key: 'refund_discrepancy',
      render: (v: number) => v ? formatPrice(v) : '0 ₽',
    },
  ]

  const snapshotColumns: ColumnsType<InternalHandlerFloatSnapshotResponse> = [
    {
      title: 'Дата',
      dataIndex: 'snapshot_date',
      key: 'snapshot_date',
      render: (v: string) => v ? formatDateTime(v, 'DD.MM.YYYY') : '—',
    },
    {
      title: 'Ожидаемый итого',
      dataIndex: 'expected_total',
      key: 'expected_total',
      render: (v: number) => formatPrice(v ?? 0),
    },
    {
      title: 'Фактический итого',
      dataIndex: 'actual_total',
      key: 'actual_total',
      render: (v: number) => formatPrice(v ?? 0),
    },
    {
      title: 'Расхождение',
      dataIndex: 'discrepancy',
      key: 'discrepancy',
      render: (v: number) => {
        const val = v ?? 0
        return <span style={{ color: val !== 0 ? '#ff4d4f' : undefined }}>{formatPrice(val)}</span>
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: reconciliationStatusTag,
    },
  ]

  return (
    <div className="bani-stack">
      <PageHeader
        size="compact"
        eyebrow="Финансы"
        title="Финансовый дашборд"
        description="Сверка выплат, возвратов и кошельков в плотном операционном интерфейсе."
        extra={(
          <Space wrap>
          <Button
            icon={<CameraOutlined />}
            onClick={handleSnapshot}
            loading={snapshotMutation.isPending}
          >
            Снять снимок
          </Button>
          <Button
            icon={<SyncOutlined />}
            onClick={handleReconcile}
            loading={reconcileMutation.isPending}
          >
            Сверка с провайдером
          </Button>
          </Space>
        )}
      />

      <Spin spinning={summaryLoading}>
        <div className="bani-admin-metric-grid">
          {summaryMetricTiles.map((tile) => (
            <div className="bani-admin-metric" key={tile.label}>
              <div className="bani-admin-metric__head">
                <span className="bani-admin-metric__label">{tile.label}</span>
                <span className="bani-admin-metric__icon">{tile.icon}</span>
              </div>
              <div className="bani-admin-metric__value">{tile.value}</div>
              <div className="bani-admin-metric__hint">{tile.hint}</div>
            </div>
          ))}
        </div>
      </Spin>

      {summary?.last_report && (
        <section className="bani-admin-panel">
          <div className="bani-admin-toolbar" style={{ marginBottom: 16 }}>
            <div className="bani-admin-toolbar__copy">
              <h2 className="bani-admin-toolbar__title">Последняя сверка</h2>
              <div className="bani-admin-toolbar__hint">Короткая сводка по последнему отчёту перед просмотром таблицы.</div>
            </div>
          </div>
          <div className="bani-info-grid">
            <div className="bani-info-card">
              <span className="bani-info-card__label">Статус</span>
              <div className="bani-info-card__value">{reconciliationStatusTag(summary.last_report.status)}</div>
            </div>
            <div className="bani-info-card">
              <span className="bani-info-card__label">Платежи</span>
              <div className="bani-info-card__value">{summary.last_report.internal_payments_count ?? 0} / {summary.last_report.provider_payments_count ?? 0}</div>
              <div className="bani-info-card__hint">внутр. / провайдер</div>
            </div>
            <div className="bani-info-card">
              <span className="bani-info-card__label">Расхождение</span>
              <div className="bani-info-card__value">{formatPrice(summary.last_report.payment_discrepancy ?? 0)}</div>
            </div>
            <div className="bani-info-card">
              <span className="bani-info-card__label">Дата</span>
              <div className="bani-info-card__value">{summary.last_report.created_at ? formatDateTime(summary.last_report.created_at) : '—'}</div>
            </div>
          </div>
        </section>
      )}

      <section className="bani-admin-table-card">
        <div className="bani-admin-toolbar">
          <div className="bani-admin-toolbar__copy">
            <h2 className="bani-admin-toolbar__title">Отчёты сверки</h2>
            <div className="bani-admin-toolbar__hint">Строки расхождений и статусы по операциям провайдера.</div>
          </div>
        </div>
        <Table
          columns={reportColumns}
          dataSource={reports}
          rowKey="id"
          loading={reportsLoading}
          pagination={{
            current: reportsPage,
            pageSize: 10,
            total: reportsMeta?.total_count,
            onChange: setReportsPage,
          }}
          locale={{ emptyText: 'Нет отчётов' }}
        />
      </section>

      <section className="bani-admin-table-card">
        <div className="bani-admin-toolbar">
          <div className="bani-admin-toolbar__copy">
            <h2 className="bani-admin-toolbar__title">Снимки флоата</h2>
            <div className="bani-admin-toolbar__hint">Ожидаемые и фактические остатки платформы по дням.</div>
          </div>
        </div>
        <Table
          columns={snapshotColumns}
          dataSource={snapshots}
          rowKey="id"
          loading={snapshotsLoading}
          pagination={{
            current: snapshotsPage,
            pageSize: 10,
            total: snapshotsMeta?.total_count,
            onChange: setSnapshotsPage,
          }}
          locale={{ emptyText: 'Нет снимков' }}
        />
      </section>
    </div>
  )
}
