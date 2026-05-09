import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ETFComparison from '../ETFComparison'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetList = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockGetComparison = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', () => ({
  etfAPI: {
    getList: (...args: unknown[]) => mockGetList(...args),
    getComparison: (...args: unknown[]) => mockGetComparison(...args),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('ETFComparison', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ETFComparison />)
    expect(screen.getAllByText(/比较|对比|Comparison/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ETFComparison />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<ETFComparison />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetList.mockRejectedValueOnce(new Error('Network Error'))
    renderWithProviders(<ETFComparison />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetList.mockResolvedValueOnce(null)
    mockGetComparison.mockResolvedValueOnce(undefined)
    renderWithProviders(<ETFComparison />)
  })
})
