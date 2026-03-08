<script lang="ts">
  import { onCharaClick } from '../lib/wails'
  
  // 素材のインポート
  import c1Open from '../assets/pixel-chara1-open.png'
  import c1Half from '../assets/pixel-chara1-half.png'
  import c1Close from '../assets/pixel-chara1-close.png'
  import c2Base from '../assets/pixel-chara2-base.png'
  import c2Happy from '../assets/pixel-chara2-happy.png'

  let {
    status = 'Idle',
    clickThrough = false,
    scale = 1,
    isTalking = false,
    onClick = undefined // デフォルトを undefined に変更
  } = $props();

  let currentImg = $state(c1Open);
  let frameIdx = $state(0);
  let timerId: any = null;

  // アニメーション定義
  const animations = {
    blink: [
      ...Array(30).fill(c1Open), // 約3秒開けておく
      c1Half,
      c1Close,
      c1Half
    ],
    happy: [c2Base, c2Happy],
    fail: [c1Close],
    thinking: [c1Open, c1Half, c1Close, c1Half]
  };
  $effect(() => {
    if (timerId) clearInterval(timerId);
    
    const s = String(status).toLowerCase();
    let activeAnim = animations.blink;
    let fps = 10; // デフォルトを10FPSに上げる

    if (s === 'success') {
      activeAnim = animations.happy;
      fps = 4;
    } else if (s === 'fail') {
      activeAnim = animations.fail;
      fps = 1;
    } else if (s === 'thinking' || s === 'running' || s === 'editing') {
      activeAnim = animations.thinking;
      fps = 3;
    }

    frameIdx = 0;
    currentImg = activeAnim[0];

    if (activeAnim.length > 1) {
      timerId = setInterval(() => {
        frameIdx = (frameIdx + 1) % activeAnim.length;
        currentImg = activeAnim[frameIdx];
      }, 1000 / fps);
    }

    return () => {
      if (timerId) clearInterval(timerId);
    };
  });

  const handleClick = (e: MouseEvent) => {
    // クリックイベントの伝播を防ぐ（念のため）
    e.stopPropagation();
    
    // 親から関数が渡されていれば実行
    if (typeof onClick === 'function') {
      onClick();
    } else {
      // 渡されていない場合はデフォルト動作
      onCharaClick();
    }
  }
</script>

<div
  class="chara-wrapper"
  style="
    pointer-events: {clickThrough ? 'none' : 'auto'}; 
    transform: scale({scale});
    opacity: {isTalking ? 1.0 : 0.4};
  "
>
  <button
    type="button"
    class="chara-button"
    onclick={handleClick} 
    style="pointer-events: {clickThrough ? 'none' : 'auto'};"
    disabled={clickThrough}
  >
    <img
      src={currentImg}
      alt="キャラクター"
      class="pixel-art"
    />
  </button>
</div>

<style>
  .chara-wrapper {
    transform-origin: bottom center;
    transition: opacity 0.5s ease-in-out;
  }

  @keyframes floating {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-5px); }
  }

  .chara-button {
    border: none;
    background: transparent;
    padding: 0;
    cursor: pointer;
    display: block;
    line-height: 0;
    /* クリック領域を確保 */
    width: 100%;
    height: 100%;
    animation: floating 3s ease-in-out infinite;
  }

  img.pixel-art {
    width: 128px; 
    height: auto;
    image-rendering: pixelated;
  }
</style>
