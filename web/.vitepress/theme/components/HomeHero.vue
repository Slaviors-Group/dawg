<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const scrollY = ref(0)
const heroRef = ref<HTMLElement | null>(null)

function onScroll() {
  scrollY.value = window.scrollY
}

onMounted(() => window.addEventListener('scroll', onScroll, { passive: true }))
onUnmounted(() => window.removeEventListener('scroll', onScroll))

function tiltStyle() {
  const progress = Math.min(scrollY.value / 400, 1)
  const rotateX = 8 - progress * 8
  const scale = 0.92 + progress * 0.08
  return {
    transform: `perspective(1200px) rotateX(${rotateX}deg) scale(${scale})`,
    transition: 'transform 0.1s linear',
  }
}
</script>

<template>
  <div class="dh-hero-wrapper">
    <section class="dh-hero">
      <div class="dh-inner">
        <h1 class="dh-display">
          <span class="dh-heading-line">Capture every glitch.</span>
          <span class="dh-accent dh-heading-line">Replay it anywhere.</span>
        </h1>

        <p class="dh-sub">
          DAWG captures browser sessions through its desktop app and extension,
          sanitizes them locally, and packages them as validated OCI artifacts
          that can be cataloged, shared, and replayed.
        </p>

        <div class="dh-hero-ctas">
          <a class="dh-cta dh-cta-lg" href="/docs/getting-started">Get Started</a>
          <a class="dh-ghost dh-ghost-lg" href="https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_x64-setup.exe">Download for Windows</a>
          <a class="dh-options" href="/docs/installation">Installation options &rarr;</a>
        </div>
      </div>

      <div class="dh-preview-wrap" ref="heroRef">
        <div class="dh-preview-frame" :style="tiltStyle()">
          <img
            src="/assets/images/desktop-preview.png"
            alt="DAWG desktop dashboard showing captured bug-reproduction artifacts"
            width="1917"
            height="1047"
            fetchpriority="high"
            decoding="async"
            class="dh-preview-img"
          />
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.dh-hero-wrapper {
  padding: 16px 16px 0;
  margin-bottom: 80px;
}

.dh-hero {
  position: relative;
  padding: 120px 24px 0;
  text-align: center;
  background-color: #ffffff;
  background-image:
    radial-gradient(80% 80% at 100% 100%, rgba(209, 200, 240, 0.9) 0%, rgba(255, 255, 255, 0) 80%),
    radial-gradient(80% 80% at 0% 100%, rgba(244, 220, 196, 0.9) 0%, rgba(255, 255, 255, 0) 80%);
  border-radius: 24px 24px 60px 60px;
  overflow: hidden;
}

:global(:root.dark) .dh-hero {
  background-color: var(--color-canvas);
  background-image:
    radial-gradient(80% 80% at 100% 100%, rgba(209, 200, 240, 0.15) 0%, transparent 80%),
    radial-gradient(80% 80% at 0% 100%, rgba(244, 220, 196, 0.15) 0%, transparent 80%);
}

.dh-inner { max-width: 800px; margin: 0 auto; position: relative; z-index: 10; }

@keyframes textReveal {
  0% {
    opacity: 0;
    transform: translateY(30px) scale(0.98);
    filter: blur(8px);
  }
  100% {
    opacity: 1;
    transform: translateY(0) scale(1);
    filter: blur(0);
  }
}

.dh-display {
  font-size: clamp(42px, 6vw, 76px);
  font-weight: 400; line-height: 1.1; letter-spacing: -0.03em;
  margin: 0 0 24px; text-wrap: balance; color: var(--color-text-primary);
  opacity: 0;
  animation: textReveal 1s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
.dh-heading-line { display: block; }
.dh-accent { color: var(--color-brand); }
.dh-sub { 
  font-size: 18px; line-height: 1.6; color: var(--color-text-secondary); max-width: 640px; margin: 0 auto; font-weight: 400; 
  opacity: 0;
  animation: textReveal 1s cubic-bezier(0.16, 1, 0.3, 1) 0.1s forwards;
}

.dh-hero-ctas { 
  display: flex; align-items: center; justify-content: center; gap: 20px; margin-top: 24px; flex-wrap: wrap; margin-bottom: 36px; 
  opacity: 0;
  animation: textReveal 1s cubic-bezier(0.16, 1, 0.3, 1) 0.2s forwards;
}
.dh-cta {
  display: inline-block; background: var(--color-brand); color: #000;
  font-weight: 600; font-size: 15px; padding: 12px 24px; border-radius: var(--radius-full);
  text-decoration: none; border: none; white-space: nowrap;
  transition: all var(--duration-fast);
}
.dh-cta:hover { filter: brightness(0.9); transform: translateY(-2px); }
.dh-cta-lg { padding: 16px 36px; font-size: 16px; }

.dh-ghost {
  display: inline-block; padding: 12px 24px; border-radius: var(--radius-full);
  border: 1px solid var(--color-border-strong); color: var(--color-text-primary);
  font-weight: 600; font-size: 15px; text-decoration: none; transition: background var(--duration-fast);
}
.dh-ghost-lg { padding: 16px 36px; font-size: 16px; }
.dh-ghost:hover { background: var(--color-surface-hover); }
.dh-options { width: 100%; color: var(--color-text-secondary); font-size: 14px; font-weight: 600; text-decoration: none; }
.dh-options:hover { color: var(--color-text-primary); }

.dh-preview-wrap {
  position: relative;
  z-index: 10;
  max-width: 1100px;
  margin: 0 auto;
}

.dh-preview-frame {
  border-radius: 16px 16px 0 0;
  overflow: hidden;
  transform-origin: bottom center;
  will-change: transform;
}

.dh-preview-img {
  width: 100%;
  display: block;

  pointer-events: none;
  -webkit-user-drag: none;
  user-select: none;
}

@media (max-width: 768px) {
  .dh-preview-wrap { max-width: 100%; }
}
</style>
