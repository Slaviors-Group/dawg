<template>
  <section class="dh-final">
    <div class="dh-inner">
      <div class="dh-final-box">
        <div class="dh-final-content">
          <h2 class="dh-final-title">Take Control of Your<br/>Policies Today</h2>
          <p class="dh-final-desc">Grab the desktop bundle or the CLI, capture once, and ship the fix with proof it's actually fixed.</p>
          <a class="dh-cta dh-cta-lg dh-cta-dark" href="/docs/installation">Get started Now &rarr;</a>
        </div>
        <div class="dh-final-graphic">
          <div class="dh-graphic-circles">
            <div class="dh-circle dh-c1"></div>
            <div class="dh-circle dh-c2"></div>
            <div class="dh-circle dh-c3"></div>
            <img src="/paw-dawg.svg" alt="DAWG" class="dh-graphic-logo" />
          </div>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

.dh-final { padding: 40px 24px 80px; background: #ffffff; }
:global(:root.dark) .dh-final { background: var(--color-canvas); }

.dh-final-box {
  background: #f8f9fa;
  border-radius: 32px;
  padding: 64px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 40px;
  position: relative;
  overflow: hidden;
}
:global(:root.dark) .dh-final-box {
  background: var(--color-surface);
}

.dh-final-content {
  flex: 1;
  max-width: 540px;
  z-index: 2;
}

/* Sizing matches Hero big title as requested */
.dh-final-title {
  font-size: clamp(42px, 5vw, 64px);
  font-weight: 400;
  line-height: 1.1;
  letter-spacing: -0.03em;
  margin: 0 0 24px;
  color: var(--color-text-primary);
  text-wrap: balance;
}

/* Sizing matches Hero/Tooling desc as requested */
.dh-final-desc {
  font-size: 18px;
  line-height: 1.6;
  color: var(--color-text-secondary);
  margin: 0 0 40px;
  font-weight: 400;
}

.dh-cta {
  display: inline-flex; align-items: center; justify-content: center;
  font-weight: 600; font-size: 15px; padding: 12px 24px; border-radius: var(--radius-full);
  text-decoration: none; border: none; white-space: nowrap; 
  transition: all var(--duration-fast);
}
.dh-cta-lg { padding: 16px 36px; font-size: 16px; }

.dh-cta-dark {
  background: #111111;
  color: #ffffff;
  box-shadow: 0 8px 24px rgba(0,0,0,0.15);
}
.dh-cta-dark:hover {
  background: #000000;
  transform: translateY(-2px);
  box-shadow: 0 12px 32px rgba(0,0,0,0.25);
}
:global(:root.dark) .dh-cta-dark { background: #ffffff; color: #000000; }
:global(:root.dark) .dh-cta-dark:hover { background: #f0f0f0; }

/* Abstract Graphic on the right */
.dh-final-graphic {
  flex: 1;
  display: flex;
  justify-content: flex-end;
  align-items: center;
  position: relative;
  z-index: 1;
}

.dh-graphic-circles {
  position: relative;
  width: 320px;
  height: 320px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dh-circle {
  position: absolute;
  border-radius: 50%;
  background: var(--color-brand);
}

.dh-c1 { width: 320px; height: 320px; opacity: 0.1; }
.dh-c2 { width: 220px; height: 220px; opacity: 0.2; }
.dh-c3 { width: 140px; height: 140px; opacity: 0.3; }

.dh-graphic-logo {
  position: relative;
  z-index: 2;
  width: 80px;
  height: 80px;
}

@media (max-width: 960px) {
  .dh-final-box {
    flex-direction: column;
    text-align: center;
    padding: 48px 32px;
  }
  .dh-final-content {
    margin: 0 auto;
    display: flex;
    flex-direction: column;
    align-items: center;
  }
  .dh-final-graphic {
    justify-content: center;
    width: 100%;
    margin-top: 24px;
  }
}
</style>
