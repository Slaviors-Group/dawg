<template>
  <footer class="dh-footer">
    <div class="dh-inner">
      <div class="dh-foot-top">
        <div class="dh-foot-brand">
          <a class="dh-brand" href="/">
            <img src="/paw-dawg.svg" alt="DAWG" />
            <span>DAWG</span>
          </a>
        </div>
        
        <div class="dh-foot-links">
          <div class="dh-foot-col">
            <p class="dh-foot-head">Product</p>
            <a href="/docs/getting-started">Getting Started</a>
            <a href="/docs/installation">Installation</a>
            <a href="/docs/cli-reference">CLI Reference</a>
          </div>
          <div class="dh-foot-col">
            <p class="dh-foot-head">Resources</p>
            <a href="/docs/architecture">Architecture</a>
            <a href="/docs/sanitizer-policy">Sanitizer Policy</a>
            <a href="/docs/changelog">Changelog</a>
          </div>
          <div class="dh-foot-col">
            <p class="dh-foot-head">Project</p>
            <a href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer">GitHub</a>
            <a href="https://github.com/Slaviors-Group/dawg/releases" target="_blank" rel="noopener noreferrer">Releases</a>
            <a href="/docs/contributing">Contributing</a>
          </div>
        </div>

        <div class="dh-foot-social">
          <!-- Phospor icons for social links -->
          <a href="https://github.com/Slaviors-Group/dawg" target="_blank" aria-label="GitHub"><i class="ph ph-github-logo"></i></a>
          <a href="#" aria-label="Discord"><i class="ph ph-discord-logo"></i></a>
          <a href="#" aria-label="X / Twitter"><i class="ph ph-x-logo"></i></a>
        </div>
      </div>
    </div>

    <!-- Huge bottom DAWG text with gradient and mask -->
    <div class="dh-foot-huge-wrap">
      <div class="dh-huge-glow"></div>
      <div class="dh-huge-text">DAWG</div>
    </div>
  </footer>
</template>

<style scoped>
.dh-inner { max-width: 1200px; margin: 0 auto; padding: 0 24px; position: relative; z-index: 10; }

.dh-footer { 
  padding: 64px 0 0; 
  background: #ffffff; 
  position: relative;
  overflow: hidden;
}
:global(:root.dark) .dh-footer { background: var(--color-canvas); }

.dh-foot-top { 
  display: flex; 
  justify-content: space-between; 
  align-items: flex-start;
  flex-wrap: wrap;
  gap: 40px;
  padding-bottom: 80px;
}

.dh-brand { display: inline-flex; align-items: center; gap: 10px; text-decoration: none; }
.dh-brand img { width: 28px; height: 28px; }
.dh-brand span { color: var(--color-text-primary); font-weight: 700; font-size: 20px; letter-spacing: -0.02em; }

.dh-foot-links {
  display: flex;
  gap: 80px;
  flex-wrap: wrap;
}

.dh-foot-head { 
  font-size: 14px; 
  font-weight: 600; 
  color: var(--color-text-primary); 
  margin-bottom: 20px; 
}
.dh-foot-col { display: flex; flex-direction: column; gap: 16px; }
.dh-foot-col a { 
  font-size: 14px; 
  color: var(--color-text-secondary); 
  text-decoration: none; 
  font-weight: 400; 
  transition: color var(--duration-fast); 
}
.dh-foot-col a:hover { color: var(--color-text-primary); }

.dh-foot-social {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.dh-foot-social a {
  color: var(--color-text-secondary);
  font-size: 20px;
  transition: color var(--duration-fast);
}
.dh-foot-social a:hover {
  color: var(--color-text-primary);
}

@media (max-width: 860px) {
  .dh-foot-top { flex-direction: column; }
  .dh-foot-links { gap: 40px; }
  .dh-foot-social { flex-direction: row; }
}

/* ── Huge Gradient Text ── */
.dh-foot-huge-wrap {
  position: relative;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: flex-end;
  /* Mask out the bottom so it fades perfectly into the edge */
  mask-image: linear-gradient(to bottom, black 40%, transparent 100%);
  -webkit-mask-image: linear-gradient(to bottom, black 40%, transparent 100%);
  padding-top: 40px;
}

.dh-huge-glow {
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 80%;
  height: 200px;
  background: var(--color-brand);
  filter: blur(80px);
  opacity: 0.3;
  z-index: 1;
}

.dh-huge-text {
  position: relative;
  z-index: 2;
  font-size: 26vw;
  font-weight: 800;
  line-height: 0.75;
  letter-spacing: -0.04em;
  color: var(--color-canvas);
  /* Create the soft gradient effect inside the text */
  background: linear-gradient(to bottom, rgba(213, 196, 250, 0.9), rgba(213, 196, 250, 0.1));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  color: transparent;
  text-align: center;
  pointer-events: none;
  user-select: none;
}
:global(:root.dark) .dh-huge-text {
  background: linear-gradient(to bottom, rgba(213, 196, 250, 0.3), rgba(213, 196, 250, 0.02));
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}
</style>
