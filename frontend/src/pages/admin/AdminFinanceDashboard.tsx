import { useState } from 'react'
import { Card, Col, Row, Spin, Statistic, Table, Tag, Typography, Button, App, Space } from 'antd'
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

const { Title } = Typography

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
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <Title level={3} style={{ margin: 0 }}>Финансовый дашборд</Title>
        <Space>
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
      </div>

      <Spin spinning={summaryLoading}>
        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Кошельки клиентов"
                value={summary?.client_wallets_total ? summary.client_wallets_total / 100 : 0}
                precision={0}
                suffix="₽"
                prefix={<WalletOutlined />}
              />
              <div style={{ marginTop: 8, fontSize: 12, color: '#888' }}>
                {summary?.client_wallets_count ?? 0} кошельков
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Кошельки владельцев"
                value={summary?.owner_wallets_total ? summary.owner_wallets_total / 100 : 0}
                precision={0}
                suffix="₽"
                prefix={<DollarOutlined />}
              />
              <div style={{ marginTop: 8, fontSize: 12, color: '#888' }}>
                {summary?.owner_wallets_count ?? 0} кошельков
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Эскроу"
                value={summary?.escrow_held_total ? summary.escrow_held_total / 100 : 0}
                precision={0}
                suffix="₽"
                prefix={<SafetyOutlined />}
              />
              <div style={{ marginTop: 8, fontSize: 12, color: '#888' }}>
                {summary?.escrow_count ?? 0} записей
              </div>
            </Card>
          </Col>
          <Col xs={24} sm={12} lg={6}>
            <Card>
              <Statistic
                title="Холды"
                value={summary?.wallet_holds_total ? summary.wallet_holds_total / 100 : 0}
                precision={0}
                suffix="₽"
                prefix={<WarningOutlined />}
              />
              <div style={{ marginTop: 8, fontSize: 12, color: '#888' }}>
                Итого на платформе: {formatPrice(summary?.platform_total ?? 0)}
              </div>
            </Card>
          </Col>
        </Row>
      </Spin>

      {summary?.last_report && (
        <Card style={{ marginBottom: 24 }}>
          <Title level={5}>Последняя сверка</Title>
          <Row gutter={16}>
            <Col span={6}>Статус: {reconciliationStatusTag(summary.last_report.status)}</Col>
            <Col span={6}>Платежи: {summary.last_report.internal_payments_count ?? 0} внутр. / {summary.last_report.provider_payments_count ?? 0} провайдер</Col>
            <Col span={6}>Расхождение: {formatPrice(summary.last_report.payment_discrepancy ?? 0)}</Col>
            <Col span={6}>Дата: {summary.last_report.created_at ? formatDateTime(summary.last_report.created_at) : '—'}</Col>
          </Row>
        </Card>
      )}

      <Title level={4} style={{ marginBottom: 16 }}>Отчёты сверки</Title>
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
        style={{ marginBottom: 24 }}
      />

      <Title level={4} style={{ marginBottom: 16 }}>Снимки флоата</Title>
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
    </div>
  )
}
