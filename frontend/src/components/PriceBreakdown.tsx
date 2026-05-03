import { Typography, Divider } from '@/components/design/system'
import { formatPrice } from '@/lib/format'

const { Text } = Typography

export interface PriceLineItem {
  label: string
  amount: number
  type?: 'addition' | 'discount' | 'subtotal' | 'total'
}

interface PriceBreakdownProps {
  basePrice: number
  addOns?: { name: string; price: number }[]
  longSessionDiscount?: number
  extraGuestSurcharge?: number
  holidaySurcharge?: number
  holidayName?: string
  seasonalTariffMultiplier?: number
  seasonalTariffName?: string
  serviceFee?: number
  promoDiscount?: number
  certificateDiscount?: number
  walletPayment?: number
  areaAveragePrice?: number
}

export default function PriceBreakdown({
  basePrice,
  addOns,
  longSessionDiscount,
  extraGuestSurcharge,
  holidaySurcharge,
  holidayName,
  seasonalTariffMultiplier,
  seasonalTariffName,
  serviceFee,
  promoDiscount,
  certificateDiscount,
  walletPayment,
  areaAveragePrice,
}: PriceBreakdownProps) {
  const lines: PriceLineItem[] = []
  const showAreaAverage = areaAveragePrice != null
    && areaAveragePrice > 0
    && areaAveragePrice >= Math.round(basePrice * 0.2)
    && areaAveragePrice <= Math.round(basePrice * 5)

  lines.push({ label: 'Базовая стоимость', amount: basePrice })

  if (addOns && addOns.length > 0) {
    addOns.forEach((addon) => {
      lines.push({ label: addon.name, amount: addon.price, type: 'addition' })
    })
  }

  if (extraGuestSurcharge && extraGuestSurcharge > 0) {
    lines.push({ label: 'Доплата за доп. гостей', amount: extraGuestSurcharge, type: 'addition' })
  }

  if (holidaySurcharge && holidaySurcharge > 0) {
    lines.push({
      label: holidayName ? `Праздничная наценка (${holidayName})` : 'Праздничная наценка',
      amount: holidaySurcharge,
      type: 'addition',
    })
  }

  if (seasonalTariffMultiplier && seasonalTariffMultiplier !== 1) {
    const seasonalAmount = Math.round(basePrice * (seasonalTariffMultiplier - 1))
    if (seasonalAmount !== 0) {
      lines.push({
        label: seasonalTariffName ? `Сезонный тариф (${seasonalTariffName})` : 'Сезонный тариф',
        amount: seasonalAmount,
        type: seasonalAmount > 0 ? 'addition' : 'discount',
      })
    }
  }

  if (longSessionDiscount && longSessionDiscount > 0) {
    lines.push({ label: 'Скидка за длительный сеанс', amount: -longSessionDiscount, type: 'discount' })
  }

  if (promoDiscount && promoDiscount > 0) {
    lines.push({ label: 'Промокод', amount: -promoDiscount, type: 'discount' })
  }

  if (certificateDiscount && certificateDiscount > 0) {
    lines.push({ label: 'Сертификат', amount: -certificateDiscount, type: 'discount' })
  }

  if (serviceFee && serviceFee > 0) {
    lines.push({ label: 'Сервисный сбор', amount: serviceFee, type: 'addition' })
  }

  const total = lines.reduce((sum, line) => sum + line.amount, 0)

  if (walletPayment && walletPayment > 0) {
    lines.push({ label: 'Оплата из кошелька', amount: -walletPayment, type: 'discount' })
  }

  const toPay = walletPayment ? Math.max(0, total - walletPayment) : total

  return (
    <div data-testid="price-breakdown">
      {lines.map((line, idx) => (
        <div
          key={idx}
          style={{
            display: 'flex',
            justifyContent: 'space-between',
            marginBottom: 4,
            color: line.type === 'discount' ? '#15803d' : undefined,
          }}
        >
          <Text type={line.type === 'discount' ? 'success' : undefined}>{line.label}</Text>
          <Text type={line.type === 'discount' ? 'success' : undefined} strong={line.type === 'total'}>
            {line.amount < 0 ? `−${formatPrice(Math.abs(line.amount))}` : formatPrice(line.amount)}
          </Text>
        </div>
      ))}

      <Divider style={{ margin: '8px 0' }} />

      <div style={{ display: 'flex', justifyContent: 'space-between' }}>
        <Text strong>Итого</Text>
        <Text strong style={{ fontSize: 16 }}>{formatPrice(Math.max(0, total))}</Text>
      </div>

      {walletPayment && walletPayment > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 4 }}>
          <Text strong>К оплате картой</Text>
          <Text strong style={{ fontSize: 16 }}>{formatPrice(toPay)}</Text>
        </div>
      )}

      {showAreaAverage && (
        <div style={{ marginTop: 8 }}>
          <Text type="secondary">
            Средняя цена в районе: {formatPrice(areaAveragePrice!)}/ч
            {basePrice > areaAveragePrice!
              ? ` (выше на ${Math.round(((basePrice - areaAveragePrice!) / areaAveragePrice!) * 100)}%)`
              : basePrice < areaAveragePrice!
                ? ` (ниже на ${Math.round(((areaAveragePrice! - basePrice) / areaAveragePrice!) * 100)}%)`
                : ' (на уровне средней)'}
          </Text>
        </div>
      )}
    </div>
  )
}
