import { defineConfig } from "vitepress";
import tailwindcss from "@tailwindcss/vite";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "DAWG - Digs Any Web-app Glitch",
  description:
    "A Portable, Versioned Bug-Reproduction Artifact Platform for Deterministic Full-Stack Web Application Debugging",

  // Keep the site light-only until the custom dark theme is ready.
  // This also disables VitePress's system-theme detection and appearance switch.
  appearance: false,

  head: [
    ["link", { rel: "preconnect", href: "https://fonts.googleapis.com" }],
    [
      "link",
      {
        rel: "preconnect",
        href: "https://fonts.gstatic.com",
        crossorigin: "",
      },
    ],
    [
      "link",
      {
        href: "https://fonts.googleapis.com/css2?family=Outfit:wght@300;400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap",
        rel: "stylesheet",
      },
    ],
    ["link", { rel: "icon", type: "image/svg+xml", href: "/paw-dawg.svg" }],
    ["script", { src: "https://unpkg.com/@phosphor-icons/web" }],
  ],

  vite: {
    plugins: [tailwindcss()],
  },

  themeConfig: {
    logo: "/paw-dawg.svg",

    nav: [
      { text: "Home", link: "/" },
      {
        text: "Guide",
        items: [
          { text: "Getting Started", link: "/docs/getting-started" },
          { text: "Installation", link: "/docs/installation" },
          { text: "CLI Reference", link: "/docs/cli-reference" },
        ],
      },
      {
        text: "Architecture",
        items: [
          { text: "Overview", link: "/docs/architecture" },
          { text: "Sanitizer Policy", link: "/docs/sanitizer-policy" },
        ],
      },
      { text: "Changelog", link: "/docs/changelog" },
      {
        text: "v0.3.5-middlechild",
        items: [
          {
            text: "Current release",
            link: "https://github.com/Slaviors-Group/dawg/releases/tag/middlechild-5",
          },
          {
            text: "Releases",
            link: "https://github.com/Slaviors-Group/dawg/releases",
          },
          {
            text: "Contributing",
            link: "/docs/contributing",
          },
        ],
      },
    ],

    sidebar: {
      "/docs/": [
        {
          text: "Introduction",
          items: [
            { text: "Getting Started", link: "/docs/getting-started" },
            { text: "Installation", link: "/docs/installation" },
          ],
        },
        {
          text: "Guide",
          items: [
            { text: "CLI Reference", link: "/docs/cli-reference" },
            { text: "Architecture", link: "/docs/architecture" },
            { text: "Sanitizer Policy", link: "/docs/sanitizer-policy" },
          ],
        },
        {
          text: "Community",
          items: [
            { text: "Contributing", link: "/docs/contributing" },
            { text: "Changelog", link: "/docs/changelog" },
          ],
        },
        {
          text: "Legal",
          items: [
            { text: "Privacy Policy", link: "/docs/privacy-policy" },
            { text: "Terms of Service", link: "/docs/terms-of-service" },
          ],
        },
      ],
    },

    socialLinks: [
      { icon: "github", link: "https://github.com/Slaviors-Group/dawg" },
      {
        icon: {
          svg: '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true"><path d="M12 21.35 10.55 20.03C5.4 15.36 2 12.28 2 8.5 2 5.42 4.42 3 7.5 3c1.74 0 3.41.81 4.5 2.09C13.09 3.81 14.76 3 16.5 3 19.58 3 22 5.42 22 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35Z"/></svg>',
        },
        link: "https://github.com/sponsors/Slaviors-Group",
        ariaLabel: "Sponsor DAWG",
      },
    ],

    search: {
      provider: "local",
    },

    editLink: {
      pattern:
        "https://github.com/Slaviors-Group/dawg/edit/update-web/web/:path",
      text: "Edit this page on GitHub",
    },
  },
});
