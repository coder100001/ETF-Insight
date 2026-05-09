import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import Dashboard from '../Dashboard'

vi.mock('echarts', () => ({
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

  it('shows today stats', () => {
    renderWithProviders(<Dashboard />)
    expect(screen.getAllByText(/\d+/).length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    renderWithProviders(<Dashboard />)
  })
})
