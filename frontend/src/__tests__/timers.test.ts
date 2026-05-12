import { describe, it, expect, vi, beforeEach } from 'vitest'

describe('Timers Test', () => {
  beforeEach(() => {
    // 每次测试前重置所有 mocks
    vi.clearAllTimers()
  })

  it('should mock setTimeout correctly', () => {
    // ARRANGE
    const callback = vi.fn()

    // ACT
    setTimeout(callback, 1000)

    // ASSERT - 验证 callback 被正确地记录为待执行
    expect(callback).not.toHaveBeenCalled()

    // 验证我们可以控制定时器
    vi.runAllTimers()
    expect(callback).toHaveBeenCalledTimes(1)
  })

  it('should mock setInterval correctly', () => {
    // ARRANGE
    const callback = vi.fn()

    // ACT
    const intervalId = setInterval(callback, 100)

    // ASSERT
    expect(callback).not.toHaveBeenCalled()
    expect(intervalId).toBeDefined()

    // 清除它
    clearInterval(intervalId)
  })

  it('should mock requestAnimationFrame correctly', () => {
    // ARRANGE
    const callback = vi.fn()

    // ACT
    const frameId = requestAnimationFrame(callback)

    // ASSERT
    expect(callback).not.toHaveBeenCalled()
    expect(frameId).toBeDefined()

    // 清除它
    cancelAnimationFrame(frameId)
  })

  it('should execute requestAnimationFrame callback when advancing timers', () => {
    // ARRANGE
    const callback = vi.fn()

    // ACT
    requestAnimationFrame(callback)

    // ASSERT - 初始时不调用
    expect(callback).not.toHaveBeenCalled()

    // 验证我们可以通过控制定时器触发 callback
    vi.runAllTimers()
    expect(callback).toHaveBeenCalledTimes(1)
  })
})
