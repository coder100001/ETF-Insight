import { describe, it, expect, vi, beforeEach } from 'vitest'
import type { FactorTimingSignal } from '../../types/factor'

declare global {
  interface Window {
    fetch: typeof fetch
  }
}

describe('API Path Consistency', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('generateView should call correct API path matching backend route', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: {} }),
    })
    ;(globalThis as { fetch: typeof fetch }).fetch = mockFetch

    const { factorTimingAPI } = await import('../api')

    const mockSignal: Partial<FactorTimingSignal> = { id: 1, factor_name: 'Mkt-RF' }
    await factorTimingAPI.generateView(mockSignal as FactorTimingSignal, 'SPY')

    const expectedPath = '/alpha-views/generate-from-factor'
    const actualCall = mockFetch.mock.calls[0][0]

    expect(actualCall).toContain(expectedPath)
  })

  it('generateView should send correct parameters matching backend expectations', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: {} }),
    })
    ;(globalThis as { fetch: typeof fetch }).fetch = mockFetch

    const { factorTimingAPI } = await import('../api')

    // 模拟真实的 FactorTimingSignal 对象
    const signal: FactorTimingSignal = {
      id: 1,
      factor_name: 'Mkt-RF',
      signal_date: '2026-05-01',
      ma_slope_60: 0.05,
      z_score: 1.5,
      percentile: 75.0,
      signal_strength: 'weak_positive',
      signal_score: 0.65,
      expected_return: 0.08,
      confidence: 65.0,
      created_at: '2026-05-01T00:00:00Z',
    }

    await factorTimingAPI.generateView(signal, 'SPY')

    // 验证请求体
    const requestBody = JSON.parse(mockFetch.mock.calls[0][1].body)

    expect(requestBody).toHaveProperty('factor_name', 'Mkt-RF')
    expect(requestBody).toHaveProperty('asset_symbol', 'SPY')
    expect(requestBody).not.toHaveProperty('signal_id')
    expect(requestBody).not.toHaveProperty('target_asset')
  })

  it('getSignalHistory should support optional date range parameters', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: [] }),
    })
    ;(globalThis as { fetch: typeof fetch }).fetch = mockFetch

    const { factorTimingAPI } = await import('../api')

    // 测试带日期范围的调用
    await factorTimingAPI.getSignalHistory('Mkt-RF', '2026-01-01', '2026-05-01')

    const actualCall = mockFetch.mock.calls[0][0]

    // 验证 URL 包含日期参数
    expect(actualCall).toContain('start_date=2026-01-01')
    expect(actualCall).toContain('end_date=2026-05-01')
    expect(actualCall).toContain('/factor/timing/history/Mkt-RF')
  })

  it('getSignalHistory should work without date range parameters', async () => {
    const mockFetch = vi.fn().mockResolvedValue({
      ok: true,
      json: async () => ({ success: true, data: [] }),
    })
    ;(globalThis as { fetch: typeof fetch }).fetch = mockFetch

    const { factorTimingAPI } = await import('../api')

    // 测试不带日期范围的调用
    await factorTimingAPI.getSignalHistory('Mkt-RF')

    const actualCall = mockFetch.mock.calls[0][0]

    // 验证 URL 不包含日期参数
    expect(actualCall).not.toContain('start_date')
    expect(actualCall).not.toContain('end_date')
    expect(actualCall).toContain('/factor/timing/history/Mkt-RF')
  })
})
