<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

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
  <section class="dh-section" :class="{ 'dh-revealed': isRevealed }" ref="sectionRef">
    <div class="dh-inner">
      <div class="dh-head">
        <h2 class="dh-h2">Everything You Need.<br/>Nothing You Don’t.</h2>
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
        <div class="dh-masonry-col" style="transition-delay: 0.2s">
          <div class="dh-res-graphic-card">
            <!-- CSS Mock Phone/Dashboard -->
            <div class="dh-mock-phone">
              <div class="dh-phone-notch"></div>
              <div class="dh-phone-screen">
                <div class="dh-phone-header">
                  <div class="dh-avatar"></div>
                  <div class="dh-header-text"></div>
                </div>
                <div class="dh-phone-title"></div>
                <div class="dh-phone-subtitle"></div>
                <div class="dh-phone-pills">
                  <div class="dh-pill-active"></div>
                  <div class="dh-pill"></div>
                  <div class="dh-pill"></div>
                </div>
                <div class="dh-phone-card"></div>
              </div>
            </div>
          </div>
          <a href="/docs/architecture" class="dh-res-card">
            <h3>Architecture</h3>
            <p>How the six-stage pipeline fits together. Understand the magic under the hood.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
        </div>

        <!-- Right Column -->
        <div class="dh-masonry-col" style="transition-delay: 0.3s">
          <a href="/docs/sanitizer-policy" class="dh-res-card">
            <h3>Sanitizer Policy</h3>
            <p>Rego gates, redaction rules, synthetic data. Keep your sensitive data strictly local.</p>
            <span class="dh-learn-more">Learn More &rarr;</span>
          </a>
          <a href="/docs/contributing" class="dh-res-card">
            <h3>Contributing</h3>
            <p>Dev setup, standards, and PR workflow. Help us build the ultimate bug catcher.</p>
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
  transform: translateY(20px);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}
.dh-masonry-col {
  opacity: 0;
  transform: translateY(30px);
  transition: all 0.8s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Revealed States ── */
.dh-revealed .dh-head { opacity: 1; transform: translateY(0); }
.dh-revealed .dh-masonry-col { opacity: 1; transform: translateY(0); }


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
  display: flex;
  justify-content: center;
  align-items: flex-end;
  overflow: hidden;
  position: relative;
}

/* CSS Mock Phone */
.dh-mock-phone {
  width: 220px;
  height: 340px; /* Cut off at bottom */
  background: #ffffff;
  border-radius: 36px 36px 0 0;
  box-shadow: 0 20px 40px rgba(0,0,0,0.2);
  border: 6px solid #111111;
  border-bottom: none;
  position: relative;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
:global(:root.dark) .dh-mock-phone { background: var(--color-canvas); }

.dh-phone-notch {
  position: absolute;
  top: 0; left: 50%;
  transform: translateX(-50%);
  width: 80px; height: 24px;
  background: #111;
  border-radius: 0 0 12px 12px;
  z-index: 2;
}

.dh-phone-screen {
  padding: 40px 16px 16px;
  flex: 1;
  background: linear-gradient(to bottom, rgba(213, 196, 250, 0.1), transparent);
}

.dh-phone-header { display: flex; align-items: center; gap: 8px; margin-bottom: 24px; }
.dh-avatar { width: 32px; height: 32px; border-radius: 50%; background: rgba(0,0,0,0.08); }
.dh-header-text { width: 60px; height: 8px; border-radius: 4px; background: rgba(0,0,0,0.08); }

.dh-phone-title { width: 80%; height: 16px; border-radius: 4px; background: rgba(0,0,0,0.8); margin-bottom: 8px; }
.dh-phone-subtitle { width: 50%; height: 16px; border-radius: 4px; background: rgba(0,0,0,0.8); margin-bottom: 24px; }
:global(:root.dark) .dh-phone-title, :global(:root.dark) .dh-phone-subtitle { background: rgba(255,255,255,0.8); }

.dh-phone-pills { display: flex; gap: 6px; margin-bottom: 24px; }
.dh-pill-active { width: 60px; height: 24px; border-radius: 12px; background: var(--color-brand); }
.dh-pill { width: 40px; height: 24px; border-radius: 12px; background: rgba(0,0,0,0.05); }
:global(:root.dark) .dh-pill { background: rgba(255,255,255,0.1); }

.dh-phone-card {
  width: 100%;
  height: 120px;
  border-radius: 16px;
  background: rgba(213, 196, 250, 0.3);
}

</style>
