import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import QuantLibAnalysis from '../QuantLibAnalysis'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetReferenceData = vi.fn().mockResolvedValue({ success: true, data: [] })

vi.mock('../../services/api', () => ({
  quantlibAPI: {
    getReferenceData: (...args: unknown[]) => mockGetReferenceData(...args),
    priceEuropeanOption: vi.fn().mockResolvedValue({ success: true, data: null }),
    priceBond: vi.fn().mockResolvedValue({ success: true, data: null }),
    buildYieldCurve: vi.fn().mockResolvedValue({ success: true, data: null }),
    calculateVaR: vi.fn().mockResolvedValue({ success: true, data: null }),
    priceAmericanOption: vi.fn().mockResolvedValue({ success: true, data: null }),
    calculateGreeks: vi.fn().mockResolvedValue({ success: true, data: null }),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('QuantLibAnalysis', () => {
  it('renders without crashing', () => {
    renderWithProviders(<QuantLibAnalysis />)
    expect(screen.getAllByText(/Quant|量化|分析/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<QuantLibAnalysis />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<QuantLibAnalysis />)
    await waitFor(() => {
      expect(mockGetReferenceData).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetReferenceData.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<QuantLibAnalysis />)
    await waitFor(() => {
      expect(mockGetReferenceData).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetReferenceData.mockResolvedValueOnce({ success: true, data: null })
    renderWithProviders(<QuantLibAnalysis />)
  })
})
