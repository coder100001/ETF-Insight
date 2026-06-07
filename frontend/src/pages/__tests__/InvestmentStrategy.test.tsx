import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import InvestmentStrategy from '../InvestmentStrategy'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetList = vi.fn().mockResolvedValue({ success: true, data: [] })

vi.mock('../../services/api', () => ({
  etfAPI: {
    getList: (...args: unknown[]) => mockGetList(...args),
  },
  portfolioAPI: {
    analyzeScenarios: vi.fn().mockResolvedValue({ success: true, data: {} }),
    getDefaultTemplates: vi.fn().mockResolvedValue({ success: true, data: [] }),
    analyzeRisk: vi.fn().mockResolvedValue({ success: true, data: {} }),
    optimize: vi.fn().mockResolvedValue({ success: true, data: {} }),
    getEfficientFrontier: vi.fn().mockResolvedValue({ success: true, data: [] }),
  },
  portfolioConfigAPI: {
    getAll: vi.fn().mockResolvedValue({ success: true, data: [] }),
    getById: vi.fn().mockResolvedValue({ success: true, data: {} }),
    create: vi.fn().mockResolvedValue({ success: true, data: {} }),
    update: vi.fn().mockResolvedValue({ success: true, data: {} }),
    delete: vi.fn().mockResolvedValue({ success: true, data: null }),
    toggleStatus: vi.fn().mockResolvedValue({ success: true, data: {} }),
    analyze: vi.fn().mockResolvedValue({ success: true, data: {} }),
  },
  financialConfigAPI: {
    get: vi.fn().mockResolvedValue({
      success: true,
      data: { risk_free_rate: 0.0435, trading_days_year: 252, default_currency: 'USD' },
    }),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('InvestmentStrategy', () => {
  it('renders without crashing', () => {
    renderWithProviders(<InvestmentStrategy />)
    expect(screen.getAllByText(/投资|策略|Strategy|Invest/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<InvestmentStrategy />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<InvestmentStrategy />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetList.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<InvestmentStrategy />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetList.mockResolvedValueOnce(null)
    renderWithProviders(<InvestmentStrategy />)
  })
})
