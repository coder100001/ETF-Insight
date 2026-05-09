import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import PortfolioConfig from '../PortfolioConfig'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetConfigs = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockCreateConfig = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', () => ({
  portfolioConfigAPI: {
    getConfigs: (...args: unknown[]) => mockGetConfigs(...args),
    createConfig: (...args: unknown[]) => mockCreateConfig(...args),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('PortfolioConfig', () => {
  it('renders without crashing', () => {
    renderWithProviders(<PortfolioConfig />)
    expect(screen.getAllByText(/配置|Config|组合/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<PortfolioConfig />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    mockGetConfigs.mockResolvedValueOnce(null)
    renderWithProviders(<PortfolioConfig />)
  })
})