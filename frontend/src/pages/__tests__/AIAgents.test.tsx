import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import AIAgents from '../AIAgents'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockDiscover = vi.fn().mockResolvedValue({ success: true, data: [] })

vi.mock('../../services/api', () => ({
  agentAPI: {
    discover: (...args: unknown[]) => mockDiscover(...args),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('AIAgents', () => {
  it('renders without crashing', () => {
    renderWithProviders(<AIAgents />)
    expect(screen.getAllByText(/AI|Agent|智能体/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<AIAgents />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles empty data gracefully', async () => {
    renderWithProviders(<AIAgents />)
    await waitFor(() => {
      expect(mockDiscover).toHaveBeenCalled()
    })
  })

  it('handles error state', async () => {
    mockDiscover.mockRejectedValueOnce(new Error('API Error'))
    renderWithProviders(<AIAgents />)
    await waitFor(() => {
      expect(mockDiscover).toHaveBeenCalled()
    })
  })

  it('handles boundary conditions without crashing', () => {
    mockDiscover.mockResolvedValueOnce(null)
    renderWithProviders(<AIAgents />)
  })
})
