import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import FactorAnalysis from '../FactorAnalysis'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetList = vi.fn().mockResolvedValue({ success: true, data: [] })

vi.mock('../../services/api', () => ({
  etfAPI: {
    getList: (...args: unknown[]) => mockGetList(...args),
  },
  factorAPI: {},
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('FactorAnalysis', () => {
  it('renders without crashing', () => {
    renderWithProviders(<FactorAnalysis />)
    expect(screen.getAllByText(/因子|Factor|分析/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<FactorAnalysis />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<FactorAnalysis />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetList.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<FactorAnalysis />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetList.mockResolvedValueOnce(null)
    renderWithProviders(<FactorAnalysis />)
  })
})