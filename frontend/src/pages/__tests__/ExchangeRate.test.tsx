import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ExchangeRate from '../ExchangeRate'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetRates = vi.fn().mockResolvedValue({ success: true, data: null })
const mockConvert = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', () => ({
  exchangeRateAPI: {
    getRates: (...args: unknown[]) => mockGetRates(...args),
    convert: (...args: unknown[]) => mockConvert(...args),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('ExchangeRate', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ExchangeRate />)
    expect(screen.getAllByText(/汇率|Exchange|Rate/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ExchangeRate />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    mockGetRates.mockResolvedValueOnce(null)
    renderWithProviders(<ExchangeRate />)
  })
})
