import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import RiskAnalysis from '../RiskAnalysis'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockAnalyzeRisk = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', () => ({
  portfolioAPI: {
    analyzeRisk: (...args: unknown[]) => mockAnalyzeRisk(...args),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('RiskAnalysis', () => {
  it('renders without crashing', { timeout: 15000 }, () => {
    renderWithProviders(<RiskAnalysis />)
    expect(screen.getAllByText(/风险|Risk|分析/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<RiskAnalysis />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<RiskAnalysis />)
    await waitFor(() => {
      expect(mockAnalyzeRisk).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockAnalyzeRisk.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<RiskAnalysis />)
    await waitFor(() => {
      expect(mockAnalyzeRisk).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockAnalyzeRisk.mockResolvedValueOnce(null)
    renderWithProviders(<RiskAnalysis />)
  })
})
