<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const navLinks = [
  { label: 'Guide', link: '/docs/getting-started' },
  { label: 'CLI', link: '/docs/cli-reference' },
  { label: 'Architecture', link: '/docs/architecture' },
  { label: 'Changelog', link: '/docs/changelog' },
]

const icons = {
  moon: '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 3c.132 0 .263 0 .393 0a7.5 7.5 0 0 0 7.92 12.446a9 9 0 1 1 -8.313 -12.454z" />',
  sun: '<path stroke="none" d="M0 0h24v24H0z" fill="none"/><path d="M12 12m-4 0a4 4 0 1 0 8 0a4 4 0 1 0 -8 0" /><path d="M3 12h1m8 -9v1m8 8h1m-9 8v1m-6.4 -15.4l.7 .7m12.1 -.7l-.7 .7m0 11.4l.7 .7m-12.1 -.7l-.7 .7" />'
}

/* ── Theme Switching Logic ── */
const isDark = ref(false)

function applyTheme(dark: boolean) {
  isDark.value = dark
  if (dark) {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}

function toggleTheme() {
  const newDark = !isDark.value
  applyTheme(newDark)
  localStorage.setItem('dawg-theme', newDark ? 'dark' : 'light')
}

function handleStorage(e: StorageEvent) {
  if (e.key === 'dawg-theme') {
    applyTheme(e.newValue === 'dark')
  }
}

onMounted(() => {
  const stored = localStorage.getItem('dawg-theme')
  if (stored) {
    applyTheme(stored === 'dark')
  } else {
    applyTheme(false)
  }
  window.addEventListener('storage', handleStorage)
})

onUnmounted(() => {
  window.removeEventListener('storage', handleStorage)
})
</script>

<template>
  <header class="dh-nav">
    <div class="dh-nav-inner">
      <a class="dh-brand" href="/">
        <img src="/paw-dawg.svg" alt="DAWG" />
        <span>DAWG</span>
      </a>
      <nav class="dh-nav-links">
        <a v-for="l in navLinks" :key="l.link" :href="l.link">{{ l.label }}</a>
      </nav>
      <div class="dh-nav-actions">
        <button class="dh-theme-toggle" @click="toggleTheme" :aria-label="isDark ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="isDark ? icons.sun : icons.moon" />
        </button>
        <a class="dh-nav-signin" href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">GitHub</a>
        <a class="dh-cta dh-cta-sm" href="/docs/getting-started">Get started</a>
      </div>
    </div>
  </header>
</template>

<style scoped>
/* ── top nav ── */
.dh-nav {
  position: sticky; top: 0; z-index: 60;
  background: hsl(0 0% 100% / 0.85); backdrop-filter: blur(16px);
  border-bottom: 1px solid var(--color-border);
}
:global(html.dark) .dh-nav { background: hsl(240 18% 8% / 0.85); }
.dh-nav-inner {
  max-width: 1200px; margin: 0 auto; padding: 16px 24px;
  display: flex; align-items: center; gap: 32px;
}
.dh-brand { display: inline-flex; align-items: center; gap: 10px; text-decoration: none; margin-right: auto; }
.dh-brand img { width: 24px; height: 24px; }
.dh-brand span { color: var(--color-text-primary); font-weight: 700; font-size: 16px; letter-spacing: -0.01em; }
.dh-nav-links { display: flex; align-items: center; gap: 32px; }
.dh-nav-links a { color: var(--color-text-secondary); font-size: 15px; font-weight: 500; text-decoration: none; transition: color var(--duration-fast); }
.dh-nav-links a:hover { color: var(--color-text-primary); }
.dh-nav-actions { display: flex; align-items: center; gap: 20px; }
.dh-nav-signin { color: var(--color-text-secondary); font-size: 15px; font-weight: 500; text-decoration: none; transition: color var(--duration-fast); }
.dh-nav-signin:hover { color: var(--color-text-primary); }
.dh-theme-toggle {
  background: transparent; border: none; cursor: pointer;
  color: var(--color-text-secondary); width: 24px; height: 24px;
  display: flex; align-items: center; justify-content: center;
  transition: color var(--duration-fast);
}
.dh-theme-toggle:hover { color: var(--color-brand-500); }
.dh-theme-toggle svg { width: 20px; height: 20px; }
.dh-cta {
  display: inline-block; background: var(--color-brand-600); color: #fff;
  font-weight: 600; font-size: 15px; padding: 12px 24px; border-radius: var(--radius-full);
  text-decoration: none; border: none; white-space: nowrap; 
  transition: all var(--duration-fast); box-shadow: 0 4px 14px 0 rgba(107, 33, 217, 0.39);
}
.dh-cta:hover { background: var(--color-brand-700); transform: translateY(-2px); box-shadow: 0 6px 20px rgba(107, 33, 217, 0.23); }
.dh-cta-sm { padding: 8px 16px; font-size: 14px; }
</style>
