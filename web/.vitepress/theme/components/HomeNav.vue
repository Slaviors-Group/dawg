<script setup lang="ts">

const navLinks = [
  { label: 'Guide', link: '/docs/getting-started' },
  { label: 'CLI', link: '/docs/cli-reference' },
  { label: 'Architecture', link: '/docs/architecture' },
  { label: 'Changelog', link: '/docs/changelog' },
]

</script>

<template>
  <header class="dh-nav">
    <div class="dh-nav-inner">
      <a class="dh-brand" href="/">
        <img src="/paw-dawg.svg" alt="" width="24" height="24" decoding="async" />
        <span>DAWG</span>
      </a>
      <nav class="dh-nav-links">
        <a v-for="l in navLinks" :key="l.link" :href="l.link">{{ l.label }}</a>
      </nav>
      <div class="dh-nav-actions">

        <a class="dh-nav-signin" href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">GitHub</a>
        <a class="dh-nav-signin" href="https://github.com/sponsors/Slaviors-Group" target="_blank" rel="noopener noreferrer">Sponsor</a>
        <a class="dh-cta dh-cta-sm" href="/docs/installation">Download</a>
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

.dh-cta {
  display: inline-block; background: var(--color-brand-600); color: #fff;
  font-weight: 600; font-size: 15px; padding: 12px 24px; border-radius: var(--radius-full);
  text-decoration: none; border: none; white-space: nowrap; 
  transition: all var(--duration-fast); box-shadow: 0 4px 14px 0 rgba(107, 33, 217, 0.39);
}
.dh-cta:hover { background: var(--color-brand-700); transform: translateY(-2px); box-shadow: 0 6px 20px rgba(107, 33, 217, 0.23); }
.dh-cta-sm { padding: 8px 16px; font-size: 14px; }
</style>
