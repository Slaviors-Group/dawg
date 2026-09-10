<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
// Technologies positioned on three concentric half-circle orbits
// Each entry: { name, logo (SVG URL from CDN), orbit (1=inner, 2=mid, 3=outer), angleDeg }
// angleDeg: 0 = left horizontal, 90 = top, 180 = right horizontal

const orbitTechs = [
  // Inner orbit (orbit 1) — 3 items
  { name: 'Docker',     logo: 'https://cdn.simpleicons.org/docker',         orbit: 1, angle: 60 },
  { name: 'React',      logo: 'https://cdn.simpleicons.org/react',          orbit: 1, angle: 90 },
  { name: 'OPA',        logo: 'https://cdn.simpleicons.org/openapiinitiative', orbit: 1, angle: 165 },

  // Middle orbit (orbit 2) — 4 items
  { name: 'Go',         logo: 'https://cdn.simpleicons.org/go',             orbit: 2, angle: 45 },
  { name: 'Playwright', logo: 'https://iconlogovector.com/uploads/images/2024/12/lg-676c8ff26c74d-Playwright.webp', orbit: 2, angle: 75 },
  { name: 'Tauri',      logo: 'https://cdn.simpleicons.org/tauri',          orbit: 2, angle: 105 },
  { name: 'mitmproxy',  logo: 'https://avatars.githubusercontent.com/u/4652787?s=280&v=4', orbit: 2, angle: 135 },

  // Outer orbit (orbit 3) — 4 items
  { name: 'rrweb',      logo: 'https://raw.githubusercontent.com/rrweb-io/rrweb/refs/heads/main/packages/web-extension/src/public/icon128.png', orbit: 3, angle: 12 },
  { name: 'OCI',        logo: 'https://cdn.simpleicons.org/opencontainersinitiative', orbit: 3, angle: 65 },
  { name: 'ORAS',       logo: 'https://cdn.simpleicons.org/cncf',           orbit: 3, angle: 100 },
  { name: 'GitHub',     logo: 'https://cdn.simpleicons.org/github',         orbit: 3, angle: 150 },
]

// Orbit radii as % of container width (half-circle, rendered as SVG arcs)
const orbits = [220, 340, 460] // px radii for inner, mid, outer

function getPos(angleDeg: number, radius: number) {
  // 90deg = top, 0=left, 180=right — map to SVG coords
  // center is at bottom center of a 960x480 viewBox
  const cx = 480
  const cy = 480
  const rad = (angleDeg * Math.PI) / 180
  return {
    x: cx - Math.cos(rad) * radius,
    y: cy - Math.sin(rad) * radius,
  }
}

const isRevealed = ref(false)
const sectionRef = ref<HTMLElement | null>(null)
let observer: IntersectionObserver | null = null

onMounted(() => {
  observer = new IntersectionObserver(
    ([entry]) => {
      if (entry.isIntersecting) {
        isRevealed.value = true
        observer?.disconnect()
      }
    },
    { threshold: 0.15 }
  )
  if (sectionRef.value) observer.observe(sectionRef.value)
})

onUnmounted(() => {
  observer?.disconnect()
})
</script>

<template>
  <section class="dh-orbit-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-orbit-head">

      <h2 class="dh-h2">Built on a rock-solid open stack</h2>
      <p class="dh-sub">Every layer is an auditable, battle-tested open-source project.</p>
    </div>

    <div class="dh-orbit-wrap">
      <svg
        viewBox="0 0 960 480"
        class="dh-orbit-svg"
        aria-hidden="true"
      >
        <!-- Backgrounds and Arcs -->
        <g class="dh-orbit-bg">
          <!-- Filled half-circle bands (outermost to innermost, darkening toward center) -->
          <!-- Band 3: outer fill -->
          <path
            d="M 20 480 A 460 460 0 0 1 940 480 Z"
            fill="rgba(213, 196, 250, 0.10)"
          />
          <!-- Band 2: mid fill -->
          <path
            d="M 140 480 A 340 340 0 0 1 820 480 Z"
            fill="rgba(213, 196, 250, 0.18)"
          />
          <!-- Band 1: inner fill -->
          <path
            d="M 260 480 A 220 220 0 0 1 700 480 Z"
            fill="rgba(213, 196, 250, 0.30)"
          />

          <!-- Half-circle arcs (dashed strokes) -->
          <path v-for="(r, i) in orbits" :key="i"
            :d="`M ${480 - r} 480 A ${r} ${r} 0 0 1 ${480 + r} 480`"
            fill="none"
            stroke="rgba(213, 196, 250, 0.8)"
            :stroke-width="i === 1 ? 1.5 : 1"
            stroke-dasharray="6 4"
          />
        </g>

        <!-- Logo nodes -->
        <g v-for="(tech, idx) in orbitTechs" :key="tech.name"
           class="dh-orbit-node"
           :style="{ transitionDelay: `${0.3 + idx * 0.08}s` }">
          <!-- White circle backdrop -->
          <circle
            :cx="getPos(tech.angle, orbits[tech.orbit - 1]).x"
            :cy="getPos(tech.angle, orbits[tech.orbit - 1]).y"
            r="26"
            fill="white"
            stroke="rgba(213, 196, 250, 0.8)"
            stroke-width="1.5"
            class="dh-orbit-chip-bg"
          />
          <image
            :href="tech.logo"
            :x="getPos(tech.angle, orbits[tech.orbit - 1]).x - 14"
            :y="getPos(tech.angle, orbits[tech.orbit - 1]).y - 14"
            width="28"
            height="28"
          />
          <text
            :x="getPos(tech.angle, orbits[tech.orbit - 1]).x"
            :y="getPos(tech.angle, orbits[tech.orbit - 1]).y + 40"
            text-anchor="middle"
            class="dh-orbit-label"
          >{{ tech.name }}</text>
        </g>

        <!-- Center DAWG node -->
        <image href="/paw-dawg.svg" x="444" y="372" width="72" height="72" class="dh-center-node" />
      </svg>
    </div>
  </section>
</template>

<style scoped>
.dh-orbit-section {
  padding: 80px 24px 0;
  text-align: center;
  background: #ffffff;
  overflow: hidden;
}
:global(:root.dark) .dh-orbit-section {
  background: var(--color-canvas);
}

/* ── Scroll Reveal Initial States ── */
.dh-orbit-head {
  max-width: 600px;
  margin: 0 auto 32px;
  opacity: 0;
  transform: translateY(20px);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}
.dh-orbit-svg {
  width: 100%;
  height: auto;
  display: block;
}
.dh-orbit-bg, .dh-center-node {
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1) 0.1s;
}
.dh-orbit-node {
  opacity: 0;
  /* In SVG, using CSS transform scale can be unpredictable, so we use translateY */
  transform: translateY(15px);
  transition: opacity 0.6s ease, transform 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* ── Revealed States ── */
.dh-revealed .dh-orbit-head {
  opacity: 1;
  transform: translateY(0);
}
.dh-revealed .dh-orbit-bg,
.dh-revealed .dh-center-node {
  opacity: 1;
  transform: translateY(0);
}
.dh-revealed .dh-orbit-node {
  opacity: 1;
  transform: translateY(0);
}

.dh-h2 {
  font-size: clamp(36px, 5vw, 64px); font-weight: 400;
  line-height: 1.1; letter-spacing: -0.03em; margin: 0 0 16px;
  color: var(--color-text-primary);
}
.dh-sub {
  font-size: 18px; color: var(--color-text-secondary); margin: 0; line-height: 1.6;
}

.dh-orbit-wrap {
  max-width: 960px;
  margin: 0 auto;
  /* clip so only the top half of the circle arcs is visible */
  overflow: hidden;
}

.dh-orbit-chip-bg {
  filter: drop-shadow(0 2px 8px rgba(0,0,0,0.08));
}

.dh-orbit-label {
  font-family: var(--font-sans);
  font-size: 11px;
  fill: var(--color-text-secondary);
  font-weight: 500;
}
</style>
