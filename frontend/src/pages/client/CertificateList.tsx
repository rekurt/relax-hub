import { useState } from 'react'
import { Typography, Table, Tag, Card, Input, Button, Row, Col, Statistic, Space, Alert, Descriptions, Spin, Empty } from 'antd'
import { GiftOutlined, CheckCircleOutlined, ClockCircleOutlined, StopOutlined, SearchOutlined } from '@ant-design/icons'
import { App } from 'antd'
import { Link } from 'react-router-dom'
import { useGetMyCertificates, usePostCertificatesRedeem, useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'
import { formatPrice, formatDateTime } from '@/lib/format'
import type { ColumnsType } from 'antd/es/table'
import type { InternalHandlerCertificateResponse } from '@/api/generated/model'

const { Title, Text } = Typography

const STATUS_MAP: Record<string, { color: string; label: string; icon: React.ReactNode }> = {
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

  const certificates = certificatesData?.data ?? []
  const meta = certificatesData?.meta

  const activeCerts = certificates.filter((c) => c.status === 'active')
  const totalBalance = activeCerts.reduce((sum, c) => sum + (c.balance ?? 0), 0)

  const handleRedeem = () => {
    if (!redeemCode.trim()) return
    redeemMutation.mutate(
      { data: { code: redeemCode.trim() } },
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
    if (redeemCode.trim().length >= 4) {
      setCheckCode(redeemCode.trim())
    }
  }

  const columns: ColumnsType<InternalHandlerCertificateResponse> = [
    {
      title: 'Код',
      dataIndex: 'code',
      key: 'code',
      render: (code: string) => <Text copyable strong>{code}</Text>,
    },
    {
      title: 'Номинал',
      dataIndex: 'amount',
      key: 'amount',
      render: (amount: number) => formatPrice(amount ?? 0),
    },
    {
      title: 'Остаток',
      dataIndex: 'balance',
      key: 'balance',
      render: (balance: number) => <Text strong>{formatPrice(balance ?? 0)}</Text>,
    },
    {
      title: 'Статус',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const s = STATUS_MAP[status] ?? { color: 'default', label: status, icon: null }
        return <Tag color={s.color} icon={s.icon}>{s.label}</Tag>
      },
    },
    {
      title: 'Действителен до',
      dataIndex: 'valid_until',
      key: 'valid_until',
      render: (date: string) => date ? formatDateTime(date, 'DD.MM.YYYY') : '—',
    },
    {
      title: 'Создан',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => date ? formatDateTime(date, 'DD.MM.YYYY') : '—',
    },
  ]

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '50px' }}>
        <Spin size="large" />
      </div>
    )
  }

  return (
    <div>
      <Title level={2}>
        <GiftOutlined /> Подарочные сертификаты
      </Title>

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={8}>
          <Card>
            <Statistic
              title="Активных сертификатов"
              value={activeCerts.length}
              prefix={<GiftOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card>
            <Statistic
              title="Общий баланс"
              value={formatPrice(totalBalance)}
              styles={{ content: { color: totalBalance > 0 ? '#52c41a' : undefined } }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8}>
          <Card>
            <Statistic
              title="Всего сертификатов"
              value={certificates.length}
            />
          </Card>
        </Col>
      </Row>

      <Card title="Активировать сертификат" style={{ marginTop: 16 }}>
        <Space orientation="vertical" style={{ width: '100%' }}>
          <Text>Введите код сертификата для активации или проверки баланса</Text>
          <Space.Compact style={{ width: '100%', maxWidth: 500 }}>
            <Input
              placeholder="BANI-XXXX-XXXX"
              value={redeemCode}
              onChange={(e) => setRedeemCode(e.target.value)}
              prefix={<GiftOutlined />}
              onPressEnter={handleCheckBalance}
            />
            <Button
              icon={<SearchOutlined />}
              onClick={handleCheckBalance}
              disabled={redeemCode.trim().length < 4}
            >
              Проверить
            </Button>
            <Button
              type="primary"
              onClick={handleRedeem}
              loading={redeemMutation.isPending}
              disabled={redeemCode.trim().length < 4}
            >
              Активировать
            </Button>
          </Space.Compact>

          {checkCode && balanceLoading && <Spin size="small" />}
          {checkCode && balanceData?.data && (
            <Descriptions bordered size="small" column={1} style={{ maxWidth: 500 }}>
              <Descriptions.Item label="Код">{balanceData.data.code}</Descriptions.Item>
              <Descriptions.Item label="Номинал">{formatPrice(balanceData.data.amount ?? 0)}</Descriptions.Item>
              <Descriptions.Item label="Остаток">
                <Text strong style={{ color: (balanceData.data.balance ?? 0) > 0 ? '#52c41a' : '#ff4d4f' }}>
                  {formatPrice(balanceData.data.balance ?? 0)}
                </Text>
              </Descriptions.Item>
              <Descriptions.Item label="Статус">
                {(() => {
                  const s = STATUS_MAP[balanceData.data.status ?? ''] ?? { color: 'default', label: balanceData.data.status, icon: null }
                  return <Tag color={s.color} icon={s.icon}>{s.label}</Tag>
                })()}
              </Descriptions.Item>
              <Descriptions.Item label="Действителен до">
                {balanceData.data.valid_until ? formatDateTime(balanceData.data.valid_until, 'DD.MM.YYYY') : '—'}
              </Descriptions.Item>
            </Descriptions>
          )}
        </Space>
      </Card>

      <Card
        title="Мои сертификаты"
        style={{ marginTop: 16 }}
        extra={
          <Link to="/client/certificates/purchase">
            <Button type="primary" icon={<GiftOutlined />}>Купить сертификат</Button>
          </Link>
        }
      >
        {certificates.length === 0 ? (
          <Empty description="У вас пока нет сертификатов">
            <Link to="/client/certificates/purchase">
              <Button type="primary" icon={<GiftOutlined />}>Купить сертификат</Button>
            </Link>
          </Empty>
        ) : (
          <Table
            columns={columns}
            dataSource={certificates}
            rowKey="id"
            pagination={{
              current: page,
              pageSize: 10,
              total: meta?.total_count,
              onChange: setPage,
              showSizeChanger: false,
            }}
          />
        )}
      </Card>

      <Alert
        title="Как использовать сертификат"
        description="Введите код сертификата при создании бронирования в поле «Подарочный сертификат». Сумма сертификата будет вычтена из стоимости бронирования. Неиспользованный остаток сохраняется на сертификате."
        type="info"
        showIcon
        style={{ marginTop: 16 }}
      />
    </div>
  )
}
