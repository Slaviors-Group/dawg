<script setup lang="ts">
import { useScrollReveal } from '../composables/useScrollReveal'

const { isRevealed, sectionRef } = useScrollReveal()

const features = [
  // Two large features for the top row
  {
    size: 'large',
    title: 'Capture from the browser you already use',
    desc: 'The desktop app and Chromium extension record rrweb events, user actions, and frontend request metadata from the tab where the bug occurs.',
    image: '/assets/images/capture-browser-workflow.svg',
    imageAlt: 'Browser capture workflow leading to a DAWG artifact',
    imageWidth: 550,
    imageHeight: 370
  },
  {
    size: 'large',
    title: 'Sanitize before packaging',
    desc: 'Built-in heuristics replace common PII and secrets, then an OPA policy gates the sanitized capture before DAWG creates the artifact.',
    image: '/assets/images/sanitize-before-packaging.svg',
    imageAlt: 'Sensitive data passing through a sanitizer and policy check into an artifact',
    imageWidth: 600,
    imageHeight: 400
  },
  // Three small features for the bottom row
  {
    size: 'small',
    title: 'Replay with recorded context',
    desc: 'Render the captured rrweb timeline in Chromium and inspect replay diagnostics without needing the original application.',
    image: '/assets/images/replay-recorded-context.svg',
    imageAlt: 'Recorded browser timeline with playback controls and diagnostics',
    imageWidth: 600,
    imageHeight: 400
  },
  {
    size: 'small',
    title: 'Persistent Artifact Catalog',
    desc: 'Keep captured, imported, and discovered legacy artifacts available across desktop restarts.',
    image: '/assets/images/persistent-artifact-catalog.svg',
    imageAlt: 'Searchable catalog containing captured, imported, and legacy artifacts',
    imageWidth: 600,
    imageHeight: 400
  },
  {
    size: 'small',
    title: 'Portable .dawg Archives',
    desc: 'Export validated OCI layouts as .dawg files, import them safely, or exchange artifacts through an OCI registry.',
    image: '/assets/images/portable-dawg-archives.svg',
    imageAlt: 'A portable DAWG archive shared from QA to development',
    imageWidth: 600,
    imageHeight: 400
  }
]
</script>

<template>
  <section class="dh-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-inner">
      <div class="dh-section-head">
        <h2 class="dh-h2">
          Powerful features to simplify your
          <span class="dh-heading-line">bug-reproduction experience</span>
        </h2>
      </div>

      <div class="dh-bento-grid">
        <div 
          v-for="(p, idx) in features" 
          :key="p.title" 
          class="dh-bento-card"
          :class="p.size === 'large' ? 'dh-bento-large' : 'dh-bento-small'"
          :style="{ transitionDelay: `${0.1 + idx * 0.25}s` }"
        >
          <div class="dh-bento-img">
            <img
              :src="p.image"
              :alt="p.imageAlt"
              :width="p.imageWidth"
              :height="p.imageHeight"
              loading="lazy"
              decoding="async"
            />
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
.dh-heading-line { display: block; }
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
  background: #f5f1ff;
  height: 280px;
  display: flex;
  justify-content: center;
  align-items: center;
  border-bottom: 1px solid var(--color-border);
  overflow: hidden;
}
.dh-bento-img img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
  pointer-events: none;
  -webkit-user-drag: none;
  user-select: none;
}

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

</style>
