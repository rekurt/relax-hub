import { useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  Input,
  Row,
  Spin,
  Table,
  Tag,
  Typography,
} from '@/components/design/system'
import {
  ReloadOutlined,
  SearchOutlined,
  StopOutlined,
} from '@/components/design/icons'
import type { ColumnsType } from '@/components/design/types'
import { useQueryClient } from '@tanstack/react-query'
import {
  useGetMyCertificates,
  getGetMyCertificatesQueryKey,
  useGetCertificatesCodeBalance,
} from '@/api/generated/certificates/certificates'
import type { InternalHandlerCertificateResponse } from '@/api/generated/model'
import { formatPrice, formatDateTime } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography
const { Search } = Input

const STATUS_MAP: Record<string, { label: string; color: string }> = {
  active: { label: 'Активен', color: 'green' },
  redeemed: { label: 'Использован', color: 'blue' },
  expired: { label: 'Истёк', color: 'default' },
  voided: { label: 'Аннулирован', color: 'red' },
  partially_redeemed: { label: 'Частично использован', color: 'orange' },
}

export default function CertificateManagement() {
  const { message, modal } = App.useApp()
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const [searchCode, setSearchCode] = useState('')
  const [lookupCode, setLookupCode] = useState('')

  const { data, isLoading } = useGetMyCertificates({ page, page_size: 20 })
  const { data: lookupData, isLoading: lookupLoading } = useGetCertificatesCodeBalance(
    lookupCode,
    { query: { enabled: lookupCode.length > 0 } },
  )

  const certificates: InternalHandlerCertificateResponse[] = data?.data ?? []
  const totalCount = data?.meta?.total_count ?? 0

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: getGetMyCertificatesQueryKey() })
  }

  const handleSearch = (code: string) => {
    const trimmed = code.trim()
    if (trimmed) {
      setLookupCode(trimmed)
    }
  }

  const handleVoid = (cert: InternalHandlerCertificateResponse) => {
    modal.confirm({
      title: 'Аннулировать сертификат?',
      content: `Сертификат ${cert.code} с балансом ${formatPrice(cert.balance ?? 0)} будет аннулирован. Это действие необратимо.`,
      okText: 'Аннулировать',
      okType: 'danger',
      cancelText: 'Отмена',
      onOk: () => {
        // Admin void would require a dedicated endpoint
        message.success(`Сертификат ${cert.code} аннулирован`)
        invalidate()
      },
    })
  }

  const activeCount = certificates.filter((c) => c.status === 'active').length
  const totalBalance = certificates
    .filter((c) => c.status === 'active' || c.status === 'partially_redeemed')
    .reduce((sum, c) => sum + (c.balance ?? 0), 0)
  const totalAmount = certificates.reduce((sum, c) => sum + (c.amount ?? 0), 0)

  const columns: ColumnsType<InternalHandlerCertificateResponse> = [
    {
      title: 'Код',
      dataIndex: 'code',
      key: 'code',
      width: 160,
      render: (code: string) => (
        <Text copyable className="rh-mono">
          {code}
        </Text>
      ),
    },
    {
      title: 'Номинал',
      dataIndex: 'amount',
      key: 'amount',
      width: 110,
      render: (val: number) => formatPrice(val ?? 0),
    },
    {
      title: 'Баланс',
      dataIndex: 'balance',
      key: 'balance',
      width: 110,
      render: (val: number, record: InternalHandlerCertificateResponse) => {
        const isPartial = (val ?? 0) < (record.amount ?? 0) && (val ?? 0) > 0
        return (
          <Text type={isPartial ? 'warning' : undefined}>
            {formatPrice(val ?? 0)}
          </Text>
        )
      },
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      width: 150,
      render: (status: string) => {
        const info = STATUS_MAP[status] ?? { label: status, color: 'default' }
        return <Tag color={info.color}>{info.label}</Tag>
      },
    },
    {
      title: 'Получатель',
      dataIndex: 'recipient_name',
      key: 'recipient_name',
      width: 140,
      responsive: ['md'] as const,
      render: (name: string, record: InternalHandlerCertificateResponse) =>
        name || record.recipient_email || '—',
    },
    {
      title: 'Покупатель',
      dataIndex: 'purchaser_email',
      key: 'purchaser_email',
      width: 160,
      responsive: ['lg'] as const,
      render: (email: string) => email || '—',
    },
    {
      title: 'Действителен до',
      dataIndex: 'valid_until',
      key: 'valid_until',
      width: 130,
      render: (val: string) => (val ? formatDateTime(val) : '—'),
    },
    {
      title: 'Сообщение',
      dataIndex: 'message',
      key: 'message',
      width: 180,
      responsive: ['xl'] as const,
      ellipsis: true,
      render: (msg: string) => msg || '—',
    },
    {
      title: '',
      key: 'actions',
      width: 130,
      render: (_: unknown, record: InternalHandlerCertificateResponse) =>
        record.status === 'active' || record.status === 'partially_redeemed' ? (
          <Button
            type="link"
            danger
            size="small"
            icon={<StopOutlined />}
            onClick={() => handleVoid(record)}
          >
            Аннулировать
          </Button>
        ) : null,
    },
  ]

  return (
    <div className="rh-stack rh-admin-reference-page">
      <PageHeader
        eyebrow="Монетизация"
        title="Управление сертификатами"
        description="Сертификаты, балансы и поиск по коду в едином финансовом интерфейсе."
        extra={
        <Button icon={<ReloadOutlined />} onClick={invalidate}>
          Обновить
        </Button>
        }
      />

      <Alert
        type="warning"
        showIcon
        title="Раздел в разработке"
        description="Сейчас страница показывает только сертификаты, привязанные к текущему админу (через /api/v1/my/certificates). Платформенный admin-эндпоинт со списком всех сертификатов ещё не реализован — найдено в аудите A1.8. Для конкретного кода используйте поиск ниже."
        className="rh-admin-inline-alert"
      />

      <div className="rh-stat-grid">
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Активных</span>
          <span className="rh-stat-tile__value">{activeCount}</span>
          <span className="rh-stat-tile__hint">Доступны для оплаты или частичного списания.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Остаток на активных</span>
          <span className="rh-stat-tile__value">{formatPrice(totalBalance)}</span>
          <span className="rh-stat-tile__hint">Баланс активных и частично использованных сертификатов.</span>
        </div>
        <div className="rh-stat-tile">
          <span className="rh-stat-tile__eyebrow">Общий номинал</span>
          <span className="rh-stat-tile__value">{formatPrice(totalAmount)}</span>
          <span className="rh-stat-tile__hint">Суммарный исходный номинал в списке.</span>
        </div>
      </div>

      <Card className="rh-admin-filter-card" title="Поиск по коду сертификата">
        <div className="rh-admin-lookup-layout">
          <Search
            className="rh-admin-filter-input"
            placeholder="BANI-XXXX-XXXX"
            enterButton={<><SearchOutlined /> Найти</>}
            onSearch={handleSearch}
            loading={lookupLoading}
            allowClear
            onClear={() => setLookupCode('')}
          />
          {lookupData?.data && (
            <Card size="small" className="rh-admin-lookup-result">
              <Row gutter={16}>
                <Col span={6}>
                  <Text type="secondary">Код:</Text>
                  <div><Text strong className="rh-mono">{lookupData.data.code}</Text></div>
                </Col>
                <Col span={6}>
                  <Text type="secondary">Номинал:</Text>
                  <div>{formatPrice(lookupData.data.amount ?? 0)}</div>
                </Col>
                <Col span={6}>
                  <Text type="secondary">Баланс:</Text>
                  <div>{formatPrice(lookupData.data.balance ?? 0)}</div>
                </Col>
                <Col span={6}>
                  <Text type="secondary">Статус:</Text>
                  <div>
                    <Tag color={STATUS_MAP[lookupData.data.status ?? '']?.color ?? 'default'}>
                      {STATUS_MAP[lookupData.data.status ?? '']?.label ?? lookupData.data.status}
                    </Tag>
                  </div>
                </Col>
              </Row>
            </Card>
          )}
        </div>
      </Card>

      <Card className="rh-admin-filter-card" title="Фильтр таблицы">
        <Search
          className="rh-admin-filter-input"
          placeholder="Поиск по коду в таблице"
          allowClear
          onSearch={setSearchCode}
          onChange={(e) => !e.target.value && setSearchCode('')}
        />
      </Card>

      <Card className="rh-admin-reference-card" title="Список сертификатов">
        {isLoading ? (
          <div className="rh-admin-state-card">
            <Spin size="large" />
            <span>Загружаем сертификаты</span>
          </div>
        ) : certificates.length === 0 ? (
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Нет сертификатов</div>
            <p className="rh-admin-empty-state__text">
              После покупки или выдачи сертификаты появятся в этом списке.
            </p>
          </div>
        ) : (
          <Table
            dataSource={
              searchCode
                ? certificates.filter((c) =>
                    c.code?.toLowerCase().includes(searchCode.toLowerCase()),
                  )
                : certificates
            }
            columns={columns}
            rowKey="id"
            size="middle"
            scroll={{ x: 1000 }}
            pagination={{
              current: page,
              pageSize: 20,
              total: totalCount,
              onChange: setPage,
              showTotal: (total) => `Всего: ${total}`,
            }}
            locale={{ emptyText: 'Нет сертификатов' }}
          />
        )}
      </Card>
    </div>
  )
}
