<script setup lang="ts">
const tooling = ['Go', 'Playwright', 'rrweb', 'mitmproxy', 'OPA', 'OCI', 'ORAS', 'Docker', 'Tauri v2', 'React 19']
</script>

<template>
  <section class="dh-strip">
    <div class="dh-inner">
      <p class="dh-mono dh-strip-label">TRUSTED TECHNOLOGIES</p>
      <div class="dh-strip-glass">
        <span v-for="t in tooling" :key="t" class="dh-tool-pill">{{ t }}</span>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* ── tooling strip ── */
.dh-strip { padding: 0 24px 80px; text-align: center; }
.dh-strip-label { font-size: 12px; letter-spacing: 0.15em; color: var(--color-text-tertiary); margin-bottom: 24px; font-weight: 600; }
.dh-strip-glass { 
  display: flex; flex-wrap: wrap; justify-content: center; gap: 12px; max-width: 900px; margin: 0 auto;
}
.dh-tool-pill {
  background: var(--color-surface); border: 1px solid var(--color-border);
  padding: 8px 16px; border-radius: var(--radius-full); font-size: 14px;
  color: var(--color-text-secondary); font-weight: 500;
  box-shadow: 0 2px 8px -2px rgba(0,0,0,0.05);
}
</style>
