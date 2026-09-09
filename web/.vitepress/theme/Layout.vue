<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import { useData, useRouter } from 'vitepress'
import { ref, computed } from 'vue'

const { Layout: DefaultLayout } = DefaultTheme
const { frontmatter } = useData()
const router = useRouter()

const isHome = computed(() => frontmatter.value.layout === 'home')

/* ── Top nav ─────────────────────────────────────────────────────── */
const navLinks = [
  { label: 'Guide', link: '/docs/getting-started' },
  { label: 'CLI', link: '/docs/cli-reference' },
  { label: 'Architecture', link: '/docs/architecture' },
  { label: 'Changelog', link: '/docs/changelog' },
]

/* ── Hero command bar ────────────────────────────────────────────── */
const heroCmd = ref('dawg capture --url https://my-app.local')
function goDocs() {
  router.go('/docs/getting-started')
}

const stats = [
  { value: '<100MB', label: 'typical artifact' },
  { value: '100%', label: 'deterministic replay' },
  { value: '0', label: 'PII leaks past the gate' },
  { value: '7', label: 'CLI subcommands' },
]

const tooling = ['Go', 'Playwright', 'rrweb', 'mitmproxy', 'OPA', 'OCI', 'ORAS', 'Docker', 'Tauri v2', 'React 19']

/* ── Minimal line-icon set (stroke-based, matches the mono/technical voice) ── */
const icons: Record<string, string> = {
  rocket: '<path d="M12 3.2c2.6 1 4.3 3.8 4.3 7.3 0 1.9-.7 3.7-1.7 4.8l-.8 2.6-1.8-1.7-1.8 1.7-.8-2.6c-1-1.1-1.7-2.9-1.7-4.8 0-3.5 1.7-6.3 4.3-7.3z"/><path d="M9.4 13.4c-1.6.9-2.5 2.6-2.5 5.1 2.6 0 4.2-.9 5.1-2.5"/><circle cx="12" cy="9.6" r="1.3"/>',
  terminal: '<rect x="4" y="5" width="16" height="14" rx="2.2"/><path d="M8 10l2.4 2-2.4 2"/><path d="M13 14h3"/>',
  layers: '<path d="M12 4.2l7.4 3.7-7.4 3.7-7.4-3.7 7.4-3.7z"/><path d="M4.6 12.2L12 15.9l7.4-3.7"/><path d="M4.6 16.3L12 20l7.4-3.7"/>',
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

/* ── "Control every step" — numbered feature list with mockup cards ── */
const controlFeatures = [
  {
    icon: 'terminal',
    tone: 'violet',
    title: 'Capture every glitch as it happens.',
    desc: 'One command opens a real browser and starts recording DOM mutations, network traffic, DB diffs, and structured logs — everything that leads to the bug.',
    link: '/docs/getting-started',
    widget: 'capture',
  },
  {
    icon: 'shield',
    tone: 'blue',
    title: 'Sanitize before anything leaves your machine.',
    desc: 'PII and secrets are redacted locally with an OPA hard gate. If the policy fails, export is blocked outright — not just flagged.',
    link: '/docs/sanitizer-policy',
    widget: 'sanitize',
  },
  {
    icon: 'layers',
    tone: 'magenta',
    title: 'Replay it anywhere, byte-for-byte.',
    desc: 'Time frozen, random seeded, network served from cassettes. The exact same failure reproduces on any machine, every single time.',
    link: '/docs/architecture',
    widget: 'replay',
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
              <span class="dh-brand-arrow">→</span>
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
          <div class="dh-wash dh-wash-a" />
          <div class="dh-inner">
            <a class="dh-eyebrow" href="https://github.com/Slaviors-Group/dawg/releases/tag/alpha-3" target="_blank" rel="noopener noreferrer">
              <span class="dh-pulse" />
              <span class="dh-mono">v0.1.2-alpha — now in alpha</span>
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

            <!-- glowing command bar -->
            <form class="dh-cmdbar" @submit.prevent="goDocs">
              <span class="dh-mono dh-prompt">$</span>
              <input v-model="heroCmd" class="dh-mono dh-cmd-input" spellcheck="false" aria-label="Try a DAWG command" />
              <button type="submit" class="dh-cta">Get started</button>
            </form>

            <div class="dh-hero-ctas">
              <a class="dh-ghost" href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">Star on GitHub</a>
              <a class="dh-quiet" href="/docs/cli-reference">Read the CLI reference →</a>
            </div>
          </div>

          <!-- ══════════ WIDGET WALL (bento dashboard) ══════════ -->
          <div class="dh-wall">
            <div class="dh-widget dh-widget-chart">
              <div class="dh-widget-head"><span>Redaction rate</span><span class="dh-widget-delta dh-delta-down">-6.0%</span></div>
              <div class="dh-widget-big">14<span class="dh-widget-unit">/run</span></div>
              <svg class="dh-spark" viewBox="0 0 200 50" preserveAspectRatio="none">
                <polyline points="0,40 25,22 50,30 75,10 100,26 125,16 150,32 175,8 200,20" />
              </svg>
            </div>
            <div class="dh-widget dh-widget-list">
              <div class="dh-widget-head"><span>Group by</span></div>
              <div class="dh-widget-row"><span class="dh-dot dh-dot-violet" />Sanitizer policy</div>
              <div class="dh-widget-row dh-widget-row-muted">Artifact digest</div>
            </div>
            <div class="dh-widget dh-widget-table">
              <div class="dh-widget-head"><span>Rollout duration</span></div>
              <div class="dh-widget-tablerow"><span>reproduced</span><span>for</span><span class="dh-mono">6</span><span>runs</span></div>
              <div class="dh-widget-tablerow"><span>verified</span><span>for</span><span class="dh-mono">12</span><span>runs</span></div>
              <div class="dh-widget-tablerow"><span>recording</span><span>for</span><span class="dh-mono">24</span><span>runs</span></div>
            </div>
            <div class="dh-widget dh-widget-code">
              <div class="dh-widget-head dh-mono"><span class="dh-widget-tag">DAWG CAPTURE SDK</span></div>
              <pre class="dh-mono dh-widget-pre"><span class="tok-flag">import</span> { capture } <span class="tok-flag">from</span> <span class="tok-val">"@dawg/sdk"</span>

capture.start({
  url: <span class="tok-val">"https://my-app.local"</span>,
  redact: <span class="tok-bin">true</span>,
})</pre>
            </div>
            <div class="dh-widget dh-widget-stat">
              <div class="dh-widget-head"><span>Avg. determinism</span></div>
              <div class="dh-widget-huge">100%</div>
              <span class="dh-widget-tiny">last 7d</span>
            </div>
            <div class="dh-widget dh-widget-prompt">
              <div class="dh-widget-head"><span class="dh-dot dh-dot-violet" />Policy<span class="dh-widget-tag dh-widget-tag-right">OPA</span></div>
              <p class="dh-widget-copy">Deny-by-default export. Field-name, regex, and Gitleaks secret detection run before anything is sealed.</p>
            </div>
          </div>
        </section>

        <!-- ══════════ TOOLING STRIP (pill badges) ══════════ -->
        <section class="dh-strip">
          <p class="dh-mono dh-strip-label">BUILT ON BATTLE-TESTED TOOLING</p>
          <div class="dh-strip-row">
            <span v-for="t in tooling" :key="t" class="dh-strip-pill">{{ t }}</span>
          </div>
        </section>

        <!-- ══════════ CONTROL EVERY STEP (white section) ══════════ -->
        <section class="dh-white">
          <div class="dh-inner">
            <div class="dh-white-head">
              <h2>Control every stage<br />of the repro.</h2>
              <div class="dh-index">
                <span v-for="(f, i) in controlFeatures" :key="f.title" :class="['dh-index-dot', `dh-index-dot-${f.tone}`, { 'dh-index-dot-active': i === 0 }]">{{ String(i + 1).padStart(2, '0') }}</span>
              </div>
            </div>

            <div class="dh-feature-list">
              <div v-for="f in controlFeatures" :key="f.title" class="dh-feature-row">
                <div class="dh-feature-text">
                  <div class="dh-feature-title">
                    <span :class="['dh-feature-icon', `dh-feature-icon-${f.tone}`]">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" v-html="icons[f.icon]" />
                    </span>
                    <h3>{{ f.title }}</h3>
                  </div>
                  <p>{{ f.desc }}</p>
                  <a class="dh-learnmore" :href="f.link">Learn more <span>→</span></a>
                </div>

                <div :class="['dh-mockup', `dh-mockup-${f.tone}`]">
                  <svg class="dh-mockup-star dh-mockup-star-a" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0l2 9 9 2-9 2-2 9-2-9-9-2 9-2z" /></svg>
                  <svg class="dh-mockup-star dh-mockup-star-b" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0l2 9 9 2-9 2-2 9-2-9-9-2 9-2z" /></svg>

                  <div v-if="f.widget === 'capture'" class="dh-mockup-card">
                    <div class="dh-mockup-card-head dh-mono">dawg — capture</div>
                    <div class="dh-mockup-card-row"><span class="dh-ok">✓</span> browser session recording</div>
                    <div class="dh-mockup-card-row"><span class="dh-ok">✓</span> proxy intercepting traffic</div>
                    <div class="dh-mockup-card-row"><span class="dh-live">●</span> db tap attached</div>
                  </div>

                  <div v-else-if="f.widget === 'sanitize'" class="dh-mockup-card">
                    <div class="dh-mockup-card-head">Add rule</div>
                    <div class="dh-mockup-card-row"><span class="dh-radio dh-radio-on" />Field-name detection</div>
                    <div class="dh-mockup-card-row"><span class="dh-radio" />Gitleaks secret scan</div>
                    <div class="dh-mockup-card-row"><span class="dh-radio" />Synthetic replacement</div>
                  </div>

                  <div v-else class="dh-mockup-card">
                    <div class="dh-mockup-card-head">Replay result</div>
                    <div class="dh-mockup-card-row dh-mockup-card-big">exit 0 <span class="dh-widget-delta dh-delta-up">match</span></div>
                    <div class="dh-mockup-card-row dh-mockup-card-muted">screenshot + HTTP diff clean</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- ══════════ COPY / PASTE / GO ══════════ -->
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

        <!-- ══════════ FINAL CTA ══════════ -->
        <section class="dh-section">
          <div class="dh-inner">
            <div class="dh-cta-panel">
              <p class="dh-mono dh-kicker dh-kicker-center">OPEN SOURCE · APACHE-2.0</p>
              <h2>Ready to fetch your bug back?</h2>
              <p>Install the desktop bundle or the CLI, capture once, and never argue about repro steps again.</p>
              <div class="dh-cta-row">
                <a class="dh-cta" href="/docs/installation">Install DAWG</a>
                <a class="dh-ghost" href="https://github.com/Slaviors-Group/dawg/releases" target="_blank" rel="noopener noreferrer">View releases</a>
              </div>
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
</style>

<style scoped>
/* ═══ DAWG control room — dark, violet pulse, pill-soft ═══ */
.dawg-home {
  --dh-bg: #101016;
  --dh-panel: #1a1a24;
  --dh-edge: rgba(255, 255, 255, 0.09);
  --dh-text: #ffffff;
  --dh-sub: #d1d1d8;
  --dh-mute: #8f8f9c;
  --dh-violet: #a78bfa;
  --dh-violet-deep: #7c3aed;
  --dh-gradient: linear-gradient(179deg, #7c3aed 1.06%, #c4b5fd 123.42%);
  --dh-glow: 0 0 44px rgba(124, 58, 237, 0.35);
  --dh-mono: "JetBrains Mono", ui-monospace, SFMono-Regular, Menlo, monospace;
  --dh-sans: "Outfit", ui-sans-serif, system-ui, sans-serif;
  background: var(--dh-bg);
  color: var(--dh-text);
  font-family: var(--dh-sans);
  overflow: clip;
}
.dawg-home .dh-mono { font-family: var(--dh-mono); }

/* ── top nav (plain bar, not a floating pill) ── */
.dh-nav {
  position: sticky; top: 0; z-index: 60;
  background: rgba(16, 16, 22, 0.85); backdrop-filter: blur(14px);
  border-bottom: 1px solid var(--dh-edge);
}
.dh-nav-inner {
  max-width: 1200px; margin: 0 auto; padding: 16px 24px;
  display: flex; align-items: center; gap: 32px;
}
.dh-brand { display: inline-flex; align-items: center; gap: 8px; text-decoration: none; margin-right: auto; }
.dh-brand img { width: 24px; height: 24px; }
.dh-brand span { color: #fff; font-weight: 700; font-size: 16px; letter-spacing: -0.01em; }
.dh-brand-arrow { color: var(--dh-mute); font-size: 14px; }
.dh-nav-links { display: flex; align-items: center; gap: 28px; }
.dh-nav-links a { color: var(--dh-sub); font-size: 14px; font-weight: 500; text-decoration: none; }
.dh-nav-links a:hover { color: #fff; }
.dh-nav-actions { display: flex; align-items: center; gap: 20px; }
.dh-nav-signin { color: var(--dh-sub); font-size: 14px; text-decoration: none; }
.dh-nav-signin:hover { color: #fff; }
.dh-cta-sm { padding: 9px 20px !important; font-size: 14px !important; }
@media (max-width: 860px) { .dh-nav-links { display: none; } }

/* ── hero ── */
.dh-hero { position: relative; padding: 72px 24px 0; overflow: clip; }
.dh-wash { position: absolute; border-radius: 50%; filter: blur(100px); pointer-events: none; }
.dh-wash-a { width: 820px; height: 440px; top: -180px; left: 50%; transform: translateX(-50%); background: radial-gradient(closest-side, rgba(124, 58, 237, 0.3), transparent); }
.dh-inner { position: relative; z-index: 1; max-width: 1200px; margin: 0 auto; }

.dh-eyebrow {
  display: inline-flex; align-items: center; gap: 10px;
  padding: 8px 18px; border-radius: 60px;
  border: 1px solid rgba(167, 139, 250, 0.45);
  color: var(--dh-violet); font-size: 13px; text-decoration: none;
  margin-bottom: 32px; transition: border-color 0.2s;
}
a.dh-eyebrow:hover { border-color: var(--dh-violet); }
.dh-pulse { width: 8px; height: 8px; border-radius: 50%; background: #4ade80; box-shadow: 0 0 12px #4ade80; animation: dh-pulse 2s infinite; }
@keyframes dh-pulse { 50% { opacity: 0.4; } }

.dh-display {
  font-size: clamp(40px, 7vw, 84px);
  font-weight: 500; line-height: 1.02; letter-spacing: -0.02em;
  margin: 0 0 24px; text-wrap: balance;
}
.dh-accent { background: var(--dh-gradient); -webkit-background-clip: text; background-clip: text; -webkit-text-fill-color: transparent; }
.dh-sub { font-size: 18px; line-height: 1.6; color: var(--dh-sub); max-width: 620px; margin: 0 auto 40px; text-align: center; }
.dh-hero .dh-inner { text-align: center; }

/* glowing command bar */
.dh-cmdbar {
  display: flex; align-items: center; gap: 12px;
  max-width: 640px; margin: 0 auto; padding: 8px 8px 8px 24px;
  background: var(--dh-panel); border: 1px solid var(--dh-edge); border-radius: 10px;
  box-shadow: var(--dh-glow); text-align: left; flex-wrap: wrap;
}
.dh-prompt { color: #4ade80; font-size: 15px; }
.dh-cmd { color: #fff; font-size: 14px; flex: 1; min-width: 200px; overflow-x: auto; white-space: nowrap; }
.dh-cmd-input {
  flex: 1; min-width: 200px; background: transparent; border: none; outline: none;
  color: #fff; font-size: 14px; padding: 16px 0; caret-color: var(--dh-violet);
}
.dh-cmd-input::placeholder { color: #5b5b68; }
.dh-cta {
  display: inline-block; background: var(--dh-violet-deep); color: #fff;
  font-weight: 600; font-size: 15px; padding: 14px 30px; border-radius: 30px;
  text-decoration: none; border: none; white-space: nowrap; transition: filter 0.2s, transform 0.2s;
}
.dh-cta:hover { filter: brightness(1.15); transform: translateY(-1px); }

.dh-hero-ctas { display: flex; align-items: center; justify-content: center; gap: 28px; margin-top: 28px; flex-wrap: wrap; }
.dh-ghost {
  display: inline-block; padding: 13px 30px; border-radius: 30px;
  border: 1px solid rgba(167, 139, 250, 0.6); color: var(--dh-violet);
  font-weight: 600; font-size: 15px; text-decoration: none; transition: background 0.2s;
}
.dh-ghost:hover { background: rgba(124, 58, 237, 0.14); }
.dh-quiet { color: var(--dh-sub); font-size: 15px; text-decoration: none; }
.dh-quiet:hover { color: #fff; }

/* ── widget wall (bento dashboard, LaunchDarkly-style) ── */
.dh-wall {
  position: relative; z-index: 1; max-width: 1200px; margin: 56px auto 0;
  padding: 0 24px 0;
  display: grid; grid-template-columns: repeat(4, 1fr); grid-auto-rows: minmax(120px, auto);
  gap: 14px;
  mask-image: linear-gradient(to bottom, black 0%, black 60%, transparent 100%);
  -webkit-mask-image: linear-gradient(to bottom, black 0%, black 60%, transparent 100%);
  max-height: 300px; overflow: hidden;
}
.dh-widget {
  background: var(--dh-panel); border: 1px solid var(--dh-edge); border-radius: 16px;
  padding: 16px 18px; font-size: 13px; color: var(--dh-sub);
}
.dh-widget-head { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--dh-mute); margin-bottom: 10px; }
.dh-widget-delta { margin-left: auto; font-size: 11px; padding: 2px 8px; border-radius: 30px; background: rgba(255,255,255,0.06); }
.dh-delta-down { color: #4ade80; }
.dh-delta-up { color: var(--dh-violet); }
.dh-widget-big { font-size: 28px; font-weight: 500; color: #fff; }
.dh-widget-unit { font-size: 13px; color: var(--dh-mute); font-weight: 400; margin-left: 4px; }
.dh-spark { width: 100%; height: 32px; margin-top: 8px; }
.dh-spark polyline { fill: none; stroke: var(--dh-violet); stroke-width: 2; }
.dh-widget-row { display: flex; align-items: center; gap: 8px; padding: 8px 0; font-size: 13px; color: #fff; border-top: 1px solid var(--dh-edge); }
.dh-widget-row:first-of-type { border-top: none; }
.dh-widget-row-muted { color: var(--dh-mute); }
.dh-dot { width: 7px; height: 7px; border-radius: 50%; flex-shrink: 0; }
.dh-dot-violet { background: var(--dh-violet); box-shadow: 0 0 8px var(--dh-violet); }
.dh-widget-table { grid-column: span 1; }
.dh-widget-tablerow { display: flex; gap: 8px; font-size: 12px; padding: 6px 0; color: var(--dh-sub); border-top: 1px solid var(--dh-edge); }
.dh-widget-tablerow:first-of-type { border-top: none; }
.dh-widget-tablerow span:first-child { color: #fff; min-width: 76px; }
.dh-widget-code { grid-column: span 2; }
.dh-widget-tag { font-size: 11px; color: var(--dh-violet); letter-spacing: 0.08em; }
.dh-widget-tag-right { margin-left: auto; border: 1px solid var(--dh-edge); padding: 2px 8px; border-radius: 30px; color: var(--dh-sub); }
.dh-widget-pre { font-size: 12px; line-height: 1.6; white-space: pre; margin: 0; color: var(--dh-sub); }
.dh-widget-stat { display: flex; flex-direction: column; }
.dh-widget-huge { font-size: 34px; font-weight: 500; color: #fff; margin-top: 4px; }
.dh-widget-tiny { font-size: 11px; color: var(--dh-mute); margin-top: auto; }
.dh-widget-prompt { grid-column: span 2; }
.dh-widget-copy { font-size: 13px; color: var(--dh-sub); line-height: 1.5; margin: 0; }
@media (max-width: 900px) {
  .dh-wall { grid-template-columns: repeat(2, 1fr); max-height: 260px; }
  .dh-widget-code, .dh-widget-prompt { grid-column: span 2; }
}

/* ── tooling strip (pill badges) ── */
.dh-strip { position: relative; z-index: 1; padding: 48px 24px 8px; text-align: center; }
.dh-strip-label { font-size: 12px; letter-spacing: 0.167em; color: var(--dh-mute); margin-bottom: 22px; }
.dh-strip-row { display: flex; justify-content: center; gap: 12px; flex-wrap: wrap; max-width: 1000px; margin: 0 auto; }
.dh-strip-pill {
  font-family: var(--dh-mono); font-size: 13px; color: var(--dh-sub);
  border: 1px solid var(--dh-edge); border-radius: 60px; padding: 8px 18px;
  background: var(--dh-panel); transition: border-color 0.2s, color 0.2s;
}
.dh-strip-pill:hover { color: #fff; border-color: rgba(167, 139, 250, 0.5); }

/* ── sections ── */
.dh-section { padding: 96px 24px 8px; }
.dh-section:last-child { padding-bottom: 120px; }
.dh-kicker { font-size: 12px; letter-spacing: 0.167em; color: var(--dh-violet); margin: 0 0 16px; }
.dh-kicker-center { text-align: center; }
.dh-h2 { font-size: clamp(32px, 4vw, 48px); font-weight: 500; line-height: 1.1; letter-spacing: -0.01em; margin: 0 0 20px; }
.dh-h2-center { text-align: center; max-width: 640px; margin: 0 auto 56px; }
.dh-lead { font-size: 17px; color: var(--dh-sub); max-width: 560px; margin: 0 0 40px; line-height: 1.6; }

.dh-cols { display: grid; grid-template-columns: 1fr 1.1fr; gap: 48px; align-items: center; }
.dh-cols-flip { grid-template-columns: 0.9fr 1.1fr; }
@media (max-width: 900px) { .dh-cols, .dh-cols-flip { grid-template-columns: 1fr; } }
.dh-col-text h2 { font-size: clamp(28px, 3.4vw, 40px); font-weight: 500; line-height: 1.12; margin: 0 0 16px; }
.dh-col-text p { font-size: 16px; color: var(--dh-sub); line-height: 1.6; margin: 0 0 28px; }
.dh-checks { list-style: none; margin: 0 0 28px; padding: 0; display: flex; flex-direction: column; gap: 14px; }
.dh-checks li { display: flex; gap: 12px; font-size: 16px; color: #fff; line-height: 1.5; }
.dh-check { color: var(--dh-violet); font-weight: 700; }

/* terminal panel */
.dh-term { background: var(--dh-panel); border: 1px solid var(--dh-edge); border-radius: 16px; overflow: hidden; box-shadow: var(--dh-glow); }
.dh-term-bar { display: flex; align-items: center; gap: 7px; padding: 14px 18px; border-bottom: 1px solid var(--dh-edge); }
.dh-term-body { padding: 22px 24px; font-size: 14px; line-height: 2; overflow-x: auto; }
.dh-line { white-space: nowrap; }
.dh-ok { color: #4ade80; } .dh-dim { color: var(--dh-mute); }
.dh-live { color: var(--dh-violet); animation: dh-pulse 1.6s infinite; }
.dh-out { color: var(--dh-violet); }
.dh-comment { color: #5b5b68; }

/* syntax tokens inside .dh-cmd / .dh-widget-pre (kept to a tight, brand-safe palette) */
.dh-cmd :deep(.tok-bin), .dh-widget-code :deep(.tok-bin) { color: #fff; font-weight: 600; }
.dh-cmd :deep(.tok-flag), .dh-widget-code :deep(.tok-flag) { color: var(--dh-violet); }
.dh-cmd :deep(.tok-val), .dh-widget-code :deep(.tok-val) { color: #4ade80; }

.dh-langtabs { display: flex; gap: 4px; margin-right: auto; }
.dh-lang { background: transparent; border: none; cursor: pointer; font-family: var(--dh-mono); font-size: 13px; color: var(--dh-mute); padding: 6px 14px; border-radius: 30px; }
.dh-lang-active { color: #fff; background: rgba(124, 58, 237, 0.25); }
.dh-copy { background: transparent; border: 1px solid var(--dh-edge); color: var(--dh-sub); font-size: 12px; padding: 6px 14px; border-radius: 30px; cursor: pointer; }
.dh-copy:hover { color: #fff; border-color: var(--dh-violet); }

/* ── "Control every stage" white section (rounded, floats over dark canvas) ── */
.dh-white {
  position: relative; z-index: 1; background: #f7f7f9; color: #111114;
  border-radius: 40px; margin: 32px 16px 0; padding: 88px 24px;
}
.dh-white-head { display: flex; align-items: flex-end; justify-content: space-between; flex-wrap: wrap; gap: 24px; margin-bottom: 64px; }
.dh-white-head h2 { font-size: clamp(32px, 4.5vw, 52px); font-weight: 700; line-height: 1.08; letter-spacing: -0.01em; margin: 0; }
.dh-index { display: flex; gap: 8px; }
.dh-index-dot {
  font-family: var(--dh-mono); font-size: 13px; font-weight: 600; color: #111;
  border: 1px solid rgba(0,0,0,0.15); border-radius: 60px; width: 40px; height: 40px;
  display: inline-flex; align-items: center; justify-content: center;
}
.dh-index-dot-active.dh-index-dot-violet { background: var(--dh-violet-deep); color: #fff; border-color: transparent; }
.dh-index-dot-active.dh-index-dot-blue { background: #6d63f5; color: #fff; border-color: transparent; }
.dh-index-dot-active.dh-index-dot-magenta { background: #c026d3; color: #fff; border-color: transparent; }

.dh-feature-list { display: flex; flex-direction: column; }
.dh-feature-row {
  display: grid; grid-template-columns: 1fr 1fr; gap: 56px; align-items: center;
  padding: 48px 0; border-top: 1px solid rgba(0,0,0,0.1);
}
.dh-feature-row:first-child { border-top: none; padding-top: 0; }
@media (max-width: 900px) { .dh-feature-row { grid-template-columns: 1fr; } }

.dh-feature-title { display: flex; align-items: center; gap: 14px; margin-bottom: 4px; }
.dh-feature-title h3 { font-size: 24px; font-weight: 700; margin: 0; line-height: 1.25; }
.dh-feature-icon {
  display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0;
  width: 36px; height: 36px; border-radius: 10px; color: #111;
}
.dh-feature-icon svg { width: 18px; height: 18px; }
.dh-feature-icon-violet { background: linear-gradient(145deg, #c4b5fd, #7c3aed); color: #fff; }
.dh-feature-icon-blue { background: linear-gradient(145deg, #93c5fd, #4f46e5); color: #fff; }
.dh-feature-icon-magenta { background: linear-gradient(145deg, #f0abfc, #c026d3); color: #fff; }
.dh-feature-text p { font-size: 15px; color: #46464e; line-height: 1.6; margin: 16px 0 24px; max-width: 420px; }
.dh-learnmore {
  display: inline-flex; align-items: center; gap: 8px; font-size: 14px; font-weight: 600;
  color: #111; text-decoration: none; border: 1px solid rgba(0,0,0,0.18);
  border-radius: 30px; padding: 11px 22px; transition: background 0.2s;
}
.dh-learnmore:hover { background: rgba(0,0,0,0.05); }
.dh-learnmore span { transition: transform 0.2s; }
.dh-learnmore:hover span { transform: translateX(2px); }

.dh-mockup {
  position: relative; border-radius: 20px; padding: 20px; min-height: 220px;
  background-image:
    linear-gradient(to right, rgba(255,255,255,0.22) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255,255,255,0.22) 1px, transparent 1px);
  background-size: 36px 36px;
  display: flex; align-items: flex-end; overflow: hidden;
}
.dh-mockup-violet { background-color: #8b5cf6; background-image: linear-gradient(160deg, #a78bfa, #6d28d9), linear-gradient(to right, rgba(255,255,255,0.22) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.22) 1px, transparent 1px); background-size: cover, 36px 36px, 36px 36px; }
.dh-mockup-blue { background-color: #6366f1; background-image: linear-gradient(160deg, #93c5fd, #4338ca), linear-gradient(to right, rgba(255,255,255,0.22) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.22) 1px, transparent 1px); background-size: cover, 36px 36px, 36px 36px; }
.dh-mockup-magenta { background-color: #c026d3; background-image: linear-gradient(160deg, #f0abfc, #86198f), linear-gradient(to right, rgba(255,255,255,0.22) 1px, transparent 1px), linear-gradient(to bottom, rgba(255,255,255,0.22) 1px, transparent 1px); background-size: cover, 36px 36px, 36px 36px; }

.dh-mockup-star { position: absolute; color: rgba(255,255,255,0.85); }
.dh-mockup-star-a { width: 34px; height: 34px; top: 16px; right: 18px; }
.dh-mockup-star-b { width: 18px; height: 18px; top: 54px; right: 4px; }

.dh-mockup-card {
  position: relative; width: 100%; background: #16161c; border-radius: 14px;
  padding: 16px 18px; color: #fff; box-shadow: 0 20px 40px -12px rgba(0,0,0,0.5);
}
.dh-mockup-card-head { font-size: 13px; font-weight: 600; color: #fff; margin-bottom: 12px; }
.dh-mockup-card-row { display: flex; align-items: center; gap: 8px; font-size: 13px; color: #b8b8c2; padding: 6px 0; }
.dh-mockup-card-big { font-size: 20px; font-weight: 600; color: #fff; gap: 12px; }
.dh-mockup-card-muted { color: #7c7c88; font-size: 12px; }
.dh-radio { width: 14px; height: 14px; border-radius: 4px; border: 1.5px solid #4a4a54; flex-shrink: 0; }
.dh-radio-on { background: var(--dh-violet); border-color: var(--dh-violet); }


/* resource grid */
.dh-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 20px; margin-top: 40px; }
@media (max-width: 900px) { .dh-grid { grid-template-columns: 1fr; } }
.dh-card {
  position: relative; display: block; padding: 36px 32px 32px;
  background: var(--dh-panel); border: 1px solid var(--dh-edge); border-radius: 30px;
  text-decoration: none; transition: border-color 0.2s, transform 0.2s;
}
.dh-card:hover { border-color: rgba(167, 139, 250, 0.5); transform: translateY(-2px); box-shadow: var(--dh-glow); }
.dh-card-icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 40px; height: 40px; border-radius: 12px;
  background: rgba(124, 58, 237, 0.14); color: var(--dh-violet);
}
.dh-card-icon svg { width: 20px; height: 20px; }
.dh-card h3 { font-size: 19px; font-weight: 500; color: #fff; margin: 18px 0 8px; }
.dh-card p { font-size: 14px; color: var(--dh-mute); line-height: 1.55; margin: 0; padding-right: 28px; }
.dh-card-arrow { position: absolute; right: 28px; bottom: 30px; color: var(--dh-violet); font-size: 18px; }

/* final CTA */
.dh-cta-panel {
  text-align: center; padding: 80px 32px; border-radius: 30px;
  background: radial-gradient(ellipse 70% 90% at 50% 110%, rgba(124, 58, 237, 0.28), transparent), var(--dh-panel);
  border: 1px solid var(--dh-edge);
}
.dh-cta-panel h2 { font-size: clamp(32px, 4.5vw, 56px); font-weight: 500; line-height: 1.05; margin: 0 0 16px; }
.dh-cta-panel p { color: var(--dh-sub); font-size: 17px; max-width: 520px; margin: 0 auto 36px; line-height: 1.6; }
.dh-cta-row { display: flex; justify-content: center; gap: 16px; flex-wrap: wrap; }

/* ── footer ── */
.dh-footer { border-top: 1px solid var(--dh-edge); padding: 72px 24px 0; margin-top: 40px; }
.dh-foot-grid { display: grid; grid-template-columns: 1.4fr 1fr 1fr 1fr; gap: 40px; }
@media (max-width: 860px) { .dh-foot-grid { grid-template-columns: 1fr 1fr; } }
.dh-foot-brand p { color: var(--dh-mute); font-size: 14px; line-height: 1.6; margin: 18px 0 0; }
.dh-foot-head { font-size: 12px; letter-spacing: 0.167em; color: var(--dh-mute); margin: 0 0 18px; }
.dh-foot-col a { display: block; color: var(--dh-sub); font-size: 14px; text-decoration: none; margin-bottom: 12px; }
.dh-foot-col a:hover { color: #fff; }
.dh-foot-base {
  display: flex; justify-content: space-between; align-items: center; gap: 16px; flex-wrap: wrap;
  margin-top: 56px; padding: 24px 0; border-top: 1px solid var(--dh-edge);
  color: var(--dh-mute); font-size: 13px;
}
.dh-foot-base .dh-mono { font-size: 12px; color: var(--dh-violet); }
</style>
