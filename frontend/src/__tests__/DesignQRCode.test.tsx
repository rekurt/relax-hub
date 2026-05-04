import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { QRCode } from '@/components/design/system'

describe('Design QRCode', () => {
  it('renders a real SVG QR pattern for otpauth URLs', () => {
    const value = 'otpauth://totp/RelaxHUB:admin@example.com?secret=UJSLUELG3VIFW7TW3Y6OYX7CQFM5CDRN&issuer=RelaxHUB&period=30&digits=6'

    const { container } = render(<QRCode value={value} size={176} />)

    const qr = screen.getByLabelText(value)
    const svg = container.querySelector('svg.rh-qr__svg')
    const path = container.querySelector('svg.rh-qr__svg path')

    expect(qr).toHaveClass('rh-qr')
    expect(svg).toBeInTheDocument()
    expect(path?.getAttribute('d')?.length).toBeGreaterThan(1000)
  })

  it('keeps an empty fallback when no value is available', () => {
    const { container } = render(<QRCode value="" />)

    expect(container.querySelector('.rh-qr__empty')).toBeInTheDocument()
    expect(container.querySelector('svg')).not.toBeInTheDocument()
  })
})
