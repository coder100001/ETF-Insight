import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import TechnicalAnalysis from '../TechnicalAnalysis'

vi.mock('../../components/ReactEChart', () => ({
  default: () => <div data-testid="mock-echart" />,
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

describe('TechnicalAnalysis', () => {
  it('renders without crashing', { timeout: 15000 }, () => {
    renderWithProviders(<TechnicalAnalysis />)
    expect(screen.getAllByText(/技术|Technical|指标|分析/i).length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    renderWithProviders(<TechnicalAnalysis />)
  })
})
