<script setup lang="ts">
/* ── Resource cards ──────────────────────────────────────────────── */
const resources = [
  { icon: 'ph-rocket-launch', title: 'Getting Started', desc: 'Capture your first glitch in under five minutes.', link: '/docs/getting-started' },
  { icon: 'ph-terminal-window', title: 'CLI Reference', desc: 'Every subcommand, flag, and exit code.', link: '/docs/cli-reference' },
  { icon: 'ph-stack', title: 'Architecture', desc: 'How the six-stage pipeline fits together.', link: '/docs/architecture' },
  { icon: 'ph-shield-check', title: 'Sanitizer Policy', desc: 'Rego gates, redaction rules, synthetic data.', link: '/docs/sanitizer-policy' },
  { icon: 'ph-git-branch', title: 'Contributing', desc: 'Dev setup, standards, and PR workflow.', link: '/docs/contributing' },
  { icon: 'ph-clock-counter-clockwise', title: 'Changelog', desc: 'What shipped in every alpha release.', link: '/docs/changelog' },
]
</script>

<template>
  <section class="dh-section dh-bg-subtle">
    <div class="dh-inner">
      <div class="dh-section-head">
        <p class="dh-kicker">RESOURCES</p>
        <h2 class="dh-h2">Dive deeper into the docs.</h2>
      </div>
      <div class="dh-grid">
        <a v-for="r in resources" :key="r.title" class="dh-resource-card" :href="r.link">
          <div class="dh-card-icon-wrap">
            <i :class="['ph', r.icon, 'ph-lg']"></i>
          </div>
          <h3>{{ r.title }}</h3>
          <p>{{ r.desc }}</p>
        </a>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dh-section { padding: 110px 24px; }
.dh-bg-subtle { background: var(--color-canvas-subtle); }
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }
.dh-section-head { text-align: center; margin-bottom: 64px; }
.dh-kicker { font-size: 13px; font-weight: 700; letter-spacing: 0.1em; color: var(--color-brand); margin: 0 0 16px; text-transform: uppercase; }
:global(:root.dark) .dh-kicker { color: var(--color-brand); }
.dh-h2 { font-size: clamp(32px, 4vw, 44px); font-weight: 700; line-height: 1.15; letter-spacing: -0.02em; margin: 0; color: var(--color-text-primary); }

/* ── Resources Grid ── */
.dh-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 24px; margin-top: 48px; }
@media (max-width: 900px) { .dh-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 620px) { .dh-grid { grid-template-columns: 1fr; } }

.dh-resource-card {
  display: block; padding: 32px; border: 1px solid var(--color-border); border-radius: var(--radius-lg);
  background: var(--color-surface); text-decoration: none; color: inherit;
  transition: all var(--duration-fast);
}
.dh-resource-card:hover { border-color: var(--color-brand); transform: translateY(-2px); box-shadow: var(--shadow-card); }
.dh-card-icon-wrap {
  width: 40px; height: 40px; border-radius: 10px; background: var(--color-canvas-subtle);
  display: flex; align-items: center; justify-content: center; margin-bottom: 20px;
  color: var(--color-brand); transition: background var(--duration-fast);
}
:global(:root.dark) .dh-card-icon-wrap { background: hsl(240 15% 20%); color: var(--color-brand); }
.dh-resource-card:hover .dh-card-icon-wrap { background: var(--color-brand); color: #000; }
:global(:root.dark) .dh-resource-card:hover .dh-card-icon-wrap { background: hsl(260 50% 16%); }
.dh-card-icon-wrap i { font-size: 24px; }
.dh-resource-card h3 { font-size: 18px; font-weight: 600; margin: 0 0 8px; }
.dh-resource-card p { font-size: 15px; color: var(--color-text-secondary); line-height: 1.5; margin: 0; }
</style>
