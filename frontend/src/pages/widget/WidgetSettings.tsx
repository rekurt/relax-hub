import { useState, useMemo } from 'react'
import {
  App,
  Button,
  Card,
  Col,
  ColorPicker,
  Form,
  Input,
  Popconfirm,
  Row,
  Select,
  Switch,
  Typography,
} from 'antd'
import {
  CodeOutlined,
  CopyOutlined,
  ReloadOutlined,
} from '@ant-design/icons'
import {
  useGetMyBathhousesIdWidgetCode,
  useGetMyBathhousesIdWidgetKey,
  usePostMyBathhousesIdWidgetKeyRegenerate,
} from '@/api/generated/widgets/widgets'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'

const { Title, Text } = Typography

const FONT_OPTIONS = [
  { value: 'Inter', label: 'Inter' },
  { value: 'Roboto', label: 'Roboto' },
  { value: 'Open Sans', label: 'Open Sans' },
  { value: 'Montserrat', label: 'Montserrat' },
  { value: 'system-ui', label: 'Системный' },
]

export default function WidgetSettings() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()

  const [color, setColor] = useState('#1890ff')
  const [fontFamily, setFontFamily] = useState('Inter')
  const [showPrice, setShowPrice] = useState(true)
  const [showRating, setShowRating] = useState(true)

  const widgetParams = useMemo(
    () => ({
      color,
      font_family: fontFamily,
      show_price: showPrice,
      show_rating: showRating,
    }),
    [color, fontFamily, showPrice, showRating],
  )

  const { data: widgetCodeData, isLoading: codeLoading } =
    useGetMyBathhousesIdWidgetCode(selectedBathhouseId ?? '', widgetParams, {
      query: { enabled: !!selectedBathhouseId },
    })

  const { data: widgetKeyData, isLoading: keyLoading } =
    useGetMyBathhousesIdWidgetKey(selectedBathhouseId ?? '', {
      query: { enabled: !!selectedBathhouseId },
    })

  const regenerateMutation = usePostMyBathhousesIdWidgetKeyRegenerate({
    mutation: {
      onSuccess: () => {
        message.success('API-ключ перегенерирован')
        queryClient.invalidateQueries({
          queryKey: [`/my/bathhouses/${selectedBathhouseId}/widget-key`],
        })
        queryClient.invalidateQueries({
          queryKey: [`/my/bathhouses/${selectedBathhouseId}/widget-code`],
        })
      },
      onError: () => message.error('Не удалось перегенерировать ключ'),
    },
  })

  const widgetCode = widgetCodeData?.data
  const apiKey = widgetKeyData?.data?.api_key ?? ''

  const handleCopyCode = () => {
    if (!widgetCode?.code) return
    navigator.clipboard.writeText(widgetCode.code).then(
      () => message.success('Код скопирован в буфер обмена'),
      () => message.error('Не удалось скопировать'),
    )
  }

  const handleCopyKey = () => {
    if (!apiKey) return
    navigator.clipboard.writeText(apiKey).then(
      () => message.success('API-ключ скопирован'),
      () => message.error('Не удалось скопировать'),
    )
  }

  const handleRegenerate = () => {
    if (!selectedBathhouseId) return
    regenerateMutation.mutate({ id: selectedBathhouseId })
  }

  if (!selectedBathhouseId) {
    return (
      <div>
        <Title level={3}>Виджет бронирования</Title>
        <div style={{ textAlign: 'center', padding: 40, color: '#999' }}>
          Выберите баню для настройки виджета
        </div>
      </div>
    )
  }

  return (
    <div>
      <Title level={3}>Виджет бронирования</Title>

      <Row gutter={[24, 24]}>
        {/* Settings */}
        <Col xs={24} md={10}>
          <Card title="Настройки виджета" loading={codeLoading}>
            <Form layout="vertical">
              <Form.Item label="Акцентный цвет">
                <ColorPicker
                  value={color}
                  onChange={(_, hex) => setColor(hex)}
                  showText
                />
              </Form.Item>

              <Form.Item label="Шрифт">
                <Select
                  value={fontFamily}
                  onChange={setFontFamily}
                  options={FONT_OPTIONS}
                />
              </Form.Item>

              <Form.Item label="Показывать цену">
                <Switch checked={showPrice} onChange={setShowPrice} />
              </Form.Item>

              <Form.Item label="Показывать рейтинг">
                <Switch checked={showRating} onChange={setShowRating} />
              </Form.Item>
            </Form>
          </Card>

          {/* API Key */}
          <Card title="API-ключ" style={{ marginTop: 24 }} loading={keyLoading}>
            <div style={{ marginBottom: 12 }}>
              <Text type="secondary">
                API-ключ используется для авторизации виджета. Не передавайте его третьим лицам.
              </Text>
            </div>
            <Input.Search
              value={apiKey}
              readOnly
              enterButton={<CopyOutlined />}
              onSearch={handleCopyKey}
              style={{ marginBottom: 12, fontFamily: 'monospace' }}
            />
            <Popconfirm
              title="Перегенерировать API-ключ?"
              description="Текущий ключ перестанет работать. Виджеты с старым ключом перестанут функционировать."
              onConfirm={handleRegenerate}
              okText="Перегенерировать"
              cancelText="Отмена"
            >
              <Button
                icon={<ReloadOutlined />}
                danger
                loading={regenerateMutation.isPending}
              >
                Перегенерировать ключ
              </Button>
            </Popconfirm>
          </Card>
        </Col>

        {/* Embed code + Preview */}
        <Col xs={24} md={14}>
          <Card
            title="Код для вставки"
            extra={
              <Button icon={<CopyOutlined />} onClick={handleCopyCode} disabled={!widgetCode?.code}>
                Копировать
              </Button>
            }
            loading={codeLoading}
          >
            <div
              style={{
                background: '#f5f5f5',
                padding: 16,
                borderRadius: 8,
                fontFamily: 'monospace',
                fontSize: 13,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-all',
                maxHeight: 300,
                overflow: 'auto',
              }}
            >
              {widgetCode?.code || 'Код не сгенерирован'}
            </div>

            <div style={{ marginTop: 16 }}>
              <Text type="secondary">
                Вставьте этот код на ваш сайт в то место, где вы хотите показать форму бронирования.
              </Text>
            </div>
          </Card>

          {/* Preview */}
          <Card title="Предпросмотр" style={{ marginTop: 24 }}>
            <div
              style={{
                border: '2px dashed #d9d9d9',
                borderRadius: 8,
                padding: 32,
                textAlign: 'center',
                minHeight: 200,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
              }}
            >
              <CodeOutlined style={{ fontSize: 48, color: '#d9d9d9', marginBottom: 16 }} />
              <div
                style={{
                  padding: '12px 24px',
                  background: color,
                  color: '#fff',
                  borderRadius: 8,
                  fontFamily,
                  fontSize: 16,
                  marginBottom: 8,
                }}
              >
                Забронировать
              </div>
              <Text type="secondary" style={{ fontFamily }}>
                {showPrice && 'Цена от 2 000 ₽/ч'}
                {showPrice && showRating && ' · '}
                {showRating && '★ 4.8'}
              </Text>
              {!showPrice && !showRating && (
                <Text type="secondary" style={{ fontFamily }}>Виджет бронирования</Text>
              )}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
