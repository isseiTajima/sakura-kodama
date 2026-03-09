<script>
  import { onMount, createEventDispatcher } from 'svelte'
  import { defaultConfig, loadConfig, saveConfig } from '../lib/wails'

  export let onClose = () => {}
  const dispatch = createEventDispatcher()

  let cfg = { ...defaultConfig }

  onMount(async () => {
    cfg = await loadConfig()
  })

  async function save() {
    await saveConfig(cfg)
    dispatch('saved', cfg)
    onClose()
  }
</script>

<div class="settings" onclick={(e) => e.stopPropagation()} onmousedown={(e) => e.stopPropagation()}>
  <div class="header">
    <h3>設定</h3>
    <button class="close-x" onclick={onClose}>×</button>
  </div>

  <div class="scroll-area">
    <label>
      あなたの呼び名
      <input type="text" bind:value={cfg.user_name} placeholder="開発者" />
    </label>

    <label>
      サクラの口調
      <select bind:value={cfg.tone}>
        <option value="genki">元気</option>
        <option value="calm">落ち着き</option>
        <option value="polite">丁寧</option>
        <option value="tsundere">ツンデレ</option>
      </select>
    </label>

    <label>
      言語 (Language)
      <select bind:value={cfg.language}>
        <option value="ja">日本語</option>
        <option value="en">English</option>
      </select>
    </label>

    <label>
      お話の頻度
      <select bind:value={cfg.speech_frequency}>
        <option value={1}>控えめ</option>
        <option value={2}>ふつう</option>
        <option value={3}>お喋り</option>
      </select>
    </label>

    <div class="section-title">脳 (LLM) 設定</div>
    
    <label>
      バックエンド
      <select bind:value={cfg.llm_backend}>
        <option value="ollama">Ollama (ローカル)</option>
        <option value="claude">Claude (API)</option>
        <option value="router">自動フォールバック</option>
      </select>
    </label>

    {#if cfg.llm_backend === 'ollama' || cfg.llm_backend === 'router'}
      <label>
        Ollama Endpoint
        <input type="text" bind:value={cfg.ollama_endpoint} />
      </label>
      <label>
        ローカルモデル
        <input type="text" bind:value={cfg.model} placeholder="gemma3:4b" />
      </label>
    {/if}

    {#if cfg.llm_backend === 'claude' || cfg.llm_backend === 'router'}
      <label>
        Claude API Key
        <input type="password" bind:value={cfg.anthropic_api_key} placeholder="sk-ant-..." />
      </label>
    {/if}

    {#if cfg.llm_backend === 'router'}
      <label>
        Gemini API Key
        <input type="password" bind:value={cfg.gemini_api_key} placeholder="AIza..." />
      </label>
    {/if}

    <div class="section-title">表示設定</div>

    <label>
      表示位置
      <select bind:value={cfg.window_position}>
        <option value="top-right">右上</option>
        <option value="bottom-right">右下</option>
      </select>
    </label>

    <div class="checkbox-group">
      <label>
        <input type="checkbox" bind:checked={cfg.monologue} />
        独り言
      </label>
      <label>
        <input type="checkbox" bind:checked={cfg.always_on_top} />
        最前面
      </label>
      <label>
        <input type="checkbox" bind:checked={cfg.mute} />
        ミュート
      </label>
      <label title="ログイン時に自動でサクラが起動します">
        <input type="checkbox" bind:checked={cfg.auto_start} />
        自動起動
      </label>
    </div>

    <div class="row">
      <label style="flex: 1;">
        大きさ ({Math.round(cfg.scale * 100)}%)
        <input type="range" min="0.8" max="2.0" step="0.1" bind:value={cfg.scale} />
      </label>
      <label style="flex: 1;">
        不透明度 ({Math.round(cfg.independent_window_opacity * 100)}%)
        <input type="range" min="0.05" max="1" step="0.05" bind:value={cfg.independent_window_opacity} />
      </label>
    </div>
  </div>

  <div class="buttons">
    <button class="save-btn" onclick={save}>保存して適用</button>
  </div>
</div>

<style>
  .settings {
    background: rgba(255, 255, 255, 0.98);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(0, 0, 0, 0.2);
    border-radius: 10px;
    padding: 8px 10px;
    font-size: 10px;
    width: 210px;
    max-height: 260px; /* 300pxのウィンドウ内に収める */
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.2);
    color: #333;
  }

  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 6px;
    border-bottom: 1px solid #eee;
    padding-bottom: 2px;
  }

  .close-x {
    background: none;
    border: none;
    font-size: 16px;
    cursor: pointer;
    color: #999;
    padding: 0 2px;
  }

  .scroll-area {
    overflow-y: auto;
    flex: 1;
    padding-right: 4px;
  }

  h3 {
    margin: 0;
    font-size: 11px;
    font-weight: bold;
  }

  .section-title {
    font-weight: bold;
    margin: 8px 0 4px;
    color: #e91e63;
    font-size: 9px;
    text-transform: uppercase;
  }

  label {
    display: flex;
    flex-direction: column;
    margin-bottom: 6px;
    gap: 1px;
  }

  .row {
    display: flex;
    gap: 6px;
  }

  .checkbox-group {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 2px;
    margin-bottom: 6px;
  }

  input[type='text'],
  input[type='password'],
  select {
    padding: 3px;
    font-size: 10px;
    border: 1px solid #ddd;
    border-radius: 4px;
    background: white;
  }

  label:has(input[type='checkbox']) {
    flex-direction: row;
    align-items: center;
    gap: 4px;
    margin-bottom: 2px;
    cursor: pointer;
  }

  .buttons {
    margin-top: 8px;
  }

  .save-btn {
    width: 100%;
    padding: 5px;
    background: #e91e63;
    color: white;
    border: none;
    border-radius: 6px;
    font-weight: bold;
    cursor: pointer;
    font-size: 10px;
  }

  .save-btn:hover {
    background: #d81b60;
  }
</style>
