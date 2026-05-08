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

vi.mock('../../services/api', () => ({
  optimizationAPI: {
    mptOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
    efficientFrontier: vi.fn().mockResolvedValue({ success: true, data: [] }),
    riskParityOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
    blackLittermanOptimize: vi.fn().mockResolvedValue({ success: true, data: null }),
  },
}))

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
