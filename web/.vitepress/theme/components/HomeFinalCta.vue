<template>
  <section class="dh-final">
    <div class="dh-inner">
      <div class="dh-final-box">
        <div class="dh-glow-orb dh-orb-center"></div>
        <h2>Take Control of Your<br/>Policies Today</h2>
        <p>Grab the desktop bundle or the CLI, capture once, and ship the fix with proof it's actually fixed.</p>
        <div class="dh-cta-row">
          <a class="dh-cta dh-cta-lg dh-cta-white" href="/docs/installation">Install DAWG</a>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

/* Decorative Orbs */
.dh-glow-orb {
  position: absolute; width: 600px; height: 600px; border-radius: 50%;
  filter: blur(100px); opacity: 0.15; z-index: 1; pointer-events: none;
}
.dh-orb-center { background: var(--color-brand); top: 50%; left: 50%; transform: translate(-50%, -50%); width: 800px; height: 800px; opacity: 0.1; }

.dh-cta {
  display: inline-block; background: var(--color-brand); color: #000;
  font-weight: 600; font-size: 15px; padding: 12px 24px; border-radius: var(--radius-full);
  text-decoration: none; border: none; white-space: nowrap; 
  transition: all var(--duration-fast); box-shadow: 0 4px 14px 0 rgba(107, 33, 217, 0.39);
}
.dh-cta:hover { background: var(--color-brand); filter: brightness(0.9); transform: translateY(-2px); box-shadow: 0 6px 20px rgba(107, 33, 217, 0.23); }
.dh-cta-lg { padding: 16px 36px; font-size: 16px; }

.dh-cta-white {
  background: #000; color: #fff; box-shadow: 0 4px 14px 0 rgba(0,0,0,0.1);
}
.dh-cta-white:hover { background: #333; color: #fff; }

/* ── Final CTA Box ── */
.dh-final { padding: 60px 24px 120px; }
.dh-final-box {
  background: var(--color-brand);
  border-radius: var(--radius-xl); padding: 80px 40px; text-align: center;
  color: #000; position: relative; overflow: hidden;
  box-shadow: 0 24px 60px -12px rgba(107, 33, 217, 0.4);
}
.dh-final-box h2 { font-size: clamp(32px, 4vw, 44px); font-weight: 700; line-height: 1.2; margin: 0 auto 16px; position: relative; z-index: 2; color: #000;}
.dh-final-box p { font-size: 18px; color: rgba(0,0,0,0.8); max-width: 500px; margin: 0 auto 40px; line-height: 1.6; position: relative; z-index: 2; }
.dh-cta-row { position: relative; z-index: 2; }
</style>
