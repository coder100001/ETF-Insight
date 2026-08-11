import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ETFDashboard from '../ETFDashboard'
import { useETFStore } from '../../stores/etfStore'

// echarts 在 jsdom 中无 canvas，需要 mock（页面通过 lib/echarts 使用 echarts）
vi.mock('../../lib/echarts', () => ({
  default: {
    init: vi.fn(() => ({
      setOption: vi.fn(),
      resize: vi.fn(),
      dispose: vi.fn(),
      clear: vi.fn(),
      getWidth: vi.fn(() => 0),
      getHeight: vi.fn(() => 0),
      on: vi.fn(),
      off: vi.fn(),
      showLoading: vi.fn(),
      hideLoading: vi.fn(),
    })),
    dispose: vi.fn(),
    registerTheme: vi.fn(),
    connect: vi.fn(),
    graphic: { LinearGradient: vi.fn() },
  },
}))

const mockGetList = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockGetHistory = vi.fn().mockResolvedValue({ success: true, data: [] })

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  // 重置全局 store，避免测试间状态泄漏（hasInitialized/loading）
  useETFStore.setState({
    ...useETFStore.getInitialState(),
    etfList: [],
    hasInitialized: false,
    loading: false,
    statsLoading: false,
    error: null,
  })
  mockGetList.mockClear()
  mockGetHistory.mockClear()
})

vi.mock('../../services/api', async (importOriginal) => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const actual = (await importOriginal()) as any
  return {
    ...actual,
    etfAPI: {
      getList: (...args: unknown[]) => mockGetList(...args),
      getHistory: (...args: unknown[]) => mockGetHistory(...args),
    },
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('ETFDashboard', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ETFDashboard />)
    expect(screen.getAllByText(/ETF|仪表盘|Dashboard|看板/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', async () => {
    // deferred promise：手动控制请求完成时机，避免立即 resolve 导致 loading 瞬态不可测
    let resolveList!: (v: unknown) => void
    mockGetList.mockImplementationOnce(
      () => new Promise(resolve => { resolveList = resolve })
    )
    renderWithProviders(<ETFDashboard />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
    resolveList({ success: true, data: [] })
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<ETFDashboard />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetList.mockRejectedValueOnce(new Error('Network Error'))
    renderWithProviders(<ETFDashboard />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetList.mockResolvedValueOnce(null)
    mockGetHistory.mockResolvedValueOnce(undefined)
    renderWithProviders(<ETFDashboard />)
  })
})
