
<template>
  <footer class="dh-footer">
    <div class="dh-inner">
      <div class="dh-foot-top">
        <div class="dh-foot-brand">
          <a class="dh-brand" href="/">
            <img src="/paw-dawg.svg" alt="" width="28" height="28" decoding="async" />
            <span>DAWG</span>
          </a>
        </div>
        
        <nav class="dh-foot-links" aria-label="Footer navigation">
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
            <a href="https://github.com/Slaviors-Group/dawg/issues/new/choose" target="_blank" rel="noopener noreferrer">Report a bug</a>
            <a href="https://github.com/Slaviors-Group/dawg/releases" target="_blank" rel="noopener noreferrer">Releases</a>
            <a href="https://github.com/sponsors/Slaviors-Group" target="_blank" rel="noopener noreferrer">Sponsor DAWG</a>
            <a href="/docs/contributing">Contributing</a>
          </div>
          <div class="dh-foot-col">
            <p class="dh-foot-head">Legal</p>
            <a href="/docs/privacy-policy">Privacy Policy</a>
            <a href="/docs/terms-of-service">Terms of Service</a>
            <a href="https://github.com/Slaviors-Group/dawg/blob/main/LICENSE" target="_blank" rel="noopener noreferrer">License (GPL-3.0)</a>
          </div>
        </nav>

        <div class="dh-foot-social">
          <a href="https://github.com/Slaviors-Group/dawg" target="_blank" rel="noopener noreferrer" aria-label="GitHub repository">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path fill="currentColor" d="M12 .7a11.5 11.5 0 0 0-3.64 22.4c.58.1.79-.25.79-.56v-2.24c-3.24.7-3.92-1.38-3.92-1.38-.53-1.35-1.3-1.71-1.3-1.71-1.06-.73.08-.71.08-.71 1.17.08 1.79 1.2 1.79 1.2 1.04 1.79 2.73 1.27 3.4.97.1-.76.41-1.27.74-1.56-2.58-.29-5.3-1.29-5.3-5.69 0-1.26.45-2.29 1.2-3.09-.12-.29-.52-1.47.11-3.05 0 0 .98-.31 3.16 1.18a10.98 10.98 0 0 1 5.75 0c2.19-1.49 3.16-1.18 3.16-1.18.63 1.58.23 2.76.11 3.05.75.8 1.2 1.83 1.2 3.09 0 4.42-2.72 5.39-5.31 5.68.42.36.79 1.07.79 2.16v3.21c0 .31.21.67.8.56A11.5 11.5 0 0 0 12 .7Z"/>
            </svg>
          </a>
          <a href="https://github.com/sponsors/Slaviors-Group" target="_blank" rel="noopener noreferrer" aria-label="Sponsor DAWG">
            <svg viewBox="0 0 24 24" aria-hidden="true">
              <path fill="currentColor" d="m12 21.35-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09A6 6 0 0 1 16.5 3C19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35Z"/>
            </svg>
          </a>
        </div>
      </div>
      <p class="dh-foot-copyright">© 2026 Slaviors Group. All rights reserved.</p>
    </div>

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
  transition: color var(--duration-fast);
}
.dh-foot-social svg {
  display: block;
  width: 20px;
  height: 20px;
}
.dh-foot-social a:hover {
  color: var(--color-text-primary);
}

.dh-foot-copyright {
  margin: 0;
  padding: 20px 0 40px;
  border-top: 1px solid var(--color-border);
  color: var(--color-text-tertiary);
  font-size: 13px;
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
  background-clip: text;
  color: transparent;
}
</style>
