import { fireEvent, render } from '@testing-library/react'
import { describe, expect, it, beforeEach, afterEach, vi } from 'vitest'
import { Modal } from '@/components/design/system'

describe('Design Modal body scroll lock', () => {
  beforeEach(() => {
    document.body.style.overflow = ''
  })
  afterEach(() => {
    document.body.style.overflow = ''
  })

  it('locks body scroll while open and restores when closed', () => {
    const { rerender } = render(<Modal open={false}>content</Modal>)
    expect(document.body.style.overflow).toBe('')

    rerender(<Modal open>content</Modal>)
    expect(document.body.style.overflow).toBe('hidden')

    rerender(<Modal open={false}>content</Modal>)
    expect(document.body.style.overflow).toBe('')
  })

  it('keeps body locked while any of two stacked modals is still open', () => {
    function Stack({ outerOpen, innerOpen }: { outerOpen: boolean; innerOpen: boolean }) {
      return (
        <>
          <Modal open={outerOpen}>outer</Modal>
          <Modal open={innerOpen}>inner</Modal>
        </>
      )
    }

    const { rerender } = render(<Stack outerOpen innerOpen />)
    expect(document.body.style.overflow).toBe('hidden')

    // Closing the inner modal must NOT re-enable page scroll while the outer
    // is still open — the previous implementation would have unlocked here.
    rerender(<Stack outerOpen innerOpen={false} />)
    expect(document.body.style.overflow).toBe('hidden')

    rerender(<Stack outerOpen={false} innerOpen={false} />)
    expect(document.body.style.overflow).toBe('')
  })

  it('routes Escape only to the topmost modal when modals are stacked', () => {
    const onCancelOuter = vi.fn()
    const onCancelInner = vi.fn()

    render(
      <>
        <Modal open onCancel={onCancelOuter}>outer</Modal>
        <Modal open onCancel={onCancelInner}>inner</Modal>
      </>,
    )

    fireEvent.keyDown(document, { key: 'Escape' })

    expect(onCancelInner).toHaveBeenCalledTimes(1)
    expect(onCancelOuter).not.toHaveBeenCalled()
  })
})
