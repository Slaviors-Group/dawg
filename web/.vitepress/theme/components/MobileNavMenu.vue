<template>
  <nav class="mobile-nav" aria-label="Mobile navigation">
    <a class="mobile-nav-link" href="/">Home</a>

    <section class="mobile-nav-group">
      <h2>Guide</h2>
      <a href="/docs/getting-started">Getting Started</a>
      <a href="/docs/installation">Installation</a>
      <a href="/docs/cli-reference">CLI Reference</a>
    </section>

    <section class="mobile-nav-group">
      <h2>Architecture</h2>
      <a href="/docs/architecture">Overview</a>
      <a href="/docs/sanitizer-policy">Sanitizer Policy</a>
    </section>

    <a class="mobile-nav-link" href="/docs/changelog">Changelog</a>

    <section class="mobile-nav-group">
      <h2>Project</h2>
      <a href="https://github.com/Slaviors-Group/dawg/releases">Releases</a>
      <a href="/docs/contributing">Contributing</a>
      <a href="https://github.com/Slaviors-Group/dawg">GitHub</a>
    </section>
  </nav>
</template>

<style scoped>
.mobile-nav {
  display: grid;
  gap: 8px;
}

.mobile-nav a {
  color: var(--vp-c-text-1);
  font-size: 16px;
  font-weight: 500;
  line-height: 1.4;
  text-decoration: none;
}

.mobile-nav-link {
  padding: 12px 0;
}

.mobile-nav-group {
  display: grid;
  gap: 10px;
  margin-top: 18px;
}

.mobile-nav-group h2 {
  margin: 0 0 2px;
  color: var(--vp-c-text-3);
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
}

.mobile-nav-group a {
  padding: 4px 0;
}
</style>
