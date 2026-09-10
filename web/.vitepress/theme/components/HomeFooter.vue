<template>
  <footer class="dh-footer">
    <div class="dh-inner">
      <div class="dh-foot-grid">
        <div class="dh-foot-brand">
          <a class="dh-brand" href="/">
            <img src="/paw-dawg.svg" alt="DAWG" />
            <span>DAWG</span>
          </a>
          <p>Digs Any Web-app Glitch.<br />Repro-as-artifact for developers.</p>
        </div>
        <div class="dh-foot-col">
          <p class="dh-foot-head">PRODUCT</p>
          <a href="/docs/getting-started">Getting Started</a>
          <a href="/docs/installation">Installation</a>
          <a href="/docs/cli-reference">CLI Reference</a>
        </div>
        <div class="dh-foot-col">
          <p class="dh-foot-head">RESOURCES</p>
          <a href="/docs/architecture">Architecture</a>
          <a href="/docs/sanitizer-policy">Sanitizer Policy</a>
          <a href="/docs/changelog">Changelog</a>
        </div>
        <div class="dh-foot-col">
          <p class="dh-foot-head">PROJECT</p>
          <a href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">GitHub</a>
          <a href="https://github.com/Slaviors-Group/dawg/releases" target="_blank" rel="noopener noreferrer">Releases</a>
          <a href="/docs/contributing">Contributing</a>
        </div>
      </div>
      <div class="dh-foot-base">
        <span>© 2026 Slaviors-Group · Apache-2.0</span>
        <span>v0.1.2-alpha</span>
      </div>
    </div>
  </footer>
</template>

<style scoped>
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

.dh-brand { display: inline-flex; align-items: center; gap: 10px; text-decoration: none; margin-right: auto; }
.dh-brand img { width: 24px; height: 24px; }
.dh-brand span { color: var(--color-text-primary); font-weight: 700; font-size: 16px; letter-spacing: -0.01em; }

/* ── footer ── */
.dh-footer { border-top: 1px solid var(--color-border); padding: 64px 24px 32px; background: var(--color-canvas); }
.dh-foot-grid { display: grid; grid-template-columns: 1.4fr 1fr 1fr 1fr; gap: 40px; padding-bottom: 48px; }
@media (max-width: 860px) { .dh-foot-grid { grid-template-columns: 1fr 1fr; } }
.dh-foot-brand p { font-size: 15px; color: var(--color-text-tertiary); line-height: 1.6; margin-top: 16px; }
.dh-foot-head { font-size: 12px; font-weight: 700; letter-spacing: 0.1em; color: var(--color-text-disabled); margin-bottom: 16px; }
.dh-foot-col { display: flex; flex-direction: column; gap: 12px; }
.dh-foot-col a { font-size: 15px; color: var(--color-text-secondary); text-decoration: none; font-weight: 500; transition: color var(--duration-fast); }
.dh-foot-col a:hover { color: var(--color-text-primary); }
.dh-foot-base {
  display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px;
  padding-top: 32px; border-top: 1px solid var(--color-border); font-size: 14px; color: var(--color-text-tertiary);
}
</style>
