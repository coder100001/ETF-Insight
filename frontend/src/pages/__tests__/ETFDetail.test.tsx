import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { Route, Routes, MemoryRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ETFDetail from '../ETFDetail'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetRealtimeData = vi.fn().mockResolvedValue({ success: true, data: null })
const mockGetMetrics = vi.fn().mockResolvedValue({ success: true, data: null })
const mockGetHistory = vi.fn().mockResolvedValue({ success: true, data: [] })

vi.mock('../../services/api', () => ({
  etfAPI: {
    getRealtimeData: (...args: unknown[]) => mockGetRealtimeData(...args),
    getMetrics: (...args: unknown[]) => mockGetMetrics(...args),
    getHistory: (...args: unknown[]) => mockGetHistory(...args),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider>
      <AntdApp>
        <MemoryRouter initialEntries={['/etf/SCHD']}>
          <Routes>
            <Route path="/etf/:symbol" element={ui} />
          </Routes>
        </MemoryRouter>
      </AntdApp>
    </ConfigProvider>
  )
}

describe('ETFDetail', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ETFDetail />)
    expect(screen.getAllByText(/详情|Detail|ETF|SCHD/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ETFDetail />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles error state from API', () => {
    mockGetRealtimeData.mockRejectedValueOnce(new Error('Network Error'))
    renderWithProviders(<ETFDetail />)
  })

  it('handles boundary conditions without crashing', () => {
    mockGetRealtimeData.mockResolvedValueOnce(null)
    renderWithProviders(<ETFDetail />)
  })
})