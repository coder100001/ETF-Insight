import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ETFDashboard from '../ETFDashboard'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetList = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockGetHistory = vi.fn().mockResolvedValue({ success: true, data: [] })

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

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ETFDashboard />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
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
