import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { Col, Row } from '@/components/design/system'

describe('design system grid', () => {
  it('preserves auto-sized columns when no grid span is provided', () => {
    render(
      <Row>
        <Col data-testid="auto-col">Auto</Col>
        <Col span={12} data-testid="span-col">Span</Col>
      </Row>,
    )

    const autoCol = screen.getByTestId('auto-col')
    const spanCol = screen.getByTestId('span-col')

    expect(autoCol.style.flex).toBe('')
    expect(autoCol.style.maxWidth).toBe('')
    expect(spanCol.style.flex).toBe('0 0 calc(50% - 0px)')
    expect(spanCol.style.maxWidth).toBe('calc(50% - 0px)')
  })
})
