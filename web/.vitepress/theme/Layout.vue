<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import { useData } from 'vitepress'
import { ref, computed } from 'vue'

const { Layout: DefaultLayout } = DefaultTheme
const { frontmatter } = useData()

const isHome = computed(() => frontmatter.value.layout === 'home')

/* ── Top nav ─────────────────────────────────────────────────────── */
const navLinks = [
  { label: 'Guide', link: '/docs/getting-started' },
  { label: 'CLI', link: '/docs/cli-reference' },
  { label: 'Architecture', link: '/docs/architecture' },
  { label: 'Changelog', link: '/docs/changelog' },
]


const tooling = ['Go', 'Playwright', 'rrweb', 'mitmproxy', 'OPA', 'OCI', 'ORAS', 'Docker', 'Tauri v2', 'React 19']

/* ── Minimal line-icon set (stroke-based, used only for resource links) ── */
const icons: Record<string, string> = {
  rocket: '<path d="M12 3.2c2.6 1 4.3 3.8 4.3 7.3 0 1.9-.7 3.7-1.7 4.8l-.8 2.6-1.8-1.7-1.8 1.7-.8-2.6c-1-1.1-1.7-2.9-1.7-4.8 0-3.5 1.7-6.3 4.3-7.3z"/><path d="M9.4 13.4c-1.6.9-2.5 2.6-2.5 5.1 2.6 0 4.2-.9 5.1-2.5"/><circle cx="12" cy="9.6" r="1.3"/>',
  terminal: '<rect x="4" y="5" width="16" height="14" rx="2.2"/><path d="M8 10l2.4 2-2.4 2"/><path d="M13 14h3"/>',
  layers: '<circle cx="12" cy="4.6" r="1.9"/><circle cx="5.5" cy="19.4" r="1.9"/><circle cx="18.5" cy="19.4" r="1.9"/><path d="M12 6.5v4.3"/><path d="M12 10.8L6.6 17.7"/><path d="M12 10.8l5.4 6.9"/>',
  shield: '<path d="M12 3.3l6.4 2.7v5.5c0 4.5-2.9 7.3-6.4 8.2-3.5-.9-6.4-3.7-6.4-8.2V6l6.4-2.7z"/>',
  branch: '<circle cx="6.5" cy="5.5" r="1.9"/><circle cx="6.5" cy="18.5" r="1.9"/><circle cx="17.5" cy="9.5" r="1.9"/><path d="M6.5 7.4v9.2"/><path d="M6.5 12c0-2.4 2-3.8 5-3.8h4"/>',
  clock: '<circle cx="12" cy="12" r="8.1"/><path d="M12 7.6v4.4l3.2 1.9"/>',
}

/* ── Resource cards ──────────────────────────────────────────────── */
const resources = [
  { icon: 'rocket', title: 'Getting Started', desc: 'Capture your first glitch in under five minutes.', link: '/docs/getting-started' },
  { icon: 'terminal', title: 'CLI Reference', desc: 'Every subcommand, flag, and exit code.', link: '/docs/cli-reference' },
  { icon: 'layers', title: 'Architecture', desc: 'How the six-stage pipeline fits together.', link: '/docs/architecture' },
  { icon: 'shield', title: 'Sanitizer Policy', desc: 'Rego gates, redaction rules, synthetic data.', link: '/docs/sanitizer-policy' },
  { icon: 'branch', title: 'Contributing', desc: 'Dev setup, standards, and PR workflow.', link: '/docs/contributing' },
  { icon: 'clock', title: 'Changelog', desc: 'What shipped in every alpha release.', link: '/docs/changelog' },
]

/* ── The pipeline — plain text, no icon chips or fake UI mockups ── */
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

/* ── Code tabs (PowerShell / Bash) ───────────────────────────────── */
const codeTab = ref(0)
const codeTabs = [
  {
    label: 'PowerShell',
    lines: [
      { comment: true, text: '# Inspect what was captured — no replay needed' },
      { text: 'dawg inspect .dawg/artifacts/9f3a…' },
      { blank: true },
      { comment: true, text: '# Push to your team registry' },
      { text: 'dawg push .dawg/artifacts/9f3a… --registry ghcr.io/org/dawgs' },
      { blank: true },
      { comment: true, text: '# Verify the fix on your branch' },
      { text: 'dawg verify .dawg/artifacts/9f3a… --against local' },
    ],
  },
  {
    label: 'Bash',
    lines: [
      { comment: true, text: '# One-time setup in your repo' },
      { text: 'dawg init && dawg doctor' },
      { blank: true },
      { comment: true, text: '# Capture, then stop when the bug appears' },
      { text: 'dawg capture --url http://localhost:3000' },
      { text: 'dawg capture stop' },
    ],
  },
]

const copied = ref(false)
async function copyCode() {
  const text = codeTabs[codeTab.value].lines
    .filter((l) => !l.blank)
    .map((l) => l.text)
    .join('\n')
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
  } catch { /* clipboard unavailable */ }
}

/* ── Tiny syntax highlighter for command/code lines — keeps the mono blocks
   feeling like real tool output instead of flat placeholder text ── */
function escapeHtml(s: string) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
}
function highlightCmd(text: string) {
  return escapeHtml(text)
    .split(' ')
    .map((tok, i) => {
      if (i === 0) return `<span class="tok-bin">${tok}</span>`
      if (tok.startsWith('--')) return `<span class="tok-flag">${tok}</span>`
      if (/^https?:\/\//.test(tok) || tok.includes('/') || tok.includes('…')) return `<span class="tok-val">${tok}</span>`
      return tok
    })
    .join(' ')
}
</script>

<template>
  <DefaultLayout :class="{ 'dh-home-root': isHome }">
    <template #home-hero-before>
      <div class="dawg-home">

        <!-- ══════════ TOP NAV ══════════ -->
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
              <a class="dh-nav-signin" href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">GitHub</a>
              <a class="dh-cta dh-cta-sm" href="/docs/getting-started">Get started</a>
            </div>
          </div>
        </header>

        <!-- ══════════ HERO ══════════ -->
        <section class="dh-hero">
          <div class="dh-inner">
            <a class="dh-eyebrow" href="https://github.com/Slaviors-Group/dawg/releases/tag/alpha-3" target="_blank" rel="noopener noreferrer">
              v0.1.2-alpha — open source on GitHub
            </a>
            <h1 class="dh-display">
              Capture every glitch.<br />
              <span class="dh-accent">Replay it anywhere.</span>
            </h1>
            <p class="dh-sub">
              DAWG packages sanitized, deterministic web-app bug reproductions
              into portable OCI artifacts — so the failure on your screen
              reproduces byte-for-byte on anyone's machine.
            </p>

            <div class="dh-hero-ctas">
              <a class="dh-cta" href="/docs/getting-started">Get started</a>
              <a class="dh-quiet" href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">View on GitHub →</a>
            </div>
          </div>
        </section>

        <!-- ══════════ TOOLING STRIP (plain text, no pill chrome) ══════════ -->
        <section class="dh-strip">
          <div class="dh-inner">
            <p class="dh-mono dh-strip-label">BUILT ON</p>
            <p class="dh-strip-row">
              <template v-for="(t, i) in tooling" :key="t">
                <span>{{ t }}</span><span v-if="i < tooling.length - 1" class="dh-strip-sep">·</span>
              </template>
            </p>
          </div>
        </section>

        <!-- ══════════ THE PIPELINE (plain text grid, no icon chips or mockups) ══════════ -->
        <section class="dh-pipeline">
          <div class="dh-inner">
            <div class="dh-pipeline-head">
              <p class="dh-mono dh-kicker">THE PIPELINE</p>
              <h2 class="dh-h2">Control every stage of the repro.</h2>
              <p class="dh-lead">
                One CLI, three deterministic stages — so a bug reported once
                reproduces exactly the same way forever.
              </p>
            </div>

            <div class="dh-pipeline-grid">
              <div v-for="p in pipeline" :key="p.title" class="dh-pipeline-cell">
                <span class="dh-mono dh-pipeline-step">{{ p.step }}</span>
                <h3>{{ p.title }}</h3>
                <p>{{ p.desc }}</p>
                <a class="dh-learnmore" :href="p.link">Learn more <span>→</span></a>
              </div>
            </div>
          </div>
        </section>

        <!-- ══════════ CODE + CLI ══════════ -->
        <section class="dh-section">
          <div class="dh-inner">
            <p class="dh-mono dh-kicker dh-kicker-center">CODE + CLI</p>
            <h2 class="dh-h2 dh-h2-center">Ship with tooling that does what you expect.</h2>
            <div class="dh-cols dh-cols-flip">
              <div class="dh-col-text">
                <ul class="dh-checks">
                  <li><span class="dh-check">✓</span>Single static binary — no daemon, no server</li>
                  <li><span class="dh-check">✓</span>Human-readable output with JSON underneath</li>
                  <li><span class="dh-check">✓</span>CI-ready exit codes for pass / fail gates</li>
                </ul>
                <a class="dh-quiet" href="/docs/getting-started">Follow the quickstart →</a>
              </div>
              <div class="dh-term">
                <div class="dh-term-bar">
                  <div class="dh-langtabs">
                    <button
                      v-for="(t, i) in codeTabs"
                      :key="t.label"
                      :class="['dh-lang', { 'dh-lang-active': codeTab === i }]"
                      @click="codeTab = i"
                    >{{ t.label }}</button>
                  </div>
                  <button class="dh-mono dh-copy" @click="copyCode">{{ copied ? 'Copied ✓' : 'Copy' }}</button>
                </div>
                <div class="dh-term-body dh-mono">
                  <div v-for="(l, i) in codeTabs[codeTab].lines" :key="i" class="dh-line">
                    <template v-if="l.blank">&nbsp;</template>
                    <template v-else-if="l.comment"><span class="dh-comment">{{ l.text }}</span></template>
                    <template v-else><span class="dh-prompt">$</span> <span class="dh-cmd" v-html="highlightCmd(l.text)" /></template>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ══════════ RESOURCE GRID ══════════ -->
        <section class="dh-section">
          <div class="dh-inner">
            <p class="dh-mono dh-kicker">RESOURCES</p>
            <h2 class="dh-h2">Everything else lives in the docs.</h2>
            <div class="dh-grid">
              <a v-for="r in resources" :key="r.title" class="dh-card" :href="r.link">
                <span class="dh-card-icon">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" v-html="icons[r.icon]" />
                </span>
                <h3>{{ r.title }}</h3>
                <p>{{ r.desc }}</p>
                <span class="dh-card-arrow">→</span>
              </a>
            </div>
          </div>
        </section>

        <!-- ══════════ FINAL CTA (plain, no gradient panel) ══════════ -->
        <section class="dh-final">
          <div class="dh-inner">
            <p class="dh-mono dh-kicker dh-kicker-center">OPEN SOURCE · APACHE-2.0</p>
            <h2>Install DAWG and stop arguing about repro steps.</h2>
            <p>Grab the desktop bundle or the CLI, capture once, and ship the fix with proof it's actually fixed.</p>
            <div class="dh-cta-row">
              <a class="dh-cta" href="/docs/installation">Install DAWG</a>
              <a class="dh-ghost" href="https://github.com/Slaviors-Group/dawg/releases" target="_blank" rel="noopener noreferrer">View releases</a>
            </div>
          </div>
        </section>

        <!-- ══════════ FOOTER ══════════ -->
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
                <p class="dh-mono dh-foot-head">PRODUCT</p>
                <a href="/docs/getting-started">Getting Started</a>
                <a href="/docs/installation">Installation</a>
                <a href="/docs/cli-reference">CLI Reference</a>
              </div>
              <div class="dh-foot-col">
                <p class="dh-mono dh-foot-head">RESOURCES</p>
                <a href="/docs/architecture">Architecture</a>
                <a href="/docs/sanitizer-policy">Sanitizer Policy</a>
                <a href="/docs/changelog">Changelog</a>
              </div>
              <div class="dh-foot-col">
                <p class="dh-mono dh-foot-head">PROJECT</p>
                <a href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">GitHub</a>
                <a href="https://github.com/Slaviors-Group/dawg/releases" target="_blank" rel="noopener noreferrer">Releases</a>
                <a href="/docs/contributing">Contributing</a>
              </div>
            </div>
            <div class="dh-foot-base">
              <span>© 2026 Slaviors-Group · Apache-2.0</span>
              <span class="dh-mono">v0.1.2-alpha</span>
            </div>
          </div>
        </footer>

      </div>
    </template>
  </DefaultLayout>
</template>

<style>
/* Home takes over chrome from the default theme */
.dh-home-root .VPNav,
.dh-home-root .VPFooter { display: none !important; }
.dh-home-root .VPContent { padding-top: 0 !important; }
.dh-home-root .VPContent.has-sidebar { padding-top: 0 !important; }
/* VitePress's default .VPHome adds a 6-8rem bottom margin, which left a slab
   of bare page background showing below our own footer. Kill it. */
.dh-home-root .VPHome { margin-bottom: 0 !important; }
</style>

<style scoped>
/* ═══ Editorial, light-first marketing shell — plain surfaces, one accent
   color used sparingly, hairline dividers instead of glow/gradient panels.
   Tracks the same design tokens as the docs pages (see theme/style.css) so
   light/dark mode and typography stay consistent site-wide. ═══ */
.dawg-home {
  background: var(--color-canvas);
  color: var(--color-text-primary);
  font-family: var(--font-sans);
}
.dawg-home .dh-mono { font-family: var(--font-mono); }
.dh-inner { max-width: 1120px; margin: 0 auto; padding: 0 24px; }

/* ── top nav ── */
.dh-nav {
  position: sticky; top: 0; z-index: 60;
  background: hsl(0 0% 100% / 0.82); backdrop-filter: blur(14px);
  border-bottom: 1px solid var(--color-border);
}
html.dark .dh-nav { background: hsl(240 18% 8% / 0.82); }
.dh-nav-inner {
  max-width: 1120px; margin: 0 auto; padding: 16px 24px;
  display: flex; align-items: center; gap: 32px;
}
.dh-brand { display: inline-flex; align-items: center; gap: 8px; text-decoration: none; margin-right: auto; }
.dh-brand img { width: 22px; height: 22px; }
.dh-brand span { color: var(--color-text-primary); font-weight: 600; font-size: 15px; letter-spacing: -0.01em; }
.dh-nav-links { display: flex; align-items: center; gap: 28px; }
.dh-nav-links a { color: var(--color-text-secondary); font-size: 14px; font-weight: 500; text-decoration: none; }
.dh-nav-links a:hover { color: var(--color-text-primary); }
.dh-nav-actions { display: flex; align-items: center; gap: 20px; }
.dh-nav-signin { color: var(--color-text-secondary); font-size: 14px; text-decoration: none; }
.dh-nav-signin:hover { color: var(--color-text-primary); }
.dh-cta-sm { padding: 8px 16px !important; font-size: 14px !important; }
@media (max-width: 860px) { .dh-nav-links { display: none; } }

/* ── hero — plain background, no blur washes ── */
.dh-hero { padding: 108px 24px 88px; text-align: center; }
.dh-hero .dh-inner { max-width: 760px; }

.dh-eyebrow {
  display: inline-block; font-size: 13px; color: var(--color-text-tertiary);
  text-decoration: none; border-bottom: 1px solid transparent;
  margin-bottom: 28px; transition: border-color var(--duration-fast), color var(--duration-fast);
}
.dh-eyebrow:hover { color: var(--color-text-primary); border-color: var(--color-border-strong); }

.dh-display {
  font-size: clamp(38px, 6vw, 68px);
  font-weight: 600; line-height: 1.08; letter-spacing: -0.02em;
  margin: 0 0 24px; text-wrap: balance;
}
.dh-accent { color: var(--color-brand-600); }
html.dark .dh-accent { color: var(--color-brand-400); }
.dh-sub { font-size: 17px; line-height: 1.6; color: var(--color-text-secondary); max-width: 560px; margin: 0 auto; }

.dh-hero-ctas { display: flex; align-items: center; justify-content: center; gap: 24px; margin-top: 36px; flex-wrap: wrap; }
.dh-cta {
  display: inline-block; background: var(--color-brand-600); color: #fff;
  font-weight: 600; font-size: 15px; padding: 12px 24px; border-radius: var(--radius-md);
  text-decoration: none; border: none; white-space: nowrap; transition: background var(--duration-fast), transform var(--duration-fast);
}
.dh-cta:hover { background: var(--color-brand-700); transform: translateY(-1px); }
.dh-ghost {
  display: inline-block; padding: 11px 23px; border-radius: var(--radius-md);
  border: 1px solid var(--color-border-strong); color: var(--color-text-primary);
  font-weight: 600; font-size: 15px; text-decoration: none; transition: background var(--duration-fast);
}
.dh-ghost:hover { background: var(--color-canvas-subtle); }
.dh-quiet { color: var(--color-text-secondary); font-size: 15px; text-decoration: none; }
.dh-quiet:hover { color: var(--color-text-primary); }


/* ── tooling strip — plain text row, no pill chrome ── */
.dh-strip { padding: 0 24px 88px; text-align: center; }
.dh-strip-label { font-size: 11px; letter-spacing: 0.14em; color: var(--color-text-disabled); margin-bottom: 16px; }
.dh-strip-row { font-size: 14px; color: var(--color-text-tertiary); line-height: 1.9; max-width: 720px; margin: 0 auto; }
.dh-strip-sep { margin: 0 10px; color: var(--color-border-strong); }

/* ── shared section rhythm: hairline dividers instead of colored panels ── */
.dh-section { padding: 88px 24px; border-top: 1px solid var(--color-border); }
.dh-pipeline { padding: 88px 24px; border-top: 1px solid var(--color-border); }
.dh-final { padding: 88px 24px 112px; border-top: 1px solid var(--color-border); text-align: center; }

.dh-kicker { font-size: 12px; letter-spacing: 0.14em; color: var(--color-brand-600); margin: 0 0 16px; }
html.dark .dh-kicker { color: var(--color-brand-400); }
.dh-kicker-center { text-align: center; }
.dh-h2 { font-size: clamp(28px, 3.4vw, 38px); font-weight: 600; line-height: 1.15; letter-spacing: -0.01em; margin: 0 0 16px; }
.dh-h2-center { text-align: center; max-width: 600px; margin: 0 auto 48px; }
.dh-lead { font-size: 16px; color: var(--color-text-secondary); max-width: 520px; margin: 0; line-height: 1.6; }

/* ── the pipeline — plain grid, hairline separated, no icon chips ── */
.dh-pipeline-head { max-width: 620px; margin-bottom: 48px; }
.dh-pipeline-grid {
  display: grid; grid-template-columns: repeat(3, 1fr); gap: 40px;
  border-top: 1px solid var(--color-border); padding-top: 32px;
}
@media (max-width: 900px) { .dh-pipeline-grid { grid-template-columns: 1fr; gap: 32px; } }
.dh-pipeline-step { display: block; font-size: 13px; color: var(--color-text-disabled); letter-spacing: 0.06em; margin-bottom: 14px; }
.dh-pipeline-cell h3 { font-size: 19px; font-weight: 600; line-height: 1.3; margin: 0 0 10px; }
.dh-pipeline-cell p { font-size: 15px; color: var(--color-text-secondary); line-height: 1.6; margin: 0 0 16px; }
.dh-learnmore { display: inline-flex; align-items: center; gap: 4px; font-size: 14px; font-weight: 600; color: var(--color-brand-600); text-decoration: none; }
html.dark .dh-learnmore { color: var(--color-brand-400); }
.dh-learnmore span { display: inline-block; transition: transform var(--duration-fast); }
.dh-learnmore:hover span { transform: translateX(3px); }

/* ── code + CLI section ── */
.dh-cols { display: grid; grid-template-columns: 1fr 1.1fr; gap: 48px; align-items: center; }
.dh-cols-flip { grid-template-columns: 0.9fr 1.1fr; }
@media (max-width: 900px) { .dh-cols, .dh-cols-flip { grid-template-columns: 1fr; } }
.dh-checks { list-style: none; margin: 0 0 24px; padding: 0; display: flex; flex-direction: column; gap: 12px; }
.dh-checks li { display: flex; gap: 12px; font-size: 15px; color: var(--color-text-primary); line-height: 1.5; }
.dh-check { color: var(--color-brand-600); font-weight: 700; }
html.dark .dh-check { color: var(--color-brand-400); }

/* terminal panel — dark code surface is expected here, not decorative glow */
.dh-term { background: #0b0b10; border: 1px solid var(--color-border); border-radius: var(--radius-lg); overflow: hidden; }
.dh-term-bar { display: flex; align-items: center; gap: 7px; padding: 12px 16px; border-bottom: 1px solid rgba(255, 255, 255, 0.08); }
.dh-term-body { padding: 20px 22px; font-size: 13px; line-height: 2; overflow-x: auto; color: #d1d1d8; }
.dh-line { white-space: nowrap; }
.dh-comment { color: #5b5b68; }
.dh-prompt { color: #4ade80; }
.dh-cmd :deep(.tok-bin) { color: #fff; font-weight: 600; }
.dh-cmd :deep(.tok-flag) { color: #a78bfa; }
.dh-cmd :deep(.tok-val) { color: #4ade80; }

.dh-langtabs { display: flex; gap: 4px; margin-right: auto; }
.dh-lang { background: transparent; border: none; cursor: pointer; font-family: var(--font-mono); font-size: 12px; color: #8f8f9c; padding: 5px 12px; border-radius: var(--radius-sm); }
.dh-lang-active { color: #fff; background: rgba(167, 139, 250, 0.22); }
.dh-copy { background: transparent; border: 1px solid rgba(255, 255, 255, 0.12); color: #d1d1d8; font-size: 12px; padding: 5px 12px; border-radius: var(--radius-sm); cursor: pointer; }
.dh-copy:hover { color: #fff; border-color: #a78bfa; }

/* ── resource grid — flat cards, no colored icon bubbles ── */
.dh-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-top: 40px; }
@media (max-width: 900px) { .dh-grid { grid-template-columns: repeat(2, 1fr); } }
@media (max-width: 620px) { .dh-grid { grid-template-columns: 1fr; } }
.dh-card {
  display: block; padding: 22px; border: 1px solid var(--color-border); border-radius: var(--radius-md);
  background: var(--color-surface); text-decoration: none; color: inherit; position: relative;
  transition: border-color var(--duration-fast);
}
.dh-card:hover { border-color: var(--color-border-strong); }
.dh-card-icon { display: inline-flex; width: 18px; height: 18px; color: var(--color-text-tertiary); margin-bottom: 16px; }
.dh-card-icon svg { width: 100%; height: 100%; }
.dh-card h3 { font-size: 15px; font-weight: 600; margin: 0 0 6px; }
.dh-card p { font-size: 13px; color: var(--color-text-secondary); line-height: 1.5; margin: 0; padding-right: 20px; }
.dh-card-arrow { position: absolute; top: 22px; right: 22px; color: var(--color-text-disabled); font-size: 14px; transition: transform var(--duration-fast); }
.dh-card:hover .dh-card-arrow { transform: translateX(2px); color: var(--color-text-tertiary); }

/* ── final CTA — plain, no gradient panel ── */
.dh-final h2 { font-size: clamp(26px, 3.2vw, 34px); font-weight: 600; line-height: 1.2; max-width: 560px; margin: 0 auto 14px; }
.dh-final p { font-size: 16px; color: var(--color-text-secondary); max-width: 460px; margin: 0 auto 32px; line-height: 1.6; }
.dh-cta-row { display: flex; align-items: center; justify-content: center; gap: 16px; flex-wrap: wrap; }

/* ── footer ── */
.dh-footer { border-top: 1px solid var(--color-border); padding: 56px 24px 32px; background: var(--color-canvas); }
.dh-foot-grid { display: grid; grid-template-columns: 1.4fr 1fr 1fr 1fr; gap: 32px; padding-bottom: 40px; }
@media (max-width: 860px) { .dh-foot-grid { grid-template-columns: 1fr 1fr; } }
.dh-foot-brand p { font-size: 14px; color: var(--color-text-tertiary); line-height: 1.6; margin-top: 14px; }
.dh-foot-head { font-size: 11px; letter-spacing: 0.1em; color: var(--color-text-disabled); margin-bottom: 14px; }
.dh-foot-col { display: flex; flex-direction: column; gap: 10px; }
.dh-foot-col a { font-size: 14px; color: var(--color-text-secondary); text-decoration: none; }
.dh-foot-col a:hover { color: var(--color-text-primary); }
.dh-foot-base {
  display: flex; align-items: center; justify-content: space-between; flex-wrap: wrap; gap: 8px;
  padding-top: 24px; border-top: 1px solid var(--color-border); font-size: 13px; color: var(--color-text-tertiary);
}
</style>
