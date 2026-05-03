import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ConfigProvider, App as AntApp } from '@/components/design/system'
import { ruRU } from '@/components/design/system'
import PhoneOTPInput from '@/components/PhoneOTPInput'

function renderPhoneOTPInput(props: Partial<React.ComponentProps<typeof PhoneOTPInput>> = {}) {
  const defaultProps = {
    onVerified: vi.fn(),
    onSendOTP: vi.fn().mockResolvedValue(undefined),
    ...props,
  }
  return render(
    <ConfigProvider locale={ruRU}>
      <AntApp>
        <PhoneOTPInput {...defaultProps} />
      </AntApp>
    </ConfigProvider>,
  )
}

describe('PhoneOTPInput', () => {
  it('renders phone input and send button', () => {
    renderPhoneOTPInput()
    expect(screen.getByPlaceholderText('Телефон')).toBeInTheDocument()
    expect(screen.getByText('Получить код')).toBeInTheDocument()
  })

  it('send button is disabled when phone is empty', () => {
    renderPhoneOTPInput()
    expect(screen.getByText('Получить код').closest('button')).toBeDisabled()
  })

  it('shows code input after sending OTP', async () => {
    const onSendOTP = vi.fn().mockResolvedValue(undefined)
    renderPhoneOTPInput({ onSendOTP })

    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79001234567' } })
    fireEvent.click(screen.getByText('Получить код'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Введите код из SMS')).toBeInTheDocument()
    })
    expect(onSendOTP).toHaveBeenCalledWith('+79001234567')
  })

  it('shows confirm button after OTP is sent', async () => {
    renderPhoneOTPInput()

    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79001234567' } })
    fireEvent.click(screen.getByText('Получить код'))

    await waitFor(() => {
      expect(screen.getByText('Подтвердить')).toBeInTheDocument()
    })
  })

  it('calls onVerified with phone and code', async () => {
    const onVerified = vi.fn()
    renderPhoneOTPInput({ onVerified })

    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79001234567' } })
    fireEvent.click(screen.getByText('Получить код'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Введите код из SMS')).toBeInTheDocument()
    })

    fireEvent.change(screen.getByPlaceholderText('Введите код из SMS'), { target: { value: '123456' } })
    fireEvent.click(screen.getByText('Подтвердить'))

    expect(onVerified).toHaveBeenCalledWith('+79001234567', '123456')
  })

  it('filters non-digit characters from code input', async () => {
    renderPhoneOTPInput()

    fireEvent.change(screen.getByPlaceholderText('Телефон'), { target: { value: '+79001234567' } })
    fireEvent.click(screen.getByText('Получить код'))

    await waitFor(() => {
      expect(screen.getByPlaceholderText('Введите код из SMS')).toBeInTheDocument()
    })

    const codeInput = screen.getByPlaceholderText('Введите код из SMS')
    fireEvent.change(codeInput, { target: { value: 'abc123' } })
    expect(codeInput).toHaveValue('123')
  })
})
