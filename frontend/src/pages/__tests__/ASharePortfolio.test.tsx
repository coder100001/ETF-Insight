import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import ASharePortfolio from '../ASharePortfolio'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetDefaultPortfolio = vi.fn().mockResolvedValue({ success: true, data: null })
const mockGetETFPrices = vi.fn().mockResolvedValue({ success: true, data: [] })

vi.mock('../../services/api', async (importOriginal) => {
  const actual = (await importOriginal()) as Record<string, unknown>
  return {
    ...actual,
    aShareAPI: {
      getDefaultPortfolio: (...args: unknown[]) => mockGetDefaultPortfolio(...args),
      getETFPrices: (...args: unknown[]) => mockGetETFPrices(...args),
    },
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('ASharePortfolio', () => {
  it('renders without crashing', () => {
    renderWithProviders(<ASharePortfolio />)
    expect(screen.getAllByText(/A股|组合|Portfolio|ashare/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<ASharePortfolio />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    mockGetDefaultPortfolio.mockResolvedValueOnce(null)
    renderWithProviders(<ASharePortfolio />)
  })
})
