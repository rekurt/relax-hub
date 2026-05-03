import { useCallback, useState } from 'react'
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
} from '@/components/design/system'
import {
  CodeOutlined,
  CopyOutlined,
  ReloadOutlined,
} from '@/components/design/icons'
import {
  useGetMyBathhousesIdWidgetCode,
  useGetMyBathhousesIdWidgetKey,
  usePostMyBathhousesIdWidgetKeyRegenerate,
} from '@/api/generated/widgets/widgets'
import { useBathhouseStore } from '@/stores/bathhouse'
import { useQueryClient } from '@tanstack/react-query'
import PageHeader from '@/components/PageHeader'

const { Text } = Typography

const FONT_OPTIONS = [
  { value: 'Manrope', label: 'Manrope' },
  { value: 'Avenir Next', label: 'Avenir Next' },
  { value: 'Segoe UI', label: 'Segoe UI' },
  { value: 'system-ui', label: 'Системный' },
]

export default function WidgetSettings() {
  const selectedBathhouseId = useBathhouseStore((s) => s.selectedBathhouseId)
  const queryClient = useQueryClient()
  const { message } = App.useApp()

  const [color, setColor] = useState('#0f766e')
  const [fontFamily, setFontFamily] = useState('Manrope')
  const [showPrice, setShowPrice] = useState(true)
  const [showRating, setShowRating] = useState(true)

  const widgetParams = {
    color,
    font_family: fontFamily,
    show_price: showPrice,
    show_rating: showRating,
  }

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
  const applyPreviewButtonStyles = useCallback((node: HTMLDivElement | null) => {
    if (!node) return
    node.style.background = color
    node.style.fontFamily = fontFamily
  }, [color, fontFamily])

  const applyPreviewFont = useCallback((node: HTMLElement | null) => {
    if (!node) return
    node.style.fontFamily = fontFamily
  }, [fontFamily])

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
      <div className="rh-stack">
        <PageHeader
          eyebrow="Интеграции"
          title="Виджет бронирования"
          description="Выберите объект, чтобы настроить внешний вид и получить код вставки."
          size="compact"
        />
        <Card className="rh-admin-detail-card">
          <div className="rh-admin-empty-state">
            <div className="rh-admin-empty-state__title">Выберите баню для настройки виджета</div>
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Интеграции"
        title="Виджет бронирования"
        description="Настройте виджет, API-ключ, код вставки и быстрый предпросмотр перед публикацией."
        size="compact"
      />

      <Row gutter={[24, 24]}>
        {/* Settings */}
        <Col xs={24} md={10}>
          <Card title="Настройки виджета" loading={codeLoading} className="rh-admin-detail-card">
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
          <Card title="API-ключ" className="rh-admin-detail-card rh-section-offset" loading={keyLoading}>
            <div className="rh-card-intro-text">
              <Text type="secondary">
                API-ключ используется для авторизации виджета. Не передавайте его третьим лицам.
              </Text>
            </div>
            <Input.Search
              value={apiKey}
              readOnly
              enterButton={<CopyOutlined />}
              onSearch={handleCopyKey}
              className="rh-widget-key-control"
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
            className="rh-admin-detail-card"
          >
            <div className="rh-widget-code-block">
              {widgetCode?.code || 'Код не сгенерирован'}
            </div>

            <div className="rh-section-offset">
              <Text type="secondary">
                Вставьте этот код на ваш сайт в то место, где вы хотите показать форму бронирования.
              </Text>
            </div>
          </Card>

          {/* Preview */}
          <Card title="Предпросмотр" className="rh-admin-detail-card rh-section-offset">
            <div className="rh-widget-preview">
              <CodeOutlined className="rh-widget-preview__icon" />
              <div
                ref={applyPreviewButtonStyles}
                className="rh-widget-preview__button"
              >
                Забронировать
              </div>
              <span ref={applyPreviewFont} className="rh-widget-preview__meta">
                {showPrice && 'Цена от 2 000 ₽/ч'}
                {showPrice && showRating && ' · '}
                {showRating && '★ 4.8'}
              </span>
              {!showPrice && !showRating && (
                <span ref={applyPreviewFont} className="rh-widget-preview__meta">Виджет бронирования</span>
              )}
            </div>
          </Card>
        </Col>
      </Row>
    </div>
  )
}
