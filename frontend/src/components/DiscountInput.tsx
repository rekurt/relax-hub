import { useEffect, useMemo, useState } from 'react'
import {
  Typography,
  Input,
  Button,
  Space,
  Tag,
  Checkbox,
  InputNumber,
  Collapse,
  App,
} from '@/components/design/system'
import { TagOutlined, GiftOutlined, ThunderboltOutlined, TeamOutlined } from '@/components/design/icons'
import { usePostPromoCodesValidate } from '@/api/generated/promo-codes/promo-codes'
import { useGetCertificatesCodeBalance } from '@/api/generated/certificates/certificates'
import { formatPrice } from '@/lib/format'

const { Text } = Typography

export interface DiscountState {
  promoCode: string
  promoDiscount: number
  promoDiscountType?: string
  certificateCode: string
  certificateAmount: number
  pointsAmount: number
  referralAmount: number
}

export const EMPTY_DISCOUNT_STATE: DiscountState = {
  promoCode: '',
  promoDiscount: 0,
  promoDiscountType: undefined,
  certificateCode: '',
  certificateAmount: 0,
  pointsAmount: 0,
  referralAmount: 0,
}

interface DiscountInputProps {
  bathhouseId: string
  basePrice: number
  totalPrice: number
  availablePoints?: number
  availableReferralBalance?: number
  value: DiscountState
  onChange: (next: DiscountState) => void
}

interface RecentPromo {
  code: string
  type?: string
  discount?: number
  validatedAt: string
}

function readRecentPromos(): RecentPromo[] {
  try {
    const stored = localStorage.getItem('rh_validated_promos')
    if (!stored) return []
    const parsed = JSON.parse(stored) as RecentPromo[]
    return Array.isArray(parsed) ? parsed.slice(0, 3) : []
  } catch {
    return []
  }
}

export default function DiscountInput({
  bathhouseId,
  basePrice,
  totalPrice,
  availablePoints = 0,
  availableReferralBalance = 0,
  value,
  onChange,
}: DiscountInputProps) {
  const { message } = App.useApp()
  const [codeInput, setCodeInput] = useState(value.promoCode || value.certificateCode || '')
  const [error, setError] = useState('')
  const [recentPromos] = useState<RecentPromo[]>(readRecentPromos)
  const [certCheck, setCertCheck] = useState(false)

  const trimmed = codeInput.trim()
  const looksLikeCertificate = /^BANI-/i.test(trimmed)

  const promoMutation = usePostPromoCodesValidate({
    mutation: {
      onSuccess: (response) => {
        const data = response?.data
        if (!data || data.discount == null) {
          setError('Промокод не вернул скидку')
          return
        }
        onChange({
          ...value,
          promoCode: trimmed,
          promoDiscount: data.discount,
          promoDiscountType: data.type,
          certificateCode: '',
          certificateAmount: 0,
        })
        setError('')
        message.success(
          `Промокод применён: ${data.type === 'percentage' ? `${data.discount}%` : formatPrice(data.discount)}`,
        )
      },
      onError: () => {
        setError('Код не найден или истёк')
      },
    },
  })

  const { data: certBalanceData, isFetching: certFetching } = useGetCertificatesCodeBalance(
    looksLikeCertificate ? trimmed : '',
    {
      query: {
        enabled: certCheck && looksLikeCertificate && trimmed.length >= 4,
        retry: false,
      },
    },
  )

  useEffect(() => {
    if (!certCheck) return
    const balance = certBalanceData?.data?.balance
    if (balance == null) return
    if (balance <= 0) {
      setError('Сертификат пуст или недействителен')
      return
    }
    const applied = Math.min(balance, totalPrice)
    onChange({
      ...value,
      certificateCode: trimmed,
      certificateAmount: applied,
      promoCode: '',
      promoDiscount: 0,
      promoDiscountType: undefined,
    })
    setError('')
    message.success(`Сертификат: доступно ${formatPrice(balance)}, будет списано ${formatPrice(applied)}`)
    setCertCheck(false)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [certBalanceData?.data?.balance, certCheck])

  const handleApply = () => {
    if (!trimmed) return
    setError('')
    if (looksLikeCertificate) {
      setCertCheck(true)
    } else {
      promoMutation.mutate({
        data: {
          code: trimmed,
          bathhouse_id: bathhouseId,
          amount: basePrice,
        },
      })
    }
  }

  const handleClear = () => {
    setCodeInput('')
    setError('')
    onChange({
      ...value,
      promoCode: '',
      promoDiscount: 0,
      promoDiscountType: undefined,
      certificateCode: '',
      certificateAmount: 0,
    })
  }

  const usePoints = value.pointsAmount > 0
  const useReferral = value.referralAmount > 0

  const hasAppliedCode = !!value.promoCode || !!value.certificateCode

  const collapsibleItems = useMemo(() => {
    return [
      {
        key: 'discount',
        label: hasAppliedCode ? '✓ Скидка применена' : 'Применить скидку',
        children: (
          <Space orientation="vertical" style={{ width: '100%' }} size={12}>
            <div>
              <Text type="secondary" style={{ fontSize: 12 }}>
                Промокод или код подарочного сертификата (BANI-XXXX-XXXX)
              </Text>
              <Space.Compact style={{ width: '100%', marginTop: 4 }}>
                <Input
                  value={codeInput}
                  onChange={(e) => {
                    setCodeInput(e.target.value)
                    setError('')
                  }}
                  onPressEnter={handleApply}
                  placeholder="Введите код"
                  status={error ? 'error' : hasAppliedCode ? '' : undefined}
                  prefix={looksLikeCertificate ? <GiftOutlined /> : <TagOutlined />}
                />
                {hasAppliedCode ? (
                  <Button onClick={handleClear} danger>
                    Убрать
                  </Button>
                ) : (
                  <Button
                    type="primary"
                    onClick={handleApply}
                    loading={promoMutation.isPending || certFetching}
                    disabled={!trimmed}
                  >
                    Применить
                  </Button>
                )}
              </Space.Compact>
              {error && <Text type="danger" style={{ fontSize: 12 }}>{error}</Text>}
              {value.promoDiscount > 0 && (
                <Text type="success" style={{ fontSize: 12 }}>
                  Скидка по промокоду: {value.promoDiscountType === 'percentage'
                    ? `${value.promoDiscount}%`
                    : formatPrice(value.promoDiscount)}
                </Text>
              )}
              {value.certificateAmount > 0 && (
                <Text type="success" style={{ fontSize: 12 }}>
                  Сертификат: {formatPrice(value.certificateAmount)}
                </Text>
              )}
            </div>

            {recentPromos.length > 0 && !hasAppliedCode && (
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>Недавно проверенные:</Text>
                <Space wrap style={{ marginTop: 4 }}>
                  {recentPromos.map((p) => (
                    <Tag.CheckableTag
                      key={p.code}
                      checked={false}
                      onChange={() => {
                        setCodeInput(p.code)
                        setError('')
                      }}
                    >
                      {p.code}
                    </Tag.CheckableTag>
                  ))}
                </Space>
              </div>
            )}

            {availablePoints > 0 && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
                <Checkbox
                  checked={usePoints}
                  onChange={(e) =>
                    onChange({
                      ...value,
                      pointsAmount: e.target.checked ? Math.min(availablePoints, totalPrice) : 0,
                    })
                  }
                >
                  <ThunderboltOutlined /> Списать баллы лояльности (доступно: {availablePoints})
                </Checkbox>
                {usePoints && (
                  <InputNumber
                    min={1}
                    max={Math.min(availablePoints, totalPrice)}
                    value={value.pointsAmount}
                    onChange={(v) => onChange({ ...value, pointsAmount: v ?? 0 })}
                    size="small"
                    style={{ width: 120 }}
                  />
                )}
              </div>
            )}

            {availableReferralBalance > 0 && (
              <div style={{ display: 'flex', alignItems: 'center', gap: 12, flexWrap: 'wrap' }}>
                <Checkbox
                  checked={useReferral}
                  onChange={(e) =>
                    onChange({
                      ...value,
                      referralAmount: e.target.checked
                        ? Math.min(availableReferralBalance, totalPrice)
                        : 0,
                    })
                  }
                >
                  <TeamOutlined /> Списать реферальный бонус (доступно: {formatPrice(availableReferralBalance)})
                </Checkbox>
                {useReferral && (
                  <InputNumber
                    min={1}
                    max={Math.min(availableReferralBalance, totalPrice)}
                    value={value.referralAmount}
                    onChange={(v) => onChange({ ...value, referralAmount: v ?? 0 })}
                    size="small"
                    style={{ width: 120 }}
                  />
                )}
              </div>
            )}
          </Space>
        ),
      },
    ]
  }, [
    codeInput, error, hasAppliedCode, recentPromos, value, availablePoints, availableReferralBalance,
    promoMutation.isPending, certFetching, looksLikeCertificate, trimmed, totalPrice,
  ])

  return (
    <Collapse
      ghost
      items={collapsibleItems}
      defaultActiveKey={hasAppliedCode || usePoints || useReferral ? ['discount'] : []}
      data-testid="discount-input"
    />
  )
}
