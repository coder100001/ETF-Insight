import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import PortfolioOptimization from '../PortfolioOptimization'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

vi.mock('../../services/api', async (importOriginal) => {
  const actual = (await importOriginal()) as Record<string, unknown>
  return {
    ...actual,
    optimizationAPI: {
      ...actual.optimizationAPI,
      mptOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
      efficientFrontier: vi.fn().mockResolvedValue({ success: true, data: [] }),
      riskParity: vi.fn().mockResolvedValue({ success: true, data: null }),
      blackLitterman: vi.fn().mockResolvedValue({ success: true, data: null }),
      covarianceMatrix: vi.fn().mockResolvedValue({ success: true, data: {} }),
      etfStatistics: vi.fn().mockResolvedValue({ success: true, data: {} }),
      marketImpliedReturns: vi.fn().mockResolvedValue({ success: true, data: {} }),
    },
    financialConfigAPI: {
      ...actual.financialConfigAPI,
      get: vi.fn().mockResolvedValue({
        success: true,
        data: { risk_free_rate: 0.0435, trading_days_year: 252, default_currency: 'USD' },
      }),
    },
    etfAPI: {
      ...actual.etfAPI,
      getList: vi.fn().mockResolvedValue({ success: true, data: [] }),
      getStatistics: vi.fn().mockResolvedValue({ success: true, data: {} }),
    },
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider>
      <AntdApp>
        <BrowserRouter>{ui}</BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  )
}

describe('PortfolioOptimization', () => {
  it('renders without crashing', () => {
    renderWithProviders(<PortfolioOptimization />)
    expect(screen.getAllByText(/组合优化|Portfolio Optimization/i).length).toBeGreaterThan(0)
  })

  it('displays optimization type selector', () => {
    renderWithProviders(<PortfolioOptimization />)
    expect(screen.getByText(/均值方差|MPT|最大夏普/i)).toBeTruthy()
  })
})
