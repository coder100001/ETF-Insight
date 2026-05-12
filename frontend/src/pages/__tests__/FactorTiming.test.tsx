import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import FactorTiming from '../FactorTiming'

beforeEach(() => {
  // Mock ResizeObserver
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }

  // Mock window and related APIs if not available
  if (typeof globalThis.window === 'undefined') {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  (globalThis as any).window = {
      matchMedia: vi.fn().mockImplementation(() => ({
        matches: false,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
      scrollTo: vi.fn(),
    };
  }
})

const mockCalculateSignal = vi.fn().mockResolvedValue({
  success: true,
  data: {
    id: 1,
    factor_name: 'Mkt-RF',
    signal_date: '2026-05-08',
    ma_slope_60: 0.0012,
    z_score: 1.5,
    percentile: 75.0,
    signal_strength: 'weak_positive',
    expected_return: 0.08,
    confidence: 65.0,
    signal_score: 0.6,
  },
})

const mockGetSignalHistory = vi.fn().mockResolvedValue({
  success: true,
  data: [],
})

vi.mock('../../services/api', () => ({
  factorTimingAPI: {
    calculateSignal: (...args: unknown[]) => mockCalculateSignal(...args),
    getSignalHistory: (...args: unknown[]) => mockGetSignalHistory(...args),
    generateView: vi.fn().mockResolvedValue({ success: true, data: null }),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider>
      <AntdApp>
        <BrowserRouter>{ui}</BrowserRouter>
      </AntdApp>
    </ConfigProvider>
  )
}

describe('FactorTiming', () => {
  it('renders without crashing', () => {
    renderWithProviders(<FactorTiming />)
    expect(screen.getAllByText(/因子择时/i).length).toBeGreaterThan(0)
  })

  it('has a calculate button', () => {
    renderWithProviders(<FactorTiming />)
    expect(screen.getByText(/计算信号/i)).toBeTruthy()
  })

  it('calls calculateSignal API on button click', async () => {
    renderWithProviders(<FactorTiming />)
    const button = screen.getByText(/计算信号/i)
    fireEvent.click(button)
    await waitFor(() => {
      expect(mockCalculateSignal).toHaveBeenCalledWith('Mkt-RF', 60)
    })
  })
})
