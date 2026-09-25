<script setup lang="ts">
import { useScrollReveal } from '../composables/useScrollReveal'

const { isRevealed, sectionRef } = useScrollReveal()

const features = [
  {
    title: 'Bundled, Local-First Runtime',
    desc: 'The packaged desktop app ships the engine, Node.js, mitmdump, and Playwright Chromium, so there is nothing extra to install and every capture stays on your machine.',
    mockType: 'table',
    panelTitle: 'Bundled components',
    rows: [
      { label: 'Engine', meta: 'bundled', ok: true },
      { label: 'Node.js', meta: 'bundled', ok: true },
      { label: 'mitmdump', meta: 'bundled', ok: true },
      { label: 'Playwright Chromium', meta: 'bundled', ok: true }
    ]
  },
  {
    title: 'Engine Doctor Diagnostics',
    desc: 'One-click health checks verify each runtime component, the schema, the OPA policy, and the bundled extension manifest, then flag anything missing before you start.',
    mockType: 'chart',
    panelTitle: 'Engine Doctor',
    rows: [
      { label: 'Runtime', meta: 'ready', ok: true },
      { label: 'Schema', meta: 'valid', ok: true },
      { label: 'OPA policy', meta: 'default.rego', ok: true },
      { label: 'Extension', meta: 'manifest found', ok: true }
    ]
  }
]

// Representative engine log stream for the bottom card (matches the desktop Log Console format)
const logLines = [
  { level: 'SYSTEM', text: 'engine ready · bundled runtime' },
  { level: 'INFO', text: 'session started · https://staging.example.com' },
  { level: 'INFO', text: 'extension connected · recording active tab' },
  { level: 'INFO', text: 'rrweb 1,284 events · 96 actions · 41 requests' },
  { level: 'WARN', text: 'sanitize: 12 PII fields replaced · 4 secrets redacted' },
  { level: 'SYSTEM', text: 'OPA allow · packaged checkout-timeout.dawg' },
  { level: 'INFO', text: 'replay diagnostics ready · screenshot saved' }
]
</script>

<template>
  <section class="dh-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-inner">
      <div class="dh-head">
        <h2 class="dh-h2">
          Ship with tooling that does
          <span class="dh-heading-line">what you expect.</span>
        </h2>
        <p class="dh-sub">One engine powers the DAWG desktop app and browser extension, with readable logs and live status at every step of capture, packaging, and replay.</p>
      </div>

      <!-- Two Columns -->
      <div class="dh-grid-2">
        <div v-for="(feat, idx) in features" :key="feat.title" class="dh-feature-card" :style="{ transitionDelay: `${0.1 + idx * 0.25}s` }">
          <div class="dh-mock-wrap">
            <div class="dh-mock-glow"></div>
            <div class="dh-mock-ui">
              <div class="dh-mock-header dh-mock-header--title">
                <span class="dh-win-dots">
                  <span class="dh-win-dot dh-win-dot--r"></span>
                  <span class="dh-win-dot dh-win-dot--y"></span>
                  <span class="dh-win-dot dh-win-dot--g"></span>
                </span>
                <span class="dh-mock-panel-title">{{ feat.panelTitle }}</span>
              </div>
              <div class="dh-mock-rows">
                <div v-for="row in feat.rows" :key="row.label" class="dh-mock-row">
                  <span class="dh-row-dot" :class="row.ok ? 'is-ok' : 'is-warn'"></span>
                  <span class="dh-row-label">{{ row.label }}</span>
                  <span class="dh-row-meta">{{ row.meta }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="dh-feature-text">
            <h3>{{ feat.title }}</h3>
            <p>{{ feat.desc }}</p>
          </div>
        </div>
      </div>

      <!-- Bottom Centered Column -->
      <div class="dh-feature-card dh-card-center" style="transition-delay: 0.6s;">
        <div class="dh-mock-wrap dh-mock-wrap-lg">
          <div class="dh-mock-glow dh-glow-lg"></div>
          <div class="dh-mock-ui dh-ui-lg">
             <div class="dh-mock-header dh-mock-header--title">
               <span class="dh-win-dots">
                 <span class="dh-win-dot dh-win-dot--r"></span>
                 <span class="dh-win-dot dh-win-dot--y"></span>
                 <span class="dh-win-dot dh-win-dot--g"></span>
               </span>
               <span class="dh-mock-panel-title">Log Console</span>
             </div>
             <div class="dh-log-stream">
               <div v-for="(line, i) in logLines" :key="i" class="dh-log-line">
                 <span class="dh-log-tag" :class="`dh-log-${line.level.toLowerCase()}`">[{{ line.level }}]</span>
                 <span class="dh-log-text">{{ line.text }}</span>
               </div>
             </div>
          </div>
        </div>
        <div class="dh-feature-text">
          <h3>Readable Logs and Live Status</h3>
          <p>A real-time engine log stream, live capture indicator, and packaging progress keep every step visible, with clear status badges so you always know what DAWG is doing.</p>
        </div>
      </div>

    </div>
  </section>
</template>

<style scoped>
.dh-section { padding: 110px 24px; background: #ffffff; overflow: hidden; }
:global(:root.dark) .dh-section { background: var(--color-canvas); }

.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

.dh-head {
  text-align: center;
  margin-bottom: 80px;
  display: flex;
  flex-direction: column;
  align-items: center;
}

/* Matches "Title" sizing request */
.dh-heading-line { display: block; }
.dh-h2 { 
  font-size: clamp(36px, 5vw, 64px); font-weight: 400; 
  line-height: 1.1; letter-spacing: -0.03em; margin: 0 0 24px; 
  color: var(--color-text-primary); text-wrap: balance;
}

/* Matches "Desc" sizing request */
.dh-sub {
  font-size: 18px; line-height: 1.6; color: var(--color-text-secondary); 
  margin: 0; font-weight: 400; max-width: 600px; text-wrap: balance;
}

/* ── Scroll Reveal Initial States ── */
.dh-head {
  opacity: 0;
  transform: translateY(30px) scale(0.98);
  filter: blur(8px);
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1), filter 1s cubic-bezier(0.16, 1, 0.3, 1);
}
.dh-feature-card {
  opacity: 0;
  transform: translateY(40px) scale(0.98);
  filter: blur(8px);
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1), filter 1s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Revealed States ── */
.dh-revealed .dh-head { opacity: 1; transform: translateY(0) scale(1); filter: blur(0); }
.dh-revealed .dh-feature-card { opacity: 1; transform: translateY(0) scale(1); filter: blur(0); }

/* ── Grid Layout ── */
.dh-grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 64px;
  margin-bottom: 80px;
}
@media (max-width: 900px) {
  .dh-grid-2 { grid-template-columns: 1fr; gap: 80px; }
}

.dh-feature-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
}

.dh-feature-text { margin-top: 32px; max-width: 480px; }
.dh-feature-text h3 {
  font-size: 22px; font-weight: 600; line-height: 1.3;
  margin: 0 0 12px; color: var(--color-text-primary);
}
.dh-feature-text p {
  font-size: 16px; color: var(--color-text-secondary);
  line-height: 1.6; margin: 0; font-weight: 400;
}

/* ── Mock UI Graphics ── */
.dh-mock-wrap {
  position: relative;
  width: 100%;
  height: 240px;
  display: flex;
  justify-content: center;
  align-items: flex-end; /* Push UI down */
  /* Fade out bottom edge */
  mask-image: linear-gradient(to bottom, black 60%, transparent 100%);
  -webkit-mask-image: linear-gradient(to bottom, black 60%, transparent 100%);
}

.dh-mock-wrap-lg {
  height: 320px;
  max-width: 800px;
  margin: 0 auto;
}

.dh-mock-glow {
  position: absolute;
  top: 50%; left: 50%;
  transform: translate(-50%, -50%);
  width: 200px; height: 200px;
  background: var(--color-brand);
  filter: blur(80px);
  opacity: 0.15;
  z-index: 1;
}
.dh-glow-lg { width: 400px; height: 200px; opacity: 0.1; }

.dh-mock-ui {
  position: relative;
  z-index: 2;
  width: 80%;
  height: 85%;
  background: #ffffff;
  border: 1px solid var(--color-border);
  border-radius: 16px 16px 0 0;
  box-shadow: 0 -10px 40px rgba(0,0,0,0.04);
  padding: 24px;
  display: flex;
  flex-direction: column;
}
:global(:root.dark) .dh-mock-ui { background: var(--color-surface); box-shadow: 0 -10px 40px rgba(0,0,0,0.2); }

.dh-ui-lg { width: 100%; }

.dh-mock-header {
  height: 24px;
  background: rgba(0,0,0,0.03);
  border-radius: 6px;
  margin-bottom: 24px;
}

/* Title-bar variant of the mock header (top feature cards) */
.dh-mock-header--title {
  display: flex;
  align-items: center;
  gap: 10px;
  height: auto;
  background: transparent;
  border-radius: 0;
  padding-bottom: 14px;
  margin-bottom: 16px;
  border-bottom: 1px solid var(--color-border-subtle, rgba(0,0,0,0.06));
}
.dh-win-dots { display: inline-flex; gap: 6px; }
.dh-win-dot { width: 10px; height: 10px; border-radius: 50%; display: inline-block; }
.dh-win-dot--r { background: #ff5f57; }
.dh-win-dot--y { background: #febc2e; }
.dh-win-dot--g { background: #28c840; }
.dh-mock-panel-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--color-text-secondary);
}

/* Status rows (real text content, no images) */
.dh-mock-rows { display: flex; flex-direction: column; gap: 12px; }
.dh-mock-row { display: flex; align-items: center; gap: 10px; }
.dh-row-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dh-row-dot.is-ok { background: var(--color-success-dot, #2ecc71); }
.dh-row-dot.is-warn { background: var(--color-warning-dot, #f1c40f); }
.dh-row-label { font-size: 13px; font-weight: 500; color: var(--color-text-primary); }
.dh-row-meta {
  margin-left: auto;
  font-size: 11px;
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
  color: var(--color-text-secondary);
  background: rgba(0,0,0,0.03);
  border-radius: 999px;
  padding: 2px 9px;
  white-space: nowrap;
}
:global(:root.dark) .dh-row-meta { background: rgba(255,255,255,0.06); }

/* Live engine log stream (bottom card) */
.dh-log-stream {
  display: flex;
  flex-direction: column;
  gap: 8px;
  text-align: left;
  font-family: var(--font-mono, ui-monospace, SFMono-Regular, Menlo, monospace);
  font-size: 12px;
  line-height: 1.5;
}
.dh-log-line { display: flex; gap: 8px; align-items: baseline; }
.dh-log-tag { font-weight: 700; flex-shrink: 0; color: var(--color-text-secondary); }
.dh-log-text { color: var(--color-text-secondary); }
.dh-log-system { color: var(--color-brand); }
.dh-log-info { color: var(--color-info-dot, #3b82f6); }
.dh-log-warn { color: var(--color-warning-dot, #f1c40f); }
.dh-log-error { color: var(--color-error-dot, #ef4444); }
</style>
