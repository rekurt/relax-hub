import { useCallback, useState } from 'react'
import {
  Button,
  Card,
  Col,
  DatePicker,
  Dropdown,
  Row,
  Select,
  Space,
  Typography,
  message,
} from 'antd'
import {
  DownloadOutlined,
  FileExcelOutlined,
  FilePdfOutlined,
} from '@ant-design/icons'
import dayjs from 'dayjs'
import { useGetMyBathhouses } from '@/api/generated/bathhouses/bathhouses'
import { AUTH_TOKEN_KEY } from '@/lib/constants'

const { Title, Text } = Typography
const { RangePicker } = DatePicker

export default function FinancialReports() {
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs | null, dayjs.Dayjs | null] | null>(null)
  const [selectedBathhouse, setSelectedBathhouse] = useState<string>('')
  const [loading, setLoading] = useState<string | null>(null)

  const { data: bathhousesData } = useGetMyBathhouses()
  const bathhouses = bathhousesData?.data ?? []

  const bathhouseOptions = bathhouses.map((b) => ({
    value: b.id ?? '',
    label: b.name ?? 'Без названия',
  }))

  const downloadFile = useCallback(async (url: string, filename: string, loadingKey: string) => {
    setLoading(loadingKey)
    try {
      const token = localStorage.getItem(AUTH_TOKEN_KEY)
      const response = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
      })

      if (!response.ok) throw new Error('Download failed')

      const blob = await response.blob()
      const blobUrl = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = blobUrl
      a.download = filename
      a.click()
      URL.revokeObjectURL(blobUrl)
      message.success('Файл скачан')
    } catch {
      message.error('Ошибка при скачивании')
    } finally {
      setLoading(null)
    }
  }, [])

  const handleWalletExport = useCallback((format: 'csv' | 'pdf') => {
    const params = new URLSearchParams({ format })
    if (dateRange?.[0]) params.set('date_from', dateRange[0].format('YYYY-MM-DD'))
    if (dateRange?.[1]) params.set('date_to', dateRange[1].format('YYYY-MM-DD'))
    downloadFile(
      `/api/v1/my/wallet/export?${params}`,
      `wallet_history.${format}`,
      `wallet_${format}`,
    )
  }, [dateRange, downloadFile])

  const handleActDownload = useCallback(() => {
    if (!selectedBathhouse) {
      message.warning('Выберите баню')
      return
    }
    const params = new URLSearchParams()
    if (dateRange?.[0]) params.set('date_from', dateRange[0].format('YYYY-MM-DD'))
    if (dateRange?.[1]) params.set('date_to', dateRange[1].format('YYYY-MM-DD'))
    downloadFile(
      `/api/v1/my/finance/acts/${selectedBathhouse}?${params}`,
      'act.pdf',
      'act',
    )
  }, [selectedBathhouse, dateRange, downloadFile])

  const handleXmlExport = useCallback(() => {
    const params = new URLSearchParams()
    if (dateRange?.[0]) params.set('date_from', dateRange[0].format('YYYY-MM-DD'))
    if (dateRange?.[1]) params.set('date_to', dateRange[1].format('YYYY-MM-DD'))
    downloadFile(
      `/api/v1/my/finance/export-xml?${params}`,
      'finance_export.xml',
      'xml',
    )
  }, [dateRange, downloadFile])

  return (
    <div>
      <Title level={3} style={{ marginBottom: 16 }}>Отчёты</Title>

      <Space wrap style={{ marginBottom: 16 }}>
        <RangePicker
          value={dateRange}
          onChange={(dates) => setDateRange(dates)}
          format="DD.MM.YYYY"
          placeholder={['С', 'По']}
        />
      </Space>

      <Row gutter={[16, 16]}>
        <Col xs={24} md={8}>
          <Card title="История кошелька" styles={{ body: { minHeight: 120 } }}>
            <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
              Выгрузка всех операций по кошельку за период
            </Text>
            <Dropdown
              menu={{
                items: [
                  { key: 'csv', label: 'CSV', icon: <FileExcelOutlined />, onClick: () => handleWalletExport('csv') },
                  { key: 'pdf', label: 'PDF', icon: <FilePdfOutlined />, onClick: () => handleWalletExport('pdf') },
                ],
              }}
            >
              <Button
                icon={<DownloadOutlined />}
                loading={loading === 'wallet_csv' || loading === 'wallet_pdf'}
              >
                Скачать
              </Button>
            </Dropdown>
          </Card>
        </Col>

        <Col xs={24} md={8}>
          <Card title="Акт оказанных услуг" styles={{ body: { minHeight: 120 } }}>
            <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
              PDF-акт для выбранной бани за период
            </Text>
            <Space direction="vertical" style={{ width: '100%' }}>
              <Select
                value={selectedBathhouse || undefined}
                onChange={setSelectedBathhouse}
                options={bathhouseOptions}
                placeholder="Выберите баню"
                style={{ width: '100%' }}
              />
              <Button
                icon={<FilePdfOutlined />}
                onClick={handleActDownload}
                loading={loading === 'act'}
                disabled={!selectedBathhouse}
              >
                Скачать акт
              </Button>
            </Space>
          </Card>
        </Col>

        <Col xs={24} md={8}>
          <Card title="Экспорт 1С" styles={{ body: { minHeight: 120 } }}>
            <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
              Выгрузка в формате XML для 1С (юр. лица)
            </Text>
            <Button
              icon={<FileExcelOutlined />}
              onClick={handleXmlExport}
              loading={loading === 'xml'}
            >
              Скачать XML
            </Button>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
