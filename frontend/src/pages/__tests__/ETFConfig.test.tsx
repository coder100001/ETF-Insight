import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ETFConfig from '../ETFConfig'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetConfigs = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockToggleStatus = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...actual,
    etfConfigAPI: {
      getConfigs: (...args: unknown[]) => mockGetConfigs(...args),
      toggleStatus: (...args: unknown[]) => mockToggleStatus(...args),
    },
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('ETFConfig', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ETFConfig />)
    expect(screen.getAllByText(/配置|Config|ETF/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ETFConfig />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    mockGetConfigs.mockResolvedValueOnce(null)
    renderWithProviders(<ETFConfig />)
  })
})
