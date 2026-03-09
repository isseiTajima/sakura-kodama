<script lang="ts">
  import { onDestroy } from 'svelte'
  import { InstallOllama } from '../lib/wails'

  let { 
    message = { id: 0, text: '' }, 
    scale = 1, 
    usingFallback = false,
    visible = $bindable(false),
    position = 'top-right'
  } = $props()

  let timer: ReturnType<typeof setTimeout> | null = null
  const maxLength = 200

  let showInstallButton = $state(false)
  let displayText = $state('')

  $effect(() => {
    const text = message.text

    if (!text) {
      visible = false
      return
    }

    if (showInstallButton && !text.includes('[INSTALL_OLLAMA]')) {
      return
    }

    clearTimeout(timer ?? undefined)
    
    if (text.includes('[INSTALL_OLLAMA]')) {
      showInstallButton = true
      displayText = text.replace('[INSTALL_OLLAMA]', '')
    } else {
      showInstallButton = false
      displayText = text
    }

    visible = true
    if (!showInstallButton) {
      const duration = Math.min(15000, 8000 + (displayText.length * 150))
      timer = setTimeout(() => { visible = false }, duration)
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
  <div class="balloon" 
    class:fallback={usingFallback} 
    style="
      transform: scale({scale});
      transform-origin: right center;
    "
  >
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
    position: relative;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(8px);
    -webkit-backdrop-filter: blur(8px);
    border: 1.2px solid rgba(0, 0, 0, 0.1);
    border-radius: 20px;
    padding: 12px 16px;
    width: auto;
    max-width: 250px;
    min-height: 40px;
    height: auto;
    word-break: normal;
    overflow-wrap: break-word;
    font-size: 11px;
    line-height: 1.4;
    color: #333;
    z-index: 10;
    display: flex;
    align-items: center;
    justify-content: flex-start;
    text-align: left; /* 左揃え */
    box-sizing: border-box;
    box-shadow: 0 4px 15px rgba(0, 0, 0, 0.08);
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
    align-self: flex-start;
  }

  /* しっぽスタイル: 右下からキャラに向かって出る */
  .balloon::before, .balloon::after {
    content: '';
    position: absolute;
    bottom: 8px;
    right: -6px;
    transform: rotate(35deg);
  }

  .balloon::before {
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
    border-left: 12px solid rgba(0, 0, 0, 0.1);
  }

  .balloon::after {
    right: -5px;
    border-top: 4px solid transparent;
    border-bottom: 4px solid transparent;
    border-left: 11px solid rgba(255, 255, 255, 0.95);
  }

  p {
    margin: 0;
    font-weight: 500;
  }

  .balloon.fallback {
    background: rgba(255, 248, 225, 0.95);
    border: 1.2px solid rgba(255, 193, 7, 0.3);
  }

  .fallback-label {
    font-size: 12px;
  }
</style>
