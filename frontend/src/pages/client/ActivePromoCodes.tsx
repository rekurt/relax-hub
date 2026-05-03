import { useState } from 'react'
import {
  Typography,
  Card,
  Input,
  Button,
  Alert,
  Tag,
  Space,
  Descriptions,
  List,
} from '@/components/design/system'
import {
  TagOutlined,
  CopyOutlined,
  SearchOutlined,
  CheckCircleOutlined,
  GiftOutlined,
  PercentageOutlined,
  ClockCircleOutlined,
} from '@/components/design/icons'
import { App } from '@/components/design/system'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import type { InternalHandlerValidatePromoResponse } from '@/api/generated/model'
import { formatPrice } from '@/lib/format'
import PageHeader from '@/components/PageHeader'

const { Text, Paragraph } = Typography

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
      const stored = localStorage.getItem('rh_validated_promos')
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
            localStorage.setItem('rh_validated_promos', JSON.stringify(updated))
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
      localStorage.setItem('rh_validated_promos', JSON.stringify(updated))
      return updated
    })
  }

  return (
    <div className="rh-stack">
      <PageHeader
        eyebrow="Выгода"
        title="Промокоды"
        description="Проверяйте промокоды заранее и храните последние успешные проверки."
      />

      <Card title="Проверить промокод" className="rh-admin-detail-card">
        <Paragraph type="secondary" className="rh-card-intro-text">
          Введите промокод, полученный от бани или по акции, чтобы узнать размер скидки.
        </Paragraph>
        <Space.Compact className="rh-promo-check-control">
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
            title="Промокод недействителен"
            description="Проверьте правильность ввода или срок действия промокода."
            type="error"
            showIcon
            className="rh-alert-spaced"
          />
        )}
      </Card>

      <Card title="Проверенные промокоды" className="rh-admin-detail-card">
        {validatedPromos.length === 0 ? (
          <div className="rh-admin-empty-state">
            <GiftOutlined className="rh-empty-state-icon" />
            <div className="rh-admin-empty-state__title">Нет проверенных промокодов</div>
            <p className="rh-admin-empty-state__text">
              Введите код выше, чтобы сохранить его в истории проверок.
            </p>
          </div>
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
                    <div className="rh-promo-avatar">
                      {promo.type === 'percentage' ? (
                        <PercentageOutlined className="rh-promo-avatar__icon rh-promo-avatar__icon--primary" />
                      ) : (
                        <GiftOutlined className="rh-promo-avatar__icon rh-promo-avatar__icon--success" />
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
                        <ClockCircleOutlined className="rh-inline-icon" />
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
