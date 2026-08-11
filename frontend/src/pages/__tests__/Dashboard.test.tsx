import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import Dashboard from '../Dashboard'

// 页面通过 lib/echarts（按需注册）使用 echarts，需 mock lib 模块而非 'echarts' 顶层
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

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('Dashboard', () => {
  it('renders without crashing', () => {
    renderWithProviders(<Dashboard />)
    expect(screen.getAllByText(/仪表板|Dashboard/i).length).toBeGreaterThan(0)
  })

  it('shows stat cards from store default data', () => {
    renderWithProviders(<Dashboard />)
    // store 默认渲染 ETF 列表，统计卡片应显示 ETF 数量
    expect(screen.getByText('ETF 基金数量')).toBeInTheDocument()
    // 热门 ETF 表格应渲染默认列表中的 symbol
    expect(screen.getAllByText(/VTI|VOO|QQQ/i).length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    renderWithProviders(<Dashboard />)
  })
})
