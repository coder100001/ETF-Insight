import '@testing-library/jest-dom'
import { vi } from 'vitest'

declare global {
  var ResizeObserver: typeof ResizeObserver
  var IntersectionObserver: typeof IntersectionObserver
}

// 抑制 antd 废弃警告，避免 CI/CD 日志膨胀
const originalWarn = console.warn
console.warn = (...args: unknown[]) => {
  const message = args[0]?.toString() || ''
  if (
    message.includes('deprecated') ||
    message.includes('TabPane') ||
    message.includes('TabPane') ||
    message.includes('message is deprecated') ||
    message.includes('valueStyle is deprecated') ||
    message.includes('Please use')
  ) {
    return
  }
  originalWarn.apply(console, args)
}

// 抑制 React 最大更新深度错误（由 antd 废弃 API 触发）
const originalError = console.error
console.error = (...args: unknown[]) => {
  const message = args[0]?.toString() || ''
  if (
    message.includes('Maximum update depth exceeded') ||
    message.includes('Tabs.TabPane')
  ) {
    return
  }
  originalError.apply(console, args)
}

if (typeof window !== 'undefined') {
  Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: vi.fn().mockImplementation(query => ({
      matches: false,
      media: query,
      onchange: null,
      addListener: vi.fn(),
      removeListener: vi.fn(),
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
    })),
  })
}

;(globalThis as typeof globalThis & { ResizeObserver: typeof ResizeObserver }).ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

;(globalThis as typeof globalThis & { IntersectionObserver: typeof IntersectionObserver }).IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
  root: null,
  rootMargin: '',
  thresholds: [],
  takeRecords: vi.fn(),
}))
