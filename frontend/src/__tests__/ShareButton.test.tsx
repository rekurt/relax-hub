import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter } from 'react-router-dom'
import { App as AntApp, ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import ShareButton from '@/components/ShareButton'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <AntApp>
          <MemoryRouter>{ui}</MemoryRouter>
        </AntApp>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('ShareButton', () => {
  beforeEach(() => {
    vi.restoreAllMocks()
    Object.defineProperty(navigator, 'share', { value: undefined, writable: true, configurable: true })
    Object.defineProperty(navigator, 'clipboard', {
      value: { writeText: vi.fn().mockResolvedValue(undefined) },
      writable: true,
      configurable: true,
    })
  })

  it('renders with share text', () => {
    renderWithProviders(<ShareButton url="https://example.com" />)
    expect(screen.getByText('Поделиться')).toBeInTheDocument()
  })

  it('copies to clipboard when Web Share API is not available', async () => {
    renderWithProviders(<ShareButton url="https://example.com/test" />)
    fireEvent.click(screen.getByText('Поделиться'))

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('https://example.com/test')
    })
  })

  it('calls onBeforeShare before sharing', async () => {
    const onBeforeShare = vi.fn().mockResolvedValue('https://custom-url.com')

    renderWithProviders(
      <ShareButton url="https://fallback.com" onBeforeShare={onBeforeShare} />,
    )
    fireEvent.click(screen.getByText('Поделиться'))

    await waitFor(() => {
      expect(onBeforeShare).toHaveBeenCalled()
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('https://custom-url.com')
    })
  })

  it('uses original URL if onBeforeShare returns undefined', async () => {
    const onBeforeShare = vi.fn().mockResolvedValue(undefined)

    renderWithProviders(
      <ShareButton url="https://original.com" onBeforeShare={onBeforeShare} />,
    )
    fireEvent.click(screen.getByText('Поделиться'))

    await waitFor(() => {
      expect(navigator.clipboard.writeText).toHaveBeenCalledWith('https://original.com')
    })
  })

  it('uses Web Share API when available', async () => {
    const shareFn = vi.fn().mockResolvedValue(undefined)
    Object.defineProperty(navigator, 'share', { value: shareFn, writable: true, configurable: true })

    renderWithProviders(
      <ShareButton url="https://example.com" title="Test" text="Hello" />,
    )
    fireEvent.click(screen.getByText('Поделиться'))

    await waitFor(() => {
      expect(shareFn).toHaveBeenCalledWith({
        title: 'Test',
        text: 'Hello',
        url: 'https://example.com',
      })
    })
  })
})
