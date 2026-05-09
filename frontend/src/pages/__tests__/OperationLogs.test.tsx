import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import OperationLogs from '../OperationLogs'

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

const mockGetLogTypes = vi.fn().mockResolvedValue({ types: [] })
const mockGetActionTypes = vi.fn().mockResolvedValue({ types: [] })
const mockGetUsers = vi.fn().mockResolvedValue({ users: [] })
const mockGetLogs = vi.fn().mockResolvedValue({ data: [], total: 0, page: 1, pageSize: 20 })

vi.mock('../../services/operationLogsService', () => ({
  operationLogsAPI: {
    getLogTypes: (...args: unknown[]) => mockGetLogTypes(...args),
    getActionTypes: (...args: unknown[]) => mockGetActionTypes(...args),
    getUsers: (...args: unknown[]) => mockGetUsers(...args),
    getLogs: (...args: unknown[]) => mockGetLogs(...args),
    exportLogs: vi.fn().mockResolvedValue({}),
    getLogDetail: vi.fn().mockResolvedValue(null),
  },
}))

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('OperationLogs', () => {
  it('renders without crashing', () => {
    renderWithProviders(<OperationLogs />)
    expect(screen.getAllByText(/日志|Log|操作/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<OperationLogs />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles initial data loading', async () => {
    renderWithProviders(<OperationLogs />)
    await waitFor(() => {
      expect(mockGetLogTypes).toHaveBeenCalled()
    })
  })
})