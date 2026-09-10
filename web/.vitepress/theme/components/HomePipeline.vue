<script setup lang="ts">
/* ── The pipeline ── */
const pipeline = [
  {
    step: '01',
    title: 'Capture every glitch as it happens.',
    desc: 'One command opens a real browser and starts recording DOM mutations, network traffic, DB diffs, and structured logs — everything that leads to the bug.',
    link: '/docs/getting-started',
  },
  {
    step: '02',
    title: 'Sanitize before anything leaves your machine.',
    desc: 'PII and secrets are redacted locally with an OPA hard gate. If the policy fails, export is blocked outright — not just flagged.',
    link: '/docs/sanitizer-policy',
  },
  {
    step: '03',
    title: 'Replay it anywhere, byte-for-byte.',
    desc: 'Time frozen, random seeded, network served from cassettes. The exact same failure reproduces on any machine, every single time.',
    link: '/docs/architecture',
  },
]
</script>

<template>
  <section class="dh-section dh-bg-subtle">
    <div class="dh-inner">
      <div class="dh-section-head">
        <p class="dh-kicker">KEY FEATURES</p>
        <h2 class="dh-h2">Everything You Need<br/>to Manage Bugs Smarter</h2>
      </div>

      <div class="dh-grid-modern">
        <div v-for="p in pipeline" :key="p.title" class="dh-glass-card">
          <span class="dh-step-badge">{{ p.step }}</span>
          <h3>{{ p.title }}</h3>
          <p>{{ p.desc }}</p>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* ── general sections ── */
.dh-section { padding: 110px 24px; }
.dh-bg-subtle { background: var(--color-canvas-subtle); }
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }
.dh-section-head { text-align: center; margin-bottom: 64px; }
.dh-kicker { font-size: 13px; font-weight: 700; letter-spacing: 0.1em; color: var(--color-brand); margin: 0 0 16px; text-transform: uppercase; }
:global(:root.dark) .dh-kicker { color: var(--color-brand); }
.dh-h2 { font-size: clamp(32px, 4vw, 44px); font-weight: 700; line-height: 1.15; letter-spacing: -0.02em; margin: 0; color: var(--color-text-primary); }

/* ── modern grids & cards ── */
.dh-grid-modern {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px;
}
@media (max-width: 900px) { .dh-grid-modern { grid-template-columns: 1fr; } }

.dh-glass-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-xl);
  padding: 40px 32px;
  box-shadow: var(--shadow-card);
  transition: transform var(--duration-base), box-shadow var(--duration-base);
  position: relative; overflow: hidden;
}
.dh-glass-card:hover {
  transform: translateY(-4px); box-shadow: var(--shadow-modal);
}
.dh-step-badge {
  display: inline-flex; align-items: center; justify-content: center;
  width: 48px; height: 48px; border-radius: 12px;
  background: var(--color-brand); color: #000;
  font-size: 18px; font-weight: 700; margin-bottom: 24px;
}
:global(:root.dark) .dh-step-badge { background: var(--color-brand); color: #000; }

.dh-glass-card h3 { font-size: 22px; font-weight: 600; line-height: 1.3; margin: 0 0 12px; color: var(--color-text-primary); }
.dh-glass-card p { font-size: 16px; color: var(--color-text-secondary); line-height: 1.6; margin: 0; }
</style>
