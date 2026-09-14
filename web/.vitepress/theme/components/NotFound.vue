<script setup lang="ts">
import { useData } from 'vitepress'
import { onMounted, ref } from 'vue'

const { site } = useData()
const rootPath = ref('/')

onMounted(() => {
  if (window.location.pathname.startsWith('/id/')) {
    rootPath.value = '/id/'
  } else {
    rootPath.value = site.value.base || '/'
  }
})
</script>

<template>
  <div class="dh-hero-wrapper">
    <section class="dh-hero">
      <div class="dh-inner">
        <p class="dh-display" style="font-size: clamp(80px, 12vw, 150px); margin-bottom: 0; line-height: 1;">
          <span class="dh-accent">404</span>
        </p>
        <h1 class="dh-display" style="margin-top: 10px;">
          PAGE NOT FOUND
        </h1>
        <p class="dh-sub">
          But if you don't change your direction, and if you keep looking, you may end up where you are heading.
        </p>
        <div class="dh-hero-ctas">
          <a class="dh-cta dh-cta-lg" :href="rootPath" aria-label="go to home">Take me home</a>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
.dh-hero-wrapper {
  padding: 16px 16px 0;
  min-height: calc(100vh - 200px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.dh-hero {
  position: relative;
  padding: 80px 24px 80px;
  text-align: center;
  width: 100%;
}

:global(:root.dark) .dh-hero {
  /* No special dark mode background needed if it's plain */
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
</style>
