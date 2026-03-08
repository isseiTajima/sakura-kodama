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
  let heartbeatTimer: ReturnType<typeof setInterval> | null = null

  let cfg: AppConfig = $state({ ...defaultConfig })

  $effect(() => {
    const shouldCaptureMouse = showSettings || showOnboarding
    if (shouldCaptureMouse) {
      SetClickThrough(false)
    } else {
      SetClickThrough(true)
    }
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
    }
  }

  onMount(async () => {
    await refreshConfig()
    if ((window as any).runtime) {
      const r = (window as any).runtime
      r.EventsOn('monitor_event', updateUI)
      r.EventsOn('observation_event', updateUI)
      r.EventsOn('greeting_event', updateUI)
      r.EventsOn('click_event', updateUI)
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
    if (heartbeatTimer) clearInterval(heartbeatTimer)
  })
</script>

<main class:interactive={showSettings || showOnboarding}>
  <div class="chara-container">
    <!-- キャラクターを右から200pxの位置に配置 -->
    <div class="chara-anchor">
      <Chara 
        status={appStatus} 
        mood={appMood} 
        scale={cfg.scale} 
        clickThrough={true}
        isTalking={isTalking}
      />
      <!-- メッセージをキャラの右（右端付近）に配置 -->
      <Balloon bind:visible={isTalking} message={speechMessage} scale={cfg.scale} {usingFallback} />
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
    <div class="modal-backdrop settings-backdrop" onclick={(e) => { if (e.target === e.currentTarget) closeSettings() }}>
      <div class="settings-content-wrapper" onclick={(e) => e.stopPropagation()}>
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
    pointer-events: none !important;
  }

  main {
    width: 100vw;
    height: 100vh;
    background: transparent !important;
    pointer-events: none !important;
  }

  main.interactive {
    pointer-events: auto !important;
  }
  main.interactive :global(*) {
    pointer-events: auto !important;
  }

  .chara-container {
    position: absolute;
    bottom: 0; 
    right: 0;   
    width: 100%;
    height: 100%;
    display: flex;
    align-items: flex-end;
    justify-content: flex-end; /* 右寄せ */
    background: transparent;
    pointer-events: none !important; 
  }

  .chara-anchor {
    position: relative;
    margin-right: 200px; /* 右端からキャラまでの距離（メッセージ用スペース） */
    margin-bottom: 20px;
    width: 128px;
    height: 128px;
    pointer-events: none !important;
  }

  .chara-container :global(*) {
    pointer-events: none !important;
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
    pointer-events: auto !important;
  }

  .settings-backdrop {
    align-items: flex-start;
    padding-top: 20px;
  }

  .onboarding-backdrop {
    align-items: flex-end;
    justify-content: flex-end;
    padding: 0 20px 20px 0;
    background: transparent;
  }
</style>
