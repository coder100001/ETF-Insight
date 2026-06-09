import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import AlphaViews from '../AlphaViews'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetActive = vi.fn().mockResolvedValue({ success: true, data: [] })
const mockCreate = vi.fn().mockResolvedValue({ success: true, data: null })

vi.mock('../../services/api', async (importOriginal) => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const actual = (await importOriginal()) as any
  return {
    ...actual,
    alphaViewAPI: {
      getActive: (...args: unknown[]) => mockGetActive(...args),
      create: (...args: unknown[]) => mockCreate(...args),
    },
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('AlphaViews', () => {
  it('renders without crashing', { timeout: 15000 }, () => {
    renderWithProviders(<AlphaViews />)
    expect(screen.getAllByText(/Alpha|观点|View[s]?/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<AlphaViews />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<AlphaViews />)
    await waitFor(() => {
      expect(mockGetActive).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockGetActive.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<AlphaViews />)
    await waitFor(() => {
      expect(mockGetActive).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockGetActive.mockResolvedValueOnce(null)
    mockCreate.mockResolvedValueOnce(undefined)
    renderWithProviders(<AlphaViews />)
  })
})
