<script lang="ts">
  import { onDestroy } from 'svelte'
  import { InstallOllama } from '../lib/wails'

  let { 
    message = { id: 0, text: '' }, 
    scale = 1, 
    usingFallback = false,
    visible = $bindable(false)
  } = $props()

  let timer: ReturnType<typeof setTimeout> | null = null
  const maxLength = 80 

  let showInstallButton = $state(false)
  let displayText = $state('')

  $effect(() => {
    const mid = message.id
    const text = message.text

    if (showInstallButton && text && !text.includes('[INSTALL_OLLAMA]')) {
      return
    }

    clearTimeout(timer ?? undefined)
    
    if (text) {
      if (text.includes('[INSTALL_OLLAMA]')) {
        showInstallButton = true
        displayText = text.replace('[INSTALL_OLLAMA]', '')
      } else {
        showInstallButton = false
        displayText = text
      }

      visible = true
      if (!showInstallButton) {
        const duration = Math.min(12000, 6000 + (displayText.length * 150))
        timer = setTimeout(() => { visible = false }, duration)
      }
    } else {
      visible = false
    }
  })

  const trimmed = $derived(trimSpeech(displayText ?? ''))

  function trimSpeech(value: string) {
    if (!value) return ''
    const chars = Array.from(value)
    if (chars.length <= maxLength) return value
    return chars.slice(0, maxLength - 3).join('') + '...'
  }

  async function handleInstall() {
    showInstallButton = false
    await InstallOllama()
  }

  onDestroy(() => { if (timer) clearTimeout(timer) })
</script>

{#if visible && trimmed}
  <div class="balloon" class:fallback={usingFallback} style="
    bottom: {Math.round(60 * scale)}px;
    left: {Math.round(105 * scale)}px; /* キャラの右肩付近に寄せる */
    transform: scale({scale});
    transform-origin: bottom left;
  ">
    <div class="content">
      <p class="balloon-text">{#if usingFallback}<span class="fallback-label">🔄</span>&nbsp;{/if}{trimmed}</p>
      
      {#if showInstallButton}
        <button class="install-btn" onclick={handleInstall}>
          今すぐインストール
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .balloon {
    position: absolute;
    background: rgba(255, 255, 255, 0.9);
    backdrop-filter: blur(4px);
    -webkit-backdrop-filter: blur(4px);
    border: 1.5px solid rgba(0, 0, 0, 0.15);
    border-radius: 12px;
    padding: 10px 12px;
    width: 150px; /* ウィンドウ端に収まるよう幅を少し制限 */
    min-height: 50px;
    height: auto;
    word-break: break-all;
    font-size: 11px;
    line-height: 1.4;
    color: #000;
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: center;
    text-align: center;
    box-sizing: border-box;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .content {
    display: flex;
    flex-direction: column;
    gap: 6px;
    width: 100%;
  }

  .install-btn {
    background: #4a90e2;
    color: white;
    border: none;
    border-radius: 6px;
    padding: 4px 8px;
    font-size: 10px;
    font-weight: bold;
    cursor: pointer;
    transition: background 0.2s;
  }

  /* 吹き出しのしっぽ（外枠） - 左側に配置 */
  .balloon::before {
    content: '';
    position: absolute;
    top: 50%;
    right: 100%;
    transform: translateY(-50%);
    border-top: 8px solid transparent;
    border-bottom: 8px solid transparent;
    border-right: 12px solid rgba(0, 0, 0, 0.15);
  }

  /* 吹き出しのしっぽ（中身） - 左側に配置 */
  .balloon::after {
    content: '';
    position: absolute;
    top: 50%;
    right: calc(100% - 1.5px);
    transform: translateY(-50%);
    border-top: 7px solid transparent;
    border-bottom: 7px solid transparent;
    border-right: 11px solid rgba(255, 255, 255, 0.9);
  }

  p {
    margin: 0;
    font-weight: 500;
  }

  .balloon.fallback {
    background: rgba(255, 193, 7, 0.15);
    border: 1.5px solid rgba(255, 193, 7, 0.4);
  }

  .fallback-label {
    font-size: 12px;
  }
</style>
