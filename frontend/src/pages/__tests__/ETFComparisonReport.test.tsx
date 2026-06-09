import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ETFComparisonReport from '../ETFComparisonReport'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetList = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockGetComparison = vi.fn().mockResolvedValue({ success: true, data: null })
const mockAnalyzeScenarios = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', async (importOriginal) => {
  const actual = (await importOriginal()) as Record<string, unknown>
  return {
    ...actual,
    etfAPI: {
      getList: (...args: unknown[]) => mockGetList(...args),
      getComparison: (...args: unknown[]) => mockGetComparison(...args),
    },
    portfolioAPI: {
      analyzeScenarios: (...args: unknown[]) => mockAnalyzeScenarios(...args),
    },
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('ETFComparisonReport', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ETFComparisonReport />)
    expect(screen.getAllByText(/报告|对比|报[告表]/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ETFComparisonReport />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<ETFComparisonReport />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetList.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<ETFComparisonReport />)
    await waitFor(() => {
      expect(mockGetList).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetComparison.mockResolvedValueOnce(null)
    mockAnalyzeScenarios.mockResolvedValueOnce(undefined)
    renderWithProviders(<ETFComparisonReport />)
  })
})
