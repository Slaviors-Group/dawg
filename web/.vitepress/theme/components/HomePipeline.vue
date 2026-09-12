<script setup lang="ts">
import { useScrollReveal } from '../composables/useScrollReveal'

const { isRevealed, sectionRef } = useScrollReveal()

const features = [
  // Two large features for the top row
  {
    size: 'large',
    title: 'Capture every glitch as it happens',
    desc: 'One command opens a real browser and starts recording DOM mutations, network traffic, DB diffs, and structured logs — everything that leads to the bug.',
    placeholderType: 'window'
  },
  {
    size: 'large',
    title: 'Sanitize before anything leaves your machine',
    desc: 'PII and secrets are redacted locally with an OPA hard gate. If the policy fails, export is blocked outright — not just flagged.',
    placeholderType: 'code'
  },
  // Three small features for the bottom row
  {
    size: 'small',
    title: 'Replay it anywhere',
    desc: 'Time frozen, random seeded, network served from cassettes. Reproduces on any machine.',
    placeholderType: 'chart'
  },
  {
    size: 'small',
    title: 'Seamless Integrations',
    desc: 'Easily connect with your CI/CD pipelines, issue trackers, and team workflows.',
    placeholderType: 'network'
  },
  {
    size: 'small',
    title: 'Portable Artifacts',
    desc: 'Packaged into standard OCI artifacts, you can store and share them like Docker images.',
    placeholderType: 'devices'
  }
]
</script>

<template>
  <section class="dh-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-inner">
      <div class="dh-section-head">
        <h2 class="dh-h2">Powerful features to simplify your<br/>bug reproducing experience</h2>
      </div>

      <div class="dh-bento-grid">
        <div 
          v-for="(p, idx) in features" 
          :key="p.title" 
          class="dh-bento-card"
          :class="p.size === 'large' ? 'dh-bento-large' : 'dh-bento-small'"
          :style="{ transitionDelay: `${0.1 + idx * 0.25}s` }"
        >
          <!-- Placeholder Graphic Area -->
          <div class="dh-bento-img">
             <div class="dh-placeholder-content" :class="`dh-type-${p.placeholderType}`">
               <!-- Mock UI shapes for placeholder -->
               <div class="dh-mock-header">
                 <div class="dh-mock-dot"></div>
                 <div class="dh-mock-dot"></div>
                 <div class="dh-mock-dot"></div>
               </div>
               <div class="dh-mock-body"></div>
             </div>
          </div>
          
          <!-- Text Content -->
          <div class="dh-bento-text">
            <h3>{{ p.title }}</h3>
            <p>{{ p.desc }}</p>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dh-section { padding: 110px 24px; background: #ffffff; }
:global(:root.dark) .dh-section { background: var(--color-canvas); }

.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

.dh-section-head { text-align: center; margin-bottom: 64px; }
/* Matches "Title" sizing request */
.dh-h2 { 
  font-size: clamp(36px, 5vw, 64px); font-weight: 400; 
  line-height: 1.1; letter-spacing: -0.03em; margin: 0; 
  color: var(--color-text-primary); text-wrap: balance;
}

/* ── Bento Grid ── */
.dh-bento-grid {
  display: grid; 
  grid-template-columns: repeat(6, 1fr); 
  gap: 24px;
}
.dh-bento-large { grid-column: span 3; }
.dh-bento-small { grid-column: span 2; }

@media (max-width: 900px) { 
  .dh-bento-large, .dh-bento-small { grid-column: span 6; }
}

/* ── Bento Cards ── */
.dh-bento-card {
  background: #ffffff;
  border: 1px solid var(--color-border);
  border-radius: 24px;
  overflow: hidden;
  box-shadow: 0 4px 20px -4px rgba(0,0,0,0.03);
  display: flex;
  flex-direction: column;
}
:global(:root.dark) .dh-bento-card {
  background: var(--color-surface);
  box-shadow: 0 4px 20px -4px rgba(0,0,0,0.2);
}

/* ── Scroll Reveal Initial States ── */
.dh-h2 {
  opacity: 0;
  transform: translateY(30px) scale(0.98);
  filter: blur(8px);
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1), filter 1s cubic-bezier(0.16, 1, 0.3, 1);
}
.dh-bento-card {
  opacity: 0;
  transform: translateY(40px) scale(0.98);
  filter: blur(8px);
  /* Use separate transition properties so inline transitionDelay works correctly */
  transition: opacity 1s cubic-bezier(0.16, 1, 0.3, 1), transform 1s cubic-bezier(0.16, 1, 0.3, 1), filter 1s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Revealed States ── */
.dh-revealed .dh-h2 {
  opacity: 1;
  transform: translateY(0) scale(1);
  filter: blur(0);
}
.dh-revealed .dh-bento-card {
  opacity: 1;
  transform: translateY(0) scale(1);
  filter: blur(0);
}

.dh-bento-img {
  background: #f8f9fa;
  height: 280px;
  padding: 32px 32px 0 32px;
  display: flex;
  justify-content: center;
  align-items: flex-end;
  border-bottom: 1px solid var(--color-border);
}
:global(:root.dark) .dh-bento-img { background: var(--color-surface-hover); }

.dh-bento-text {
  padding: 32px;
  flex: 1;
}
.dh-bento-text h3 { 
  font-size: 20px; font-weight: 600; line-height: 1.3; 
  margin: 0 0 12px; color: var(--color-text-primary); 
}
.dh-bento-text p { 
  /* Matches "Desc" sizing request */
  font-size: 16px; color: var(--color-text-secondary); 
  line-height: 1.6; margin: 0; font-weight: 400;
}

/* ── Dummy Graphic Placeholders ── */
.dh-placeholder-content {
  width: 100%;
  height: 100%;
  background: #ffffff;
  border-radius: 16px 16px 0 0;
  box-shadow: 0 -4px 20px rgba(0,0,0,0.04);
  border: 1px solid var(--color-border);
  border-bottom: none;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
:global(:root.dark) .dh-placeholder-content { background: var(--color-surface); }

.dh-mock-header {
  height: 36px;
  border-bottom: 1px solid var(--color-border-subtle);
  display: flex;
  align-items: center;
  padding: 0 16px;
  gap: 6px;
  background: rgba(0,0,0,0.02);
}
.dh-mock-dot { width: 10px; height: 10px; border-radius: 50%; background: var(--color-border); }

.dh-mock-body {
  flex: 1;
  padding: 24px;
  background: repeating-linear-gradient(
    180deg,
    transparent,
    transparent 16px,
    var(--color-border-subtle) 16px,
    var(--color-border-subtle) 17px
  );
  opacity: 0.5;
}
</style>
