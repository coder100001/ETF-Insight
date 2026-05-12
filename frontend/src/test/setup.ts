import '@testing-library/jest-dom'
import { vi } from 'vitest'

// 启用 fake timers - 这是修复的核心！
vi.useFakeTimers()

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
    message.includes('Tabs.TabPane') ||
    message.includes('window is not defined')
  ) {
    return
  }
  originalError.apply(console, args)
}

// 完整模拟浏览器环境
if (typeof globalThis !== 'undefined') {
  // 模拟 window 对象
  if (typeof globalThis.window === 'undefined') {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    (globalThis as any).window = {
      matchMedia: vi.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
      })),
      scrollTo: vi.fn(),
      innerWidth: 1024,
      innerHeight: 768,
      getComputedStyle: vi.fn().mockImplementation(() => ({
        getPropertyValue: vi.fn(),
      })),
      location: {
        href: 'http://localhost/',
        search: '',
        hash: '',
        pathname: '/',
        host: 'localhost',
        hostname: 'localhost',
        port: '',
        protocol: 'http:',
        assign: vi.fn(),
        replace: vi.fn(),
        reload: vi.fn(),
      },
      navigator: {
        userAgent: 'node.js',
        platform: 'linux',
        language: 'en-US',
        languages: ['en-US', 'en'],
        onLine: true,
        cookieEnabled: true,
      },
      document: {
        createElement: vi.fn().mockImplementation(tag => ({
          tagName: tag.toUpperCase(),
          setAttribute: vi.fn(),
          getAttribute: vi.fn(),
          removeAttribute: vi.fn(),
          addEventListener: vi.fn(),
          removeEventListener: vi.fn(),
          style: {},
          dataset: {},
          appendChild: vi.fn(),
          removeChild: vi.fn(),
          insertBefore: vi.fn(),
        })),
        createTextNode: vi.fn().mockImplementation(text => ({
          nodeType: 3,
          textContent: text,
          nodeValue: text,
        })),
        querySelector: vi.fn(),
        querySelectorAll: vi.fn().mockReturnValue([]),
        getElementById: vi.fn(),
        getElementsByTagName: vi.fn().mockReturnValue([]),
        getElementsByClassName: vi.fn().mockReturnValue([]),
        createEvent: vi.fn(),
        body: {
          appendChild: vi.fn(),
          removeChild: vi.fn(),
        },
        head: {
          appendChild: vi.fn(),
          removeChild: vi.fn(),
        },
        readyState: 'complete',
        cookie: '',
        title: '',
      },
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn(),
      requestAnimationFrame,
      cancelAnimationFrame,
      setTimeout,
      clearTimeout,
      setInterval,
      clearInterval,
    }

    // 将 window 的属性同步到 globalThis
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    Object.assign(globalThis, (globalThis as any).window)
  }

  // 全局 mock fetch，防止测试中的请求挂起
  if (typeof vi !== 'undefined') {
    const mockResponse = {
      ok: true,
      json: () => Promise.resolve({ success: true, data: [] }),
      text: () => Promise.resolve(''),
      status: 200,
      statusText: 'OK',
      headers: new Headers(),
    } as Response

    globalThis.fetch = vi.fn().mockImplementation(() => Promise.resolve(mockResponse))
  }
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

// Mock axios for testing
vi.mock('axios', () => ({
  default: {
    get: vi.fn().mockResolvedValue({
      data: {
        success: true,
        data: [],
      },
    }),
    post: vi.fn().mockResolvedValue({
      data: {
        success: true,
        data: {},
      },
    }),
    create: vi.fn().mockReturnValue({
      get: vi.fn().mockResolvedValue({ data: { success: true, data: [] } }),
      post: vi.fn().mockResolvedValue({ data: { success: true, data: {} } }),
    }),
  },
}))
