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
    message.includes('Please use') ||
    message.includes('Not implemented: Window\'s getComputedStyle() method: with pseudo-elements')
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
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const globalAny = globalThis as any;

  // jsdom 缺失的核心 window 属性
  const essentialWindowMock = {
    matchMedia: vi.fn().mockImplementation((query: string) => ({
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
    // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unused-vars
    getComputedStyle: vi.fn().mockImplementation((_element: any, _pseudoElement?: string) => ({
      getPropertyValue: vi.fn(),
    })),
    requestAnimationFrame,
    cancelAnimationFrame,
  };

  // 完整 window mock（仅在 window 不存在时使用）
  const fullWindowMock = {
    ...essentialWindowMock,
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
      createElement: vi.fn().mockImplementation((tag: string) => ({
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
      createTextNode: vi.fn().mockImplementation((text: string) => ({
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
    setTimeout,
    clearTimeout,
    setInterval,
    clearInterval,
  };

  if (!globalAny.window) {
    // window 不存在（非 jsdom 环境），使用完整 mock
    globalAny.window = fullWindowMock;
  } else {
    // window 已存在（jsdom 环境），仅合并 jsdom 缺失的属性
    Object.assign(globalAny.window, essentialWindowMock);
  }

  // 同步到 globalThis (只复制 globalThis 上不存在的属性，避免覆盖 crypto 等只读 getter)
  Object.keys(globalAny.window).forEach(key => {
    if (!(key in globalAny)) {
      try {
        globalAny[key] = globalAny.window[key];
      } catch {
        // 跳过只读属性
      }
    }
  });

  // 全局 mock fetch，防止测试中的请求挂起
  if (typeof vi !== 'undefined') {
    const mockResponse = {
      ok: true,
      json: () => Promise.resolve({ success: true, data: [] }),
      text: () => Promise.resolve(''),
      status: 200,
      statusText: 'OK',
      headers: new Headers(),
    } as Response;

    globalAny.fetch = vi.fn().mockImplementation(() => Promise.resolve(mockResponse));
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

// Global mock for services/api to include financialConfigAPI
vi.mock('../services/api', async (importOriginal) => {
  const actual = await importOriginal()
  return {
    ...(actual as Record<string, unknown>),
    financialConfigAPI: {
      get: vi.fn().mockResolvedValue({
        success: true,
        data: {
          risk_free_rate: 0.0435,
          trading_days_year: 252,
          default_currency: 'USD',
        },
      }),
      update: vi.fn().mockResolvedValue({ success: true, message: 'updated' }),
    },
  }
})
