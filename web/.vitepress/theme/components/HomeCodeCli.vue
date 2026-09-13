<script setup lang="ts">
import { useScrollReveal } from '../composables/useScrollReveal'

const { isRevealed, sectionRef } = useScrollReveal()

const features = [
  {
    title: 'Local Capture & Replay',
    desc: 'Inspect what was captured with no replay needed. Push to your team registry instantly.',
    mockType: 'table'
  },
  {
    title: 'CI-Ready Exit Codes',
    desc: 'Automate your workflows with pass/fail gates that integrate perfectly into your CI pipelines.',
    mockType: 'chart'
  }
]
</script>

<template>
  <section class="dh-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-inner">
      <div class="dh-head">
        <h2 class="dh-h2">Ship with tooling that does<br/>what you expect.</h2>
        <p class="dh-sub">Single static binary — no daemon, no server. Human-readable output with JSON underneath.</p>
      </div>

      <!-- Two Columns -->
      <div class="dh-grid-2">
        <div v-for="(feat, idx) in features" :key="feat.title" class="dh-feature-card" :style="{ transitionDelay: `${0.1 + idx * 0.25}s` }">
          <div class="dh-mock-wrap">
            <div class="dh-mock-glow"></div>
            <div class="dh-mock-ui">
              <div class="dh-mock-header"></div>
              <div class="dh-mock-lines">
                <div class="dh-mock-line" style="width: 80%"></div>
                <div class="dh-mock-line" style="width: 60%"></div>
                <div class="dh-mock-line" style="width: 90%"></div>
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
             <div class="dh-mock-header"></div>
             <div class="dh-mock-columns">
                <div class="dh-mock-bar" style="height: 60%"></div>
                <div class="dh-mock-bar" style="height: 80%"></div>
                <div class="dh-mock-bar" style="height: 40%"></div>
                <div class="dh-mock-bar" style="height: 90%"></div>
                <div class="dh-mock-bar" style="height: 50%"></div>
             </div>
          </div>
        </div>
        <div class="dh-feature-text">
          <h3>Verify Fixes Locally</h3>
          <p>Capture the bug, then stop when it appears. Verify the fix on your local branch against the actual recorded failure.</p>
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

.dh-mock-lines {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.dh-mock-line {
  height: 16px;
  background: rgba(0,0,0,0.04);
  border-radius: 4px;
}
:global(:root.dark) .dh-mock-lines .dh-mock-line { background: rgba(255,255,255,0.04); }

.dh-mock-columns {
  display: flex;
  align-items: flex-end;
  gap: 16px;
  flex: 1;
  padding-top: 24px;
}
.dh-mock-bar {
  flex: 1;
  background: var(--color-brand);
  opacity: 0.8;
  border-radius: 4px 4px 0 0;
}
</style>
