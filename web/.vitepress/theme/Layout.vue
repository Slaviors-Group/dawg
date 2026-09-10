<script setup lang="ts">
import DefaultTheme from 'vitepress/theme'
import { useData } from 'vitepress'
import { computed } from 'vue'

import HomeHero from './components/HomeHero.vue'
import HomeTooling from './components/HomeTooling.vue'
import HomePipeline from './components/HomePipeline.vue'
import HomeCodeCli from './components/HomeCodeCli.vue'
import HomeResources from './components/HomeResources.vue'
import HomeFinalCta from './components/HomeFinalCta.vue'
import HomeFooter from './components/HomeFooter.vue'

const { Layout: DefaultLayout } = DefaultTheme
const { frontmatter } = useData()

const isHome = computed(() => frontmatter.value.layout === 'home')
</script>

<template>
  <DefaultLayout :class="{ 'dh-home-root': isHome }">
    <template #home-hero-before>
      <div class="dawg-home">
        <HomeHero />
        <HomeTooling />
        <HomePipeline />
        <HomeCodeCli />
        <HomeResources />
        <HomeFinalCta />
        <HomeFooter />
      </div>
    </template>
  </DefaultLayout>
</template>

<style>
/* Home takes over chrome from the default theme */
.dh-home-root .VPFooter { display: none !important; }
.dh-home-root .VPContent { padding-top: 0 !important; }
.dh-home-root .VPContent.has-sidebar { padding-top: 0 !important; }
/* VitePress's default .VPHome adds a 6-8rem bottom margin, which left a slab
   of bare page background showing below our own footer. Kill it. */
.dh-home-root .VPHome { margin-bottom: 0 !important; }

/* Global variables that need to be accessed by all components if they use them in :global */
/* Dynamic Navbar Title */
.VPNavBarTitle .title span {
  font-size: 0;
}
.VPNavBarTitle .title span::after {
  content: "DAWG";
  font-size: 16px;
}
.dh-home-root .VPNavBarTitle .title span::after {
  content: "Digs Any Web-app Glitch";
}
</style>

<style scoped>
/* ═══ Modern Editorial Style ═══ */
.dawg-home {
  background: #ffffff;
  color: var(--color-text-primary);
  font-family: var(--font-sans);
  position: relative;
  overflow-x: hidden;
}
</style>
