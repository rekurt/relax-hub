import { useState } from 'react'
import {
  Typography,
  Card,
  Input,
  Button,
  Alert,
  Tag,
  Empty,
  Space,
  Descriptions,
  List,
} from 'antd'
import {
  TagOutlined,
  CopyOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  GiftOutlined,
  PercentageOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import { App } from 'antd'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import type { InternalHandlerValidatePromoResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'

const { Title, Text, Paragraph } = Typography

interface ValidatedPromo extends InternalHandlerValidatePromoResponse {
  validatedAt: string
}

const PROMO_TYPE_LABELS: Record<string, string> = {
  percentage: 'Процентная скидка',
  fixed_amount: 'Фиксированная скидка',
  free_hour: 'Бесплатный час',
  free_addon: 'Бесплатная доп. услуга',
}

const PROMO_TYPE_COLORS: Record<string, string> = {
  percentage: 'blue',
  fixed_amount: 'green',
  free_hour: 'orange',
  free_addon: 'purple',
}

function getDiscountDisplay(promo: InternalHandlerValidatePromoResponse): string {
  switch (promo.type) {
    case 'percentage':
      return `${promo.value}%`
    case 'fixed_amount':
      return formatPrice(promo.value ?? 0)
    case 'free_hour':
      return `${promo.value ?? 1} ч бесплатно`
    case 'free_addon':
      return 'Бесплатная доп. услуга'
    default:
      return String(promo.value ?? '')
  }
}

export default function ActivePromoCodes() {
  const { message } = App.useApp()
  const [code, setCode] = useState('')
  const [validatedPromos, setValidatedPromos] = useState<ValidatedPromo[]>(() => {
    try {
      const stored = localStorage.getItem('bani_validated_promos')
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  })

  const validateMutation = usePostPromoCodesValidate({
    mutation: {
      onSuccess: (response) => {
        const data = response?.data
        if (data) {
          const newPromo: ValidatedPromo = {
            ...data,
            validatedAt: new Date().toISOString(),
          }
          setValidatedPromos((prev) => {
            const filtered = prev.filter((p) => p.code !== data.code)
            const updated = [newPromo, ...filtered].slice(0, 20)
            localStorage.setItem('bani_validated_promos', JSON.stringify(updated))
            return updated
          })
          message.success('Промокод действителен!')
          setCode('')
        }
      },
      onError: () => {
        message.error('Промокод недействителен или истёк')
      },
    },
  })

  const handleValidate = () => {
    const trimmed = code.trim()
    if (!trimmed) return
    validateMutation.mutate({ data: { code: trimmed } })
  }

  const handleCopy = (promoCode: string) => {
    navigator.clipboard.writeText(promoCode).then(
      () => message.success('Промокод скопирован'),
      () => message.error('Не удалось скопировать'),
    )
  }

  const handleRemovePromo = (promoCode: string) => {
    setValidatedPromos((prev) => {
      const updated = prev.filter((p) => p.code !== promoCode)
      localStorage.setItem('bani_validated_promos', JSON.stringify(updated))
      return updated
    })
  }

  return (
    <div>
      <Title level={2}>
        <TagOutlined /> Промокоды
      </Title>

      <Card title="Проверить промокод" style={{ marginBottom: 24 }}>
        <Paragraph type="secondary">
          Введите промокод, полученный от бани или по акции, чтобы узнать размер скидки.
        </Paragraph>
        <Space.Compact style={{ width: '100%', maxWidth: 500 }}>
          <Input
            placeholder="Введите промокод"
            value={code}
            onChange={(e) => setCode(e.target.value.toUpperCase())}
            onPressEnter={handleValidate}
            prefix={<TagOutlined />}
            size="large"
            allowClear
          />
          <Button
            type="primary"
            icon={<SearchOutlined />}
            onClick={handleValidate}
            loading={validateMutation.isPending}
            size="large"
          >
            Проверить
          </Button>
        </Space.Compact>

        {validateMutation.isError && (
          <Alert
            message="Промокод недействителен"
            description="Проверьте правильность ввода или срок действия промокода."
            type="error"
            showIcon
            style={{ marginTop: 16 }}
          />
        )}
      </Card>

      <Card title="Проверенные промокоды">
        {validatedPromos.length === 0 ? (
          <Empty
            image={<GiftOutlined style={{ fontSize: 48, color: '#999' }} />}
            description="Нет проверенных промокодов"
          />
        ) : (
          <List
            dataSource={validatedPromos}
            renderItem={(promo) => (
              <List.Item
                key={promo.code}
                actions={[
                  <Button
                    key="copy"
                    type="link"
                    icon={<CopyOutlined />}
                    onClick={() => handleCopy(promo.code ?? '')}
                  >
                    Копировать
                  </Button>,
                  <Button
                    key="remove"
                    type="link"
                    danger
                    size="small"
                    onClick={() => handleRemovePromo(promo.code ?? '')}
                  >
                    Убрать
                  </Button>,
                ]}
              >
                <List.Item.Meta
                  avatar={
                    <div
                      style={{
                        width: 48,
                        height: 48,
                        borderRadius: 8,
                        background: '#f5f5f5',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                      }}
                    >
                      {promo.type === 'percentage' ? (
                        <PercentageOutlined style={{ fontSize: 24, color: '#1677ff' }} />
                      ) : (
                        <GiftOutlined style={{ fontSize: 24, color: '#52c41a' }} />
                      )}
                    </div>
                  }
                  title={
                    <Space>
                      <Text strong copyable={{ text: promo.code }}>
                        {promo.code}
                      </Text>
                      <Tag color={PROMO_TYPE_COLORS[promo.type ?? ''] ?? 'default'}>
                        {PROMO_TYPE_LABELS[promo.type ?? ''] ?? promo.type}
                      </Tag>
                      <Tag icon={<CheckCircleOutlined />} color="success">
                        Действителен
                      </Tag>
                    </Space>
                  }
                  description={
                    <Descriptions size="small" column={1}>
                      <Descriptions.Item label="Скидка">
                        <Text strong>{getDiscountDisplay(promo)}</Text>
                      </Descriptions.Item>
                      {promo.discount != null && promo.discount > 0 && (
                        <Descriptions.Item label="Расчётная скидка">
                          {formatPrice(promo.discount)}
                        </Descriptions.Item>
                      )}
                      <Descriptions.Item label="Проверен">
                        <ClockCircleOutlined style={{ marginRight: 4 }} />
                        {new Date(promo.validatedAt).toLocaleDateString('ru-RU')}
                      </Descriptions.Item>
                    </Descriptions>
                  }
                />
              </List.Item>
            )}
          />
        )}
      </Card>
    </div>
  )
}
