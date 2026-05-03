import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ConfigProvider, App as AntApp } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import TwoFactorChallenge from '@/components/TwoFactorChallenge'

const mockPostAuth2faVerify = vi.fn()
vi.mock('@/api/generated/2fa/2fa', () => ({
  postAuth2faVerify: (...args: unknown[]) => mockPostAuth2faVerify(...args),
}))

function renderChallenge(props: Partial<React.ComponentProps<typeof TwoFactorChallenge>> = {}) {
  const defaultProps = {
    partialToken: 'test-partial-token',
    onSuccess: vi.fn(),
    onCancel: vi.fn(),
    ...props,
  }
  return render(
    <ConfigProvider locale={ruRU}>
      <AntApp>
        <TwoFactorChallenge {...defaultProps} />
      </AntApp>
    </ConfigProvider>,
  )
}

describe('TwoFactorChallenge', () => {
  beforeEach(() => {
    mockPostAuth2faVerify.mockReset()
  })

  it('renders 2FA form', () => {
    renderChallenge()
    expect(screen.getByText('Двухфакторная аутентификация')).toBeInTheDocument()
    expect(screen.getByPlaceholderText('000000')).toBeInTheDocument()
    expect(screen.getByText('Подтвердить')).toBeInTheDocument()
  })

  it('confirm button is disabled when code is incomplete', () => {
    renderChallenge()
    expect(screen.getByText('Подтвердить').closest('button')).toBeDisabled()
  })

  it('calls verify API and onSuccess on valid code', async () => {
    const onSuccess = vi.fn()
    mockPostAuth2faVerify.mockResolvedValue({
      success: true,
      data: { token: 'full-token', user: { id: '1', role: 'client' } },
    })

    renderChallenge({ onSuccess })

    fireEvent.change(screen.getByPlaceholderText('000000'), { target: { value: '123456' } })
    fireEvent.click(screen.getByText('Подтвердить'))

    await waitFor(() => {
      expect(mockPostAuth2faVerify).toHaveBeenCalledWith({
        partial_token: 'test-partial-token',
        code: '123456',
      })
    })

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('full-token', { id: '1', role: 'client' })
    })
  })

  it('calls onCancel when cancel is clicked', () => {
    const onCancel = vi.fn()
    renderChallenge({ onCancel })

    fireEvent.click(screen.getByText('Отмена'))
    expect(onCancel).toHaveBeenCalled()
  })

  it('toggles between TOTP and SMS mode', () => {
    renderChallenge()
    expect(screen.getByText('Введите код из приложения-аутентификатора')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Получить SMS'))
    expect(screen.getByText('Введите код из SMS')).toBeInTheDocument()

    fireEvent.click(screen.getByText('Использовать приложение'))
    expect(screen.getByText('Введите код из приложения-аутентификатора')).toBeInTheDocument()
  })

  it('filters non-digit characters from code', () => {
    renderChallenge()
    const input = screen.getByPlaceholderText('000000')
    fireEvent.change(input, { target: { value: 'abc456' } })
    expect(input).toHaveValue('456')
  })
})
