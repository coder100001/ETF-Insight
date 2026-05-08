import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import BlackLittermanConfig from '../BlackLittermanConfig'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

vi.mock('../../services/api', () => ({
  blackLittermanAPI: {
    createConfig: vi.fn().mockResolvedValue({ success: true, data: null }),
    calculate: vi.fn().mockResolvedValue({ success: true, data: null }),
    getPosteriorReturns: vi.fn().mockResolvedValue({ success: true, data: null }),
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

describe('BlackLittermanConfig', () => {
  it('renders without crashing', () => {
    renderWithProviders(<BlackLittermanConfig />)
    expect(screen.getAllByText(/Black-Litterman|BL模型/i).length).toBeGreaterThan(0)
  })
})
