import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { mount, unmount } from '../node_modules/svelte/src/internal/client/render.js'
import { settled } from '../node_modules/svelte/src/internal/client/runtime.js'
import App from '@/App.svelte'
import Settings from '@/components/Settings.svelte'
import Chara from '@/components/Chara.svelte'
import { defaultConfig, loadConfig, saveConfig } from '@/lib/wails'

vi.mock('svelte', () => import('../node_modules/svelte/src/index-client.js'))

class MockWebSocket {
  readyState = 1
  onopen: ((event: Event) => void) | null = null
  onmessage: ((event: MessageEvent) => void) | null = null
  onclose: ((event: CloseEvent) => void) | null = null
  onerror: ((event: Event) => void) | null = null

  constructor(public url: string) {}

  send() {}

  close() {
    this.readyState = 3
    this.onclose?.(new Event('close') as CloseEvent)
  }
}

async function flushSvelteJobs() {
  await settled()
}

describe('Wails fallbacks', () => {
  let originalWebSocket: typeof WebSocket | undefined

  beforeEach(async () => {
    document.body.innerHTML = ''
    window.localStorage?.clear()
    originalWebSocket = globalThis.WebSocket
    globalThis.WebSocket = MockWebSocket as unknown as typeof WebSocket
    await saveConfig({ ...defaultConfig })
  })

  afterEach(() => {
    document.body.innerHTML = ''
    vi.restoreAllMocks()
    if (originalWebSocket) {
      globalThis.WebSocket = originalWebSocket
    } else {
      // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
      delete (globalThis as Record<string, unknown>).WebSocket
    }
  })

  it('App does not log config load failures when the runtime is missing', async () => {
    const errorSpy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const app = mount(App, { target: document.body })
    await flushSvelteJobs()

    const loggedLoadFailure = errorSpy.mock.calls.some(
      call => call[0] === 'load config failed'
    )
    expect(loggedLoadFailure).toBe(false)

    await unmount(app)
  })

  it('Settings mount does not trigger unhandled rejections without the runtime', async () => {
    const rejections: unknown[] = []
    const handler = (event: PromiseRejectionEvent) => {
      rejections.push(event.reason)
    }
    window.addEventListener('unhandledrejection', handler)

    const settings = mount(Settings, { target: document.body })
    await flushSvelteJobs()

    expect(rejections).toHaveLength(0)

    await unmount(settings)
    window.removeEventListener('unhandledrejection', handler)
  })

  it('Chara click handler does not throw when the runtime is missing', async () => {
    const chara = mount(Chara, {
      target: document.body,
      props: {
        clickThrough: false,
        scale: 1,
        state: 'Idle',
        mood: 'Calm',
      },
    })

    const button = document.querySelector<HTMLButtonElement>('.chara-button')
    expect(button).not.toBeNull()

    const triggerClick = () => {
      button?.dispatchEvent(new window.MouseEvent('click', { bubbles: true }))
    }

    expect(triggerClick).not.toThrow()

    await unmount(chara)
  })

  it('retains the last saved config when the runtime is missing', async () => {
    const initial = await loadConfig()
    expect(initial.name).toBe(defaultConfig.name)

    const updated = { ...initial, name: 'テスト', scale: 1.25 }
    await saveConfig(updated)

    const reloaded = await loadConfig()
    expect(reloaded).toMatchObject({ name: 'テスト', scale: 1.25 })
  })
})

describe('App WebSocket using_fallback handling', () => {
  let originalWebSocket: typeof WebSocket | undefined
  let socketInstance: MockWebSocket | undefined

  beforeEach(async () => {
    document.body.innerHTML = ''
    window.localStorage?.clear()
    originalWebSocket = globalThis.WebSocket

    // MockWebSocket をグローバルに設定
    globalThis.WebSocket = MockWebSocket as unknown as typeof WebSocket
    await saveConfig({ ...defaultConfig })
  })

  afterEach(() => {
    document.body.innerHTML = ''
    vi.restoreAllMocks()
    socketInstance = undefined
    if (originalWebSocket) {
      globalThis.WebSocket = originalWebSocket
    } else {
      // eslint-disable-next-line @typescript-eslint/no-dynamic-delete
      delete (globalThis as Record<string, unknown>).WebSocket
    }
  })

  it('using_fallback: true が WebSocket イベントで受け取られた場合、.fallback クラスが適用される', async () => {
    // MockWebSocket のコンストラクタをオーバーライドしてインスタンスを保持
    const originalMockWebSocket = MockWebSocket
    class TrackedMockWebSocket extends MockWebSocket {
      constructor(url: string) {
        super(url)
        socketInstance = this
      }
    }
    globalThis.WebSocket = TrackedMockWebSocket as unknown as typeof WebSocket

    const app = mount(App, { target: document.body })
    await flushSvelteJobs()

    // WebSocket メッセージをシミュレート
    if (socketInstance?.onmessage) {
      const event = new MessageEvent('message', {
        data: JSON.stringify({
          state: 'Speaking',
          mood: 'Happy',
          speech: 'こんにちは',
          using_fallback: true,
        }),
      })
      socketInstance.onmessage(event)
      await flushSvelteJobs()

      // .fallback クラスが適用されていることを確認
      const balloonElement = document.querySelector('.balloon.fallback')
      expect(balloonElement).not.toBeNull()
    }

    globalThis.WebSocket = originalMockWebSocket as unknown as typeof WebSocket
    await unmount(app)
  })

  it('using_fallback フィールドが欠落した場合、false にデフォルト設定される', async () => {
    const originalMockWebSocket = MockWebSocket
    class TrackedMockWebSocket extends MockWebSocket {
      constructor(url: string) {
        super(url)
        socketInstance = this
      }
    }
    globalThis.WebSocket = TrackedMockWebSocket as unknown as typeof WebSocket

    const app = mount(App, { target: document.body })
    await flushSvelteJobs()

    // using_fallback フィールドなしのメッセージを送信
    if (socketInstance?.onmessage) {
      const event = new MessageEvent('message', {
        data: JSON.stringify({
          state: 'Idle',
          mood: 'Calm',
          speech: 'テスト',
        }),
      })
      socketInstance.onmessage(event)
      await flushSvelteJobs()

      // .fallback クラスが存在しないことを確認（デフォルト false）
      const fallbackBalloons = document.querySelectorAll('.balloon.fallback')
      expect(fallbackBalloons.length).toBe(0)
    }

    globalThis.WebSocket = originalMockWebSocket as unknown as typeof WebSocket
    await unmount(app)
  })

  it('Balloon コンポーネントが usingFallback=true で 🔄 アイコンを表示する', async () => {
    const originalMockWebSocket = MockWebSocket
    class TrackedMockWebSocket extends MockWebSocket {
      constructor(url: string) {
        super(url)
        socketInstance = this
      }
    }
    globalThis.WebSocket = TrackedMockWebSocket as unknown as typeof WebSocket

    const app = mount(App, { target: document.body })
    await flushSvelteJobs()

    // using_fallback: true でメッセージを送信
    if (socketInstance?.onmessage) {
      const event = new MessageEvent('message', {
        data: JSON.stringify({
          state: 'Speaking',
          mood: 'Focused',
          speech: 'フォールバック中です',
          using_fallback: true,
        }),
      })
      socketInstance.onmessage(event)
      await flushSvelteJobs()

      // 🔄 アイコンが表示されていることを確認
      const fallbackLabel = document.querySelector('.fallback-label')
      expect(fallbackLabel?.textContent).toBe('🔄')

      // .fallback クラスが適用されていることを確認
      const balloon = document.querySelector('.balloon.fallback')
      expect(balloon).not.toBeNull()
    }

    globalThis.WebSocket = originalMockWebSocket as unknown as typeof WebSocket
    await unmount(app)
  })
})
