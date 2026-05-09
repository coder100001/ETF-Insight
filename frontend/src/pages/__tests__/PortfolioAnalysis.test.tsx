import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import PortfolioAnalysis from '../PortfolioAnalysis'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockFetch = vi.fn().mockResolvedValue({
  json: () => Promise.resolve({ success: true, data: [] }),
})

vi.stubGlobal('fetch', mockFetch)

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('PortfolioAnalysis', () => {
  it('renders without crashing', () => {
    renderWithProviders(<PortfolioAnalysis />)
    expect(screen.getAllByText(/组合|Portfolio|分析/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<PortfolioAnalysis />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<PortfolioAnalysis />)
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockFetch.mockRejectedValueOnce(new Error('Network Error'))
    renderWithProviders(<PortfolioAnalysis />)
    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockFetch.mockResolvedValueOnce({
      json: () => Promise.resolve({ success: false, data: null }),
    })
    renderWithProviders(<PortfolioAnalysis />)
  })
})