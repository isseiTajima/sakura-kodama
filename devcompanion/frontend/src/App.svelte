<script lang="ts">
  import { onMount, onDestroy } from 'svelte'
  import Chara from './components/Chara.svelte'
  import Balloon from './components/Balloon.svelte'
  import Settings from './components/Settings.svelte'
  import Onboarding from './components/Onboarding.svelte'
  import {
    defaultConfig,
    loadConfig,
    type AppConfig,
    onCharaClick,
    DetectSetupStatus,
    ExpandForOnboarding,
    SetClickThrough,
    onOpenSettings,
  } from './lib/wails'

  let appStatus = $state('Idle')
  let appMood = $state('Calm')
  let speechMessage = $state({ id: 0, text: '' })
  let speechSeq = 0
  let usingFallback = $state(false)
  let isTalking = $state(false)
  let showSettings = $state(false)
  let showOnboarding = $state(false)
  let isHoveringSettings = $state(false)
  let socket: WebSocket | null = null
  let reconnectDelay = 1000
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null

  let cfg: AppConfig = $state({ ...defaultConfig })

  // OSレベルのクリック透過制御
  $effect(() => {
    const isModalOpen = showSettings || showOnboarding || isHoveringSettings
    const ghostMode = !isModalOpen
    SetClickThrough(ghostMode)
    document.body.dataset.ghostMode = String(ghostMode)
  })

  const refreshConfig = async () => {
    try {
      const loaded = await loadConfig()
      cfg = { ...cfg, ...loaded }
    } catch (err) {
      console.error('Failed to load config', err)
    }
  }

  const closeSettings = async () => {
    showSettings = false
    await refreshConfig()
  }

  function updateUI(e: any) {
    appStatus = e.state
    appMood = e.mood
    if (e.speech) {
      speechMessage = { id: ++speechSeq, text: e.speech }
      usingFallback = e.using_fallback
      if (e.profile) {
        cfg.name = e.profile.name
        cfg.tone = e.profile.tone
      }
    }
  }

  function connectWebSocket() {
    socket = new WebSocket('ws://localhost:34567/')
    socket.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        updateUI(data)
      } catch (err) {
        console.error('Failed to parse WS message', err)
      }
    }
    socket.onopen = () => {
      reconnectDelay = 1000
      heartbeatTimer = setInterval(() => {
        if (socket?.readyState === WebSocket.OPEN) {
          socket.send('ping')
        }
      }, 30000)
    }
    socket.onclose = scheduleReconnect
    socket.onerror = scheduleReconnect
  }

  function scheduleReconnect() {
    cleanupSocket()
    setTimeout(() => {
      reconnectDelay = Math.min(reconnectDelay * 1.5, 8000)
      connectWebSocket()
    }, reconnectDelay)
  }

  function cleanupSocket() {
    if (heartbeatTimer) {
      clearInterval(heartbeatTimer)
      heartbeatTimer = null
    }
    if (socket) {
      socket.onclose = null
      socket.onerror = null
      socket.close()
      socket = null
    }
  }

  const handleCharaClick = () => {
    if (showOnboarding) return
    onCharaClick()
  }

  onMount(async () => {
    await refreshConfig()

    if ((window as any).runtime) {
      const r = (window as any).runtime
      const eventHandler = (payload: any) => {
        updateUI(payload)
      }
      r.EventsOn('monitor_event', eventHandler)
      r.EventsOn('observation_event', eventHandler)
      r.EventsOn('greeting_event', eventHandler)
      r.EventsOn('click_event', eventHandler)
    } else {
      connectWebSocket()
    }

    onOpenSettings(() => {
      showSettings = true
    })

    const status = await DetectSetupStatus()
    if (status.is_first_run) {
      showOnboarding = true
      await ExpandForOnboarding()
    }
  })

  onDestroy(() => {
    cleanupSocket()
  })

  // 右側に固定するためのロジック
  const isTopSide = $derived(cfg.window_position?.startsWith('top'))
</script>

<main>
  <!-- 右側にキャラを固定し、左向きに反転。吹き出しはその左に密着。 -->
  <div class="chara-container" class:pos-top={isTopSide} class:pos-bottom={!isTopSide}>
    <div class="balloon-positioner">
      <Balloon bind:visible={isTalking} message={speechMessage} scale={cfg.scale} {usingFallback} position={cfg.window_position} />
    </div>
    
    <div class="chara-flip-wrapper" style="pointer-events: {(showSettings || showOnboarding) ? 'auto' : 'none'};">
      <Chara 
        status={appStatus} 
        mood={appMood} 
        scale={cfg.scale} 
        isTalking={isTalking}
        onClick={handleCharaClick}
      />
    </div>
  </div>

  {#if showOnboarding}
    <div class="modal-backdrop onboarding-backdrop">
      <Onboarding
        onClose={() => showOnboarding = false}
        oncompleted={refreshConfig}
        currentSpeech={speechMessage.text}
      />
    </div>
  {:else if showSettings}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div class="modal-backdrop settings-backdrop" onclick={(e) => { if (e.target === e.currentTarget) closeSettings() }}>
      <div class="settings-content-wrapper" onclick={(e) => e.stopPropagation()} onmousedown={(e) => e.stopPropagation()}>
        <Settings onClose={closeSettings} on:saved={refreshConfig} />
      </div>
    </div>
  {/if}
</main>

<style>
  :global(html), :global(body) {
    margin: 0;
    padding: 0;
    width: 100%;
    height: 100%;
    background: transparent !important;
    overflow: hidden;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
    pointer-events: none;
  }

  main {
    width: 100vw;
    height: 100vh;
    position: relative;
    background: transparent !important;
    pointer-events: none;
  }

  .chara-container {
    position: absolute;
    display: flex;
    pointer-events: none;
    width: 100%;
    height: 100%;
    box-sizing: border-box;
    padding: 0;
    justify-content: flex-end; /* 右端に寄せる */
    flex-direction: row; /* [Balloon][Chara] の順 */
  }

  .pos-top { align-items: flex-start; }
  .pos-bottom { align-items: flex-end; }

  .chara-flip-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 20;
    /* 右側配置のときは常に左を向かせる */
    transform: scaleX(-1);
  }

  .balloon-positioner {
    position: relative;
    pointer-events: none;
    z-index: 10;
    /* キャラの透明部分に重ねて密着させる（反転しているのでマイナス値を調整） */
    margin-right: -25px; 
    margin-top: 10px;
  }

  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
    background: rgba(0, 0, 0, 0.1);
    z-index: 1000;
    pointer-events: auto;
  }

  .onboarding-backdrop {
    background: transparent;
  }
</style>
