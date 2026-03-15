import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConfigProvider } from 'antd'
import ruRU from 'antd/locale/ru_RU'
import App from '../App'

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ConfigProvider locale={ruRU}>
        <BrowserRouter>{ui}</BrowserRouter>
      </ConfigProvider>
    </QueryClientProvider>,
  )
}

describe('App', () => {
  it('renders the root route', () => {
    renderWithProviders(<App />)
    expect(
      screen.getByText('Личный кабинет владельца бань'),
    ).toBeInTheDocument()
  })
})
