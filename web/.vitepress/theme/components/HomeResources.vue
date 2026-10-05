<script setup lang="ts">
import { useScrollReveal } from '../composables/useScrollReveal'

const { isRevealed, sectionRef } = useScrollReveal()

// Phosphor Icons PawPrint, fill weight: https://github.com/phosphor-icons/core/blob/main/assets/fill/paw-print-fill.svg
const pawPath = 'M240,108a28,28,0,1,1-28-28A28,28,0,0,1,240,108ZM72,108a28,28,0,1,0-28,28A28,28,0,0,0,72,108ZM92,88A28,28,0,1,0,64,60,28,28,0,0,0,92,88Zm72,0a28,28,0,1,0-28-28A28,28,0,0,0,164,88Zm23.12,60.86a35.3,35.3,0,0,1-16.87-21.14,44,44,0,0,0-84.5,0A35.25,35.25,0,0,1,69,148.82,40,40,0,0,0,88,224a39.48,39.48,0,0,0,15.52-3.13,64.09,64.09,0,0,1,48.87,0,40,40,0,0,0,34.73-72Z'

const pawSteps = [
  { x: 18, y: 82, tilt: -20 },
  { x: 33, y: 76, tilt: 19 },
  { x: 32, y: 64, tilt: -18 },
  { x: 47, y: 58, tilt: 18 },
  { x: 47, y: 46, tilt: -17 },
  { x: 62, y: 40, tilt: 17 },
  { x: 62, y: 28, tilt: -16 },
  { x: 77, y: 22, tilt: 16 }
]
</script>

<template>
  <section class="dh-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-inner">
      <div class="dh-head">
        <h2 class="dh-h2">
          Everything You Need.
          <span class="dh-heading-line">Nothing You Don’t.</span>
        </h2>
        <p class="dh-sub">Comprehensive tools designed with simplicity and security in mind.</p>
      </div>

      <div class="dh-masonry">
        <!-- Left Column -->
        <div class="dh-masonry-col" style="transition-delay: 0.1s">
          <a href="/docs/getting-started" class="dh-res-card">
            <h3>Getting Started</h3>
            <p>Capture your first glitch in under five minutes. Quick installation and setup guide.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
          <a href="/docs/cli-reference" class="dh-res-card">
            <h3>CLI Reference</h3>
            <p>Every subcommand, flag, and exit code. Complete documentation for terminal ninjas.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
        </div>

        <!-- Center Column -->
        <div class="dh-masonry-col" style="transition-delay: 0.35s">
          <div class="dh-res-graphic-card" role="img" aria-label="DAWG paw prints lead to a captured artifact ready for replay">
            <div class="dh-paw-trail" aria-hidden="true">
              <span
                v-for="(step, index) in pawSteps"
                :key="index"
                class="dh-paw-step"
                :class="index % 2 === 0 ? 'dh-paw-step--left' : 'dh-paw-step--right'"
                :style="{
                  left: `${step.x}%`,
                  top: `${step.y}%`,
                  '--paw-tilt': `${step.tilt}deg`,
                  '--paw-delay': `${index * 0.38}s`
                }"
              >
                <svg viewBox="0 0 256 256" fill="currentColor" aria-hidden="true" focusable="false">
                  <path :d="pawPath" />
                </svg>
              </span>
            </div>
            <div class="dh-artifact-reveal" aria-hidden="true">
              <img src="/assets/images/resources-artifact.svg" alt="" width="360" height="400" loading="lazy" decoding="async" />
            </div>
          </div>
          <a href="/docs/architecture" class="dh-res-card">
            <h3>Architecture</h3>
            <p>How the six-stage pipeline fits together. Understand the magic under the hood.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
        </div>

        <!-- Right Column -->
        <div class="dh-masonry-col" style="transition-delay: 0.6s">
          <a href="/docs/sanitizer-policy" class="dh-res-card">
            <h3>Sanitizer Policy</h3>
            <p>Understand the built-in redaction rules, deterministic replacements, OPA gate, reports, and security boundaries.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
          <a href="/docs/contributing" class="dh-res-card">
            <h3>Contributing</h3>
            <p>Development setup, validation, versioning, and feedback channels.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dh-section { padding: 110px 24px; background: #ffffff; }
:global(:root.dark) .dh-section { background: var(--color-canvas); }

.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

.dh-head { text-align: center; margin-bottom: 64px; }
.dh-heading-line { display: block; }
/* Matches "Title" sizing request */
.dh-h2 { 
  font-size: clamp(36px, 5vw, 64px); font-weight: 400; 
  line-height: 1.1; letter-spacing: -0.03em; margin: 0 0 16px; 
  color: var(--color-text-primary); text-wrap: balance;
}
/* Matches "Desc" sizing request */
.dh-sub { 
  font-size: 18px; line-height: 1.6; color: var(--color-text-secondary); 
  margin: 0 auto; font-weight: 400; max-width: 600px; text-wrap: balance;
}

/* ── Scroll Reveal Initial States ── */
.dh-head {
  opacity: 0;
  transform: translateY(30px) scale(0.98);
  filter: blur(8px);
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1), filter 1s cubic-bezier(0.16, 1, 0.3, 1);
}
.dh-masonry-col {
  opacity: 0;
  transform: translateY(40px) scale(0.98);
  filter: blur(8px);
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1), filter 1s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Revealed States ── */
.dh-revealed .dh-head { opacity: 1; transform: translateY(0) scale(1); filter: blur(0); }
.dh-revealed .dh-masonry-col { opacity: 1; transform: translateY(0) scale(1); filter: blur(0); }


/* ── Masonry Grid ── */
.dh-masonry {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 24px;
}
@media (max-width: 900px) {
  .dh-masonry { grid-template-columns: 1fr; }
}

.dh-masonry-col {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* ── Text Cards ── */
.dh-res-card {
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
  border: 1px solid var(--color-border);
  border-radius: 20px;
  padding: 32px;
  text-decoration: none;
  color: inherit;
  transition: all var(--duration-fast);
  box-shadow: 0 4px 20px -4px rgba(0,0,0,0.02);
}
:global(:root.dark) .dh-res-card {
  background: var(--color-surface);
  box-shadow: 0 4px 20px -4px rgba(0,0,0,0.2);
}
.dh-res-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 30px rgba(0,0,0,0.06);
  border-color: var(--color-brand);
}

.dh-res-card h3 {
  font-size: 22px;
  font-weight: 600;
  line-height: 1.3;
  margin: 0 0 16px;
  color: var(--color-text-primary);
}

.dh-res-card p {
  font-size: 16px;
  line-height: 1.6;
  color: var(--color-text-secondary);
  margin: 0 0 24px;
  flex: 1;
}

.dh-learn-more {
  font-size: 15px;
  font-weight: 600;
  color: var(--color-text-secondary);
  transition: color var(--duration-fast);
}
.dh-res-card:hover .dh-learn-more {
  color: var(--color-brand);
}

/* ── Graphic Card ── */
.dh-res-graphic-card {
  background: var(--color-brand);
  border-radius: 20px;
  height: 400px;
  overflow: hidden;
  position: relative;
}

.dh-paw-trail {
  position: absolute;
  inset: 0;
  animation: dh-paw-trail-phase 8s linear infinite;
}

.dh-paw-step {
  position: absolute;
  width: clamp(38px, 4.5vw, 54px);
  aspect-ratio: 1;
  transform: translate(-50%, -50%) rotate(var(--paw-tilt));
}

.dh-paw-step--left { color: #5036b9; }
.dh-paw-step--right { color: #7656dc; }

.dh-paw-step svg {
  display: block;
  width: 100%;
  height: 100%;
  opacity: 0;
  filter: drop-shadow(0 5px 4px rgb(65 42 136 / 14%));
  animation: dh-paw-walk 8s ease-in-out var(--paw-delay) infinite;
}

@keyframes dh-paw-walk {
  0%, 3% { opacity: 0; transform: translateY(12px) scale(.72); }
  7%, 47% { opacity: .95; transform: translateY(0) scale(1); }
  55%, 100% { opacity: 0; transform: translateY(-6px) scale(.93); }
}

@keyframes dh-paw-trail-phase {
  0%, 49% { opacity: 1; }
  56%, 97% { opacity: 0; }
  100% { opacity: 1; }
}

.dh-artifact-reveal {
  position: absolute;
  inset: 0;
  display: grid;
  place-items: center;
  opacity: 0;
  animation: dh-artifact-reveal 8s ease-in-out infinite;
}

.dh-artifact-reveal img {
  display: block;
  width: min(90%, 360px);
  height: 100%;
  object-fit: contain;
}

@keyframes dh-artifact-reveal {
  0%, 51% {
    opacity: 0;
    transform: translateY(18px) scale(.9);
    clip-path: inset(100% 0 0 0);
  }
  60%, 88% {
    opacity: 1;
    transform: translateY(0) scale(1);
    clip-path: inset(0 0 0 0);
  }
  97%, 100% {
    opacity: 0;
    transform: translateY(-8px) scale(1.02);
    clip-path: inset(0 0 0 0);
  }
}

:global(:root.dark) .dh-res-graphic-card {
  background: #4d3a82;
}
:global(:root.dark) .dh-paw-step--left { color: #d5c4fa; }
:global(:root.dark) .dh-paw-step--right { color: #ac91ec; }

@media (prefers-reduced-motion: reduce) {
  .dh-paw-trail { display: none; }
  .dh-artifact-reveal {
    animation: none;
    opacity: 1;
  }
}

</style>
