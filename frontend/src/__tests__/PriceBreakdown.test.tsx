import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import PriceBreakdown from '@/components/PriceBreakdown'

describe('PriceBreakdown', () => {
  it('renders base price and total', () => {
    render(<PriceBreakdown basePrice={300000} />)

    expect(screen.getByText('Базовая стоимость')).toBeInTheDocument()
    expect(screen.getByText('Итого')).toBeInTheDocument()
    expect(screen.getByTestId('price-breakdown')).toBeInTheDocument()
  })

  it('renders add-ons', () => {
    render(
      <PriceBreakdown
        basePrice={300000}
        addOns={[
          { name: 'Веники', price: 50000 },
          { name: 'Простыня', price: 20000 },
        ]}
      />,
    )

    expect(screen.getByText('Веники')).toBeInTheDocument()
    expect(screen.getByText('Простыня')).toBeInTheDocument()
  })

  it('renders discounts with negative sign', () => {
    render(
      <PriceBreakdown
        basePrice={300000}
        longSessionDiscount={30000}
        promoDiscount={15000}
      />,
    )

    expect(screen.getByText('Скидка за длительный сеанс')).toBeInTheDocument()
    expect(screen.getByText('Промокод')).toBeInTheDocument()
  })

  it('renders service fee', () => {
    render(<PriceBreakdown basePrice={300000} serviceFee={30000} />)

    expect(screen.getByText('Сервисный сбор')).toBeInTheDocument()
  })

  it('renders area average comparison when higher', () => {
    render(<PriceBreakdown basePrice={300000} areaAveragePrice={200000} />)

    expect(screen.getByText(/Средняя цена в районе/)).toBeInTheDocument()
    expect(screen.getByText(/выше на 50%/)).toBeInTheDocument()
  })

  it('renders area average comparison when lower', () => {
    render(<PriceBreakdown basePrice={200000} areaAveragePrice={300000} />)

    expect(screen.getByText(/ниже на 33%/)).toBeInTheDocument()
  })

  it('renders certificate discount', () => {
    render(<PriceBreakdown basePrice={300000} certificateDiscount={100000} />)

    expect(screen.getByText('Сертификат')).toBeInTheDocument()
  })

  it('renders wallet payment and card amount', () => {
    render(<PriceBreakdown basePrice={300000} walletPayment={100000} />)

    expect(screen.getByText('Оплата из кошелька')).toBeInTheDocument()
    expect(screen.getByText('К оплате картой')).toBeInTheDocument()
  })

  it('renders holiday surcharge', () => {
    render(<PriceBreakdown basePrice={300000} holidaySurcharge={45000} holidayName="Новый год" />)

    expect(screen.getByText('Праздничная наценка (Новый год)')).toBeInTheDocument()
  })

  it('renders seasonal tariff', () => {
    render(<PriceBreakdown basePrice={300000} seasonalTariffMultiplier={1.3} seasonalTariffName="Лето" />)

    expect(screen.getByText('Сезонный тариф (Лето)')).toBeInTheDocument()
  })

  it('renders extra guest surcharge', () => {
    render(<PriceBreakdown basePrice={300000} extraGuestSurcharge={20000} />)

    expect(screen.getByText('Доплата за доп. гостей')).toBeInTheDocument()
  })
})
