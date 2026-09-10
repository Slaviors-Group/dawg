<script setup lang="ts">
import { ref } from 'vue'

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

/* ── Tiny syntax highlighter for command/code lines ── */
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
  <section class="dh-section">
    <div class="dh-inner">
      <div class="dh-cols dh-cols-flip">
        <div class="dh-col-text">
          <p class="dh-kicker">CODE + CLI</p>
          <h2 class="dh-h2">Ship with tooling that does what you expect.</h2>
          <ul class="dh-checks">
            <li>
              <div class="dh-check-icon"><i class="ph-bold ph-check"></i></div>
              <span>Single static binary — no daemon, no server</span>
            </li>
            <li>
              <div class="dh-check-icon"><i class="ph-bold ph-check"></i></div>
              <span>Human-readable output with JSON underneath</span>
            </li>
            <li>
              <div class="dh-check-icon"><i class="ph-bold ph-check"></i></div>
              <span>CI-ready exit codes for pass / fail gates</span>
            </li>
          </ul>
          <a class="dh-cta dh-cta-outline" href="/docs/getting-started">Read the Docs</a>
        </div>
        
        <div class="dh-term-window">
          <div class="dh-term-header">
            <div class="dh-term-dots">
              <span style="background: #ff5f56"></span>
              <span style="background: #ffbd2e"></span>
              <span style="background: #27c93f"></span>
            </div>
            <div class="dh-langtabs">
              <button
                v-for="(t, i) in codeTabs"
                :key="t.label"
                :class="['dh-lang', { 'dh-lang-active': codeTab === i }]"
                @click="codeTab = i"
              >{{ t.label }}</button>
            </div>
            <button class="dh-copy" @click="copyCode">
              {{ copied ? 'Copied' : 'Copy' }}
            </button>
          </div>
          <div class="dh-term-body dh-mono">
            <div v-for="(l, i) in codeTabs[codeTab].lines" :key="i" class="dh-line">
              <template v-if="l.blank">&nbsp;</template>
              <template v-else-if="l.comment"><span class="dh-comment">{{ l.text }}</span></template>
              <template v-else><span class="dh-prompt">$</span> <span class="dh-cmd" v-html="highlightCmd(l.text || '')" /></template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dh-section { padding: 110px 24px; }
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }
.dh-mono { font-family: var(--font-mono); }
.dh-kicker { font-size: 13px; font-weight: 700; letter-spacing: 0.1em; color: var(--color-brand); margin: 0 0 16px; text-transform: uppercase; }
:global(:root.dark) .dh-kicker { color: var(--color-brand); }
.dh-h2 { font-size: clamp(32px, 4vw, 44px); font-weight: 700; line-height: 1.15; letter-spacing: -0.02em; margin: 0; color: var(--color-text-primary); }

.dh-cols { display: grid; grid-template-columns: 1fr 1fr; gap: 64px; align-items: center; }
.dh-cols-flip { grid-template-columns: 0.9fr 1.1fr; }
@media (max-width: 900px) { .dh-cols, .dh-cols-flip { grid-template-columns: 1fr; gap: 48px; } }
.dh-checks { list-style: none; margin: 32px 0 40px; padding: 0; display: flex; flex-direction: column; gap: 16px; }
.dh-checks li { display: flex; align-items: flex-start; gap: 16px; font-size: 16px; color: var(--color-text-primary); line-height: 1.5; font-weight: 500; }
.dh-check-icon {
  width: 24px; height: 24px; border-radius: 50%; background: var(--color-brand);
  color: #000; display: flex; align-items: center; justify-content: center;
  font-size: 14px; font-weight: bold; flex-shrink: 0;
}
:global(:root.dark) .dh-check-icon { background: var(--color-brand); color: #000; }

.dh-cta-outline {
  display: inline-block; padding: 12px 24px; border-radius: var(--radius-full);
  font-weight: 600; font-size: 15px; text-decoration: none; transition: all var(--duration-fast);
  background: transparent; border: 2px solid var(--color-brand); color: var(--color-text-primary);
}
:global(:root.dark) .dh-cta-outline { color: var(--color-text-primary); border-color: var(--color-brand); }
.dh-cta-outline:hover { background: var(--color-brand); color: #000; transform: translateY(0); }
:global(:root.dark) .dh-cta-outline:hover { background: var(--color-brand); color: #000; }

/* Terminal Window */
.dh-term-window {
  background: #0d0d12; border: 1px solid #2a2a35; border-radius: var(--radius-xl);
  overflow: hidden; box-shadow: 0 24px 60px -12px rgba(0,0,0,0.4);
}
.dh-term-header {
  display: flex; align-items: center; padding: 16px 20px; background: #16161d;
  border-bottom: 1px solid #2a2a35;
}
.dh-term-dots { display: flex; gap: 8px; margin-right: 24px; }
.dh-term-dots span { width: 12px; height: 12px; border-radius: 50%; }
.dh-langtabs { display: flex; gap: 4px; margin-right: auto; }
.dh-lang {
  background: transparent; border: none; cursor: pointer; font-family: var(--font-mono); font-size: 13px;
  color: #8f8f9c; padding: 6px 14px; border-radius: var(--radius-sm); transition: all var(--duration-fast);
}
.dh-lang-active { color: #fff; background: rgba(167, 139, 250, 0.15); }
.dh-copy {
  background: transparent; border: 1px solid #3f3f4e; color: #a1a1aa; font-size: 13px;
  padding: 6px 14px; border-radius: var(--radius-sm); cursor: pointer; transition: all var(--duration-fast);
}
.dh-copy:hover { color: #fff; border-color: var(--color-brand); }

.dh-term-body { padding: 24px; font-size: 14px; line-height: 2; overflow-x: auto; color: #d1d1d8; }
.dh-line { white-space: nowrap; }
.dh-comment { color: #6b7280; }
.dh-prompt { color: #4ade80; margin-right: 8px; }
.dh-cmd :deep(.tok-bin) { color: #fff; font-weight: 600; }
.dh-cmd :deep(.tok-flag) { color: #a78bfa; }
.dh-cmd :deep(.tok-val) { color: #38bdf8; }
</style>
