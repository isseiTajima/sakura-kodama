import { beforeEach, afterEach, describe, expect, it, vi } from 'vitest'
import { mount, unmount } from '../node_modules/svelte/src/internal/client/render.js'
import { settled } from '../node_modules/svelte/src/internal/client/runtime.js'
import Balloon from '@/components/Balloon.svelte'

vi.mock('svelte', () => import('../node_modules/svelte/src/index-client.js'))

async function flush() {
  await settled()
}

describe('Balloon', () => {
  beforeEach(() => {
    document.body.innerHTML = ''
    vi.useFakeTimers()
  })

  afterEach(async () => {
    document.body.innerHTML = ''
    vi.useRealTimers()
  })

  describe('初期状態', () => {
    it('テキストが空のときバルーンを表示しない', async () => {
      // Given: テキストが空のメッセージ
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 0, text: '' }, scale: 1 },
      })
      await flush()

      // Then: .balloon 要素が存在しない
      expect(document.querySelector('.balloon')).toBeNull()

      await unmount(balloon)
    })
  })

  describe('表示制御', () => {
    it('テキストがあるときバルーンを表示する', async () => {
      // Given: テキストありのメッセージ
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'こんにちは' }, scale: 1 },
      })
      await flush()

      // Then: .balloon 要素が存在する
      expect(document.querySelector('.balloon')).not.toBeNull()

      await unmount(balloon)
    })

    it('8000ms後にバルーンを非表示にする', async () => {
      // Given: テキストありでバルーンが表示されている
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'こんにちは' }, scale: 1 },
      })
      await flush()
      expect(document.querySelector('.balloon')).not.toBeNull()

      // When: 8000ms 経過（5文字 → max(8000, 5*200) = 8000ms）
      vi.advanceTimersByTime(8000)
      await flush()

      // Then: バルーンが非表示になる
      expect(document.querySelector('.balloon')).toBeNull()

      await unmount(balloon)
    })

    it('8000ms未満ではバルーンを表示し続ける', async () => {
      // Given: テキストありでバルーンが表示されている
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'こんにちは' }, scale: 1 },
      })
      await flush()

      // When: 7999ms 経過（タイムアウト直前）
      vi.advanceTimersByTime(7999)
      await flush()

      // Then: まだ表示されている
      expect(document.querySelector('.balloon')).not.toBeNull()

      await unmount(balloon)
    })
  })

  describe('テキスト表示', () => {
    it('メッセージのテキストをそのまま表示する', async () => {
      // Given: 短いテキスト
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'こんにちは' }, scale: 1 },
      })
      await flush()

      // Then: テキストが表示されている
      expect(document.querySelector('.balloon p')?.textContent).toBe('こんにちは')

      await unmount(balloon)
    })
  })

  describe('テキストトリミング', () => {
    it('40文字以下のテキストはそのまま表示する', async () => {
      // Given: ちょうど40文字のテキスト（境界値）
      const text = 'a'.repeat(40)
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text }, scale: 1 },
      })
      await flush()

      // Then: 切り詰めなし
      expect(document.querySelector('.balloon p')?.textContent).toBe(text)

      await unmount(balloon)
    })

    it('41文字以上のテキストは37文字+...に切り詰める', async () => {
      // Given: 41文字のテキスト（境界値+1）
      const text = 'a'.repeat(41)
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text }, scale: 1 },
      })
      await flush()

      // Then: 37文字 + '...' に切り詰められる（maxLength=40, slice(0, 40-3)=37）
      expect(document.querySelector('.balloon p')?.textContent).toBe('a'.repeat(37) + '...')

      await unmount(balloon)
    })

    it('100文字のテキストは37文字+...に切り詰める', async () => {
      // Given: 100文字のテキスト
      const text = 'x'.repeat(100)
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text }, scale: 1 },
      })
      await flush()

      // Then: 37文字 + '...' に切り詰められる
      expect(document.querySelector('.balloon p')?.textContent).toBe('x'.repeat(37) + '...')

      await unmount(balloon)
    })

    it('絵文字を含む長いテキストを文字単位で正しく切り詰める', async () => {
      // Given: 絵文字41個（Array.from で正しく1文字ずつカウントされる）
      const text = '😀'.repeat(41)
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text }, scale: 1 },
      })
      await flush()

      // Then: 絵文字37個 + '...'（バイト数でなく文字数でカウント）
      expect(document.querySelector('.balloon p')?.textContent).toBe('😀'.repeat(37) + '...')

      await unmount(balloon)
    })
  })

  describe('usingFallback フラグ表示', () => {
    it('usingFallback=true で🔄アイコンがDOMに存在する', async () => {
      // Given: usingFallback=true でテキストありのメッセージ
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'フォールバック中' }, scale: 1, usingFallback: true },
      })
      await flush()

      // Then: .fallback-label要素が存在し、🔄アイコンを含む
      const fallbackLabel = document.querySelector('.fallback-label')
      expect(fallbackLabel).not.toBeNull()
      expect(fallbackLabel?.textContent).toBe('🔄')

      await unmount(balloon)
    })

    it('usingFallback=false でアイコンが非表示になる', async () => {
      // Given: usingFallback=false でテキストありのメッセージ
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'テスト' }, scale: 1, usingFallback: false },
      })
      await flush()

      // Then: .fallback-label要素が存在しない
      expect(document.querySelector('.fallback-label')).toBeNull()

      await unmount(balloon)
    })

    it('.fallback クラスがusingFallback=true で条件付き適用される', async () => {
      // Given: usingFallback=true でバルーンが表示されている
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'フォールバック' }, scale: 1, usingFallback: true },
      })
      await flush()

      // Then: .balloon要素に.fallbackクラスが適用されている
      const balloonElement = document.querySelector('.balloon')
      expect(balloonElement?.classList.contains('fallback')).toBe(true)

      await unmount(balloon)
    })

    it('.fallback クラスがusingFallback=false で未適用になる', async () => {
      // Given: usingFallback=false でバルーンが表示されている
      const balloon = mount(Balloon, {
        target: document.body,
        props: { message: { id: 1, text: 'テスト' }, scale: 1, usingFallback: false },
      })
      await flush()

      // Then: .balloon要素に.fallbackクラスが適用されていない
      const balloonElement = document.querySelector('.balloon')
      expect(balloonElement?.classList.contains('fallback')).toBe(false)

      await unmount(balloon)
    })
  })
})
