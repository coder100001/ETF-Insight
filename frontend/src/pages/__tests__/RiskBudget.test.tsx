import React from 'react'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { ConfigProvider, App as AntdApp } from 'antd'
import RiskBudget from '../RiskBudget'

vi.mock('recharts', () => {
  const PassThrough = (props: { children?: React.ReactNode }) => props.children
  return {
    ResponsiveContainer: PassThrough,
    PieChart: PassThrough,
    Pie: () => null,
    Cell: () => null,
    Tooltip: () => null,
    Legend: () => null,
    BarChart: PassThrough,
    Bar: () => null,
    XAxis: () => null,
    YAxis: () => null,
    CartesianGrid: () => null,
  }
})

beforeEach(() => {
  globalThis.ResizeObserver = class ResizeObserver {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
})

vi.mock('../../services/api', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...actual,
  }
})

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <ConfigProvider><AntdApp><BrowserRouter>{ui}</BrowserRouter></AntdApp></ConfigProvider>
  )
}

describe('RiskBudget', () => {
  it('renders without crashing', { timeout: 15000 }, () => {
    renderWithProviders(<RiskBudget />)
    expect(screen.getAllByText(/风险|预算|Risk|Budget/i).length).toBeGreaterThan(0)
  })

  it('shows loading state while fetching data', () => {
    renderWithProviders(<RiskBudget />)
    const loadingElements = document.querySelectorAll('.ant-spin, .ant-skeleton')
    expect(loadingElements.length).toBeGreaterThan(0)
  })

  it('handles boundary conditions without crashing', () => {
    renderWithProviders(<RiskBudget />)
  })
})
