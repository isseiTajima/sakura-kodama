<script lang="ts">
  import c1Open from '../assets/pixel-chara1-open.png'
  import c1Half from '../assets/pixel-chara1-half.png'
  import c1Close from '../assets/pixel-chara1-close.png'
  
  import c2Happy from '../assets/pixel-chara2-happy.png'

  let { 
    status = 'Idle', 
    mood = 'Calm', 
    scale = 1,
    isTalking = false,
    onClick = () => {} 
  } = $props()

  // 目パチの状態管理
  let eyeState = $state(0) // 0:open, 1:half, 2:close
  let blinkTimer: ReturnType<typeof setTimeout>

  function blink() {
    eyeState = 1
    setTimeout(() => {
      eyeState = 2
      setTimeout(() => {
        eyeState = 1
        setTimeout(() => {
          eyeState = 0
          const nextBlink = 2000 + Math.random() * 4000
          blinkTimer = setTimeout(blink, nextBlink)
        }, 80)
      }, 100)
    }, 80)
  }

  $effect(() => {
    blinkTimer = setTimeout(blink, 3000)
    return () => clearTimeout(blinkTimer)
  })

  // 状態に応じた画像選択
  const currentImg = $derived.by(() => {
    const isHappy = status === 'Success' || mood === 'Happy'
    
    if (isHappy) {
      // 喜び状態（瞬きなし、または1枚絵）
      return c2Happy
    }
    
    // 通常状態（瞬き）
    if (eyeState === 0) return c1Open
    if (eyeState === 1) return c1Half
    return c1Close
  })
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div 
  class="chara-wrapper" 
  class:talking={isTalking}
  style="transform: scale({scale})"
  onclick={onClick}
>
  <img 
    src={currentImg} 
    alt="Character" 
    class="pixel-art"
  />
</div>

<style>
  .chara-wrapper {
    display: inline-block;
    line-height: 0;
    width: auto;
    height: auto;
    animation: floating 3s ease-in-out infinite;
    cursor: pointer;
    transition: transform 0.2s, opacity 0.5s; /* 透過度の変化を滑らかに */
    opacity: 0.6; /* 通常時は少し透ける */
  }

  .chara-wrapper.talking {
    opacity: 1; /* 喋っている時はくっきり表示 */
  }

  img.pixel-art {
    width: 128px; 
    height: auto;
    image-rendering: pixelated;
  }

  @keyframes floating {
    0%, 100% { transform: translateY(0); }
    50% { transform: translateY(-10px); }
  }
</style>
