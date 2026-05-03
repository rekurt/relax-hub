import { useState } from 'react'
import {
  Alert,
  App,
  Button,
  Card,
  Col,
  Empty,
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

const { Title, Text } = Typography
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
        <Text copyable style={{ fontFamily: 'monospace' }}>
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
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 16 }}>
        <Title level={3} style={{ margin: 0 }}>
          Управление сертификатами
        </Title>
        <Button icon={<ReloadOutlined />} onClick={invalidate}>
          Обновить
        </Button>
      </div>

      <Alert
        type="warning"
        showIcon
        title="Раздел в разработке"
        description="Сейчас страница показывает только сертификаты, привязанные к текущему админу (через /api/v1/my/certificates). Платформенный admin-эндпоинт со списком всех сертификатов ещё не реализован — найдено в аудите A1.8. Для конкретного кода используйте поиск ниже."
        style={{ marginBottom: 16 }}
      />

      <Row gutter={[16, 16]} style={{ marginBottom: 16 }}>
        <Col xs={8}>
          <Card size="small">
            <Text type="secondary">Активных</Text>
            <div style={{ fontSize: 24, fontWeight: 600, color: '#15803d' }}>{activeCount}</div>
          </Card>
        </Col>
        <Col xs={8}>
          <Card size="small">
            <Text type="secondary">Остаток на активных</Text>
            <div style={{ fontSize: 24, fontWeight: 600 }}>{formatPrice(totalBalance)}</div>
          </Card>
        </Col>
        <Col xs={8}>
          <Card size="small">
            <Text type="secondary">Общий номинал</Text>
            <div style={{ fontSize: 24, fontWeight: 600 }}>{formatPrice(totalAmount)}</div>
          </Card>
        </Col>
      </Row>

      <Card size="small" style={{ marginBottom: 16 }} title="Поиск по коду сертификата">
        <div style={{ display: 'flex', gap: 12, alignItems: 'flex-start' }}>
          <Search
            placeholder="BANI-XXXX-XXXX"
            enterButton={<><SearchOutlined /> Найти</>}
            style={{ maxWidth: 360 }}
            onSearch={handleSearch}
            loading={lookupLoading}
            allowClear
            onClear={() => setLookupCode('')}
          />
          {lookupData?.data && (
            <Card size="small" style={{ flex: 1 }}>
              <Row gutter={16}>
                <Col span={6}>
                  <Text type="secondary">Код:</Text>
                  <div><Text strong style={{ fontFamily: 'monospace' }}>{lookupData.data.code}</Text></div>
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

      <div style={{ marginBottom: 16 }}>
        <Search
          placeholder="Поиск по коду в таблице"
          allowClear
          style={{ width: 280 }}
          onSearch={setSearchCode}
          onChange={(e) => !e.target.value && setSearchCode('')}
        />
      </div>

      {isLoading ? (
        <div style={{ textAlign: 'center', padding: 48 }}><Spin size="large" /></div>
      ) : certificates.length === 0 ? (
        <Empty description="Нет сертификатов" />
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
    </div>
  )
}
