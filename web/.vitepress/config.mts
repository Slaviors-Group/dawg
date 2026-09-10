import { defineConfig } from "vitepress";
import tailwindcss from "@tailwindcss/vite";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "DAWG - Digs Any Web-app Glitch",
  description:
    "A Portable, Versioned Bug-Reproduction Artifact Platform for Deterministic Full-Stack Web Application Debugging",

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
        text: "v0.1.2-alpha",
        items: [
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
      ],
    },

    socialLinks: [
      { icon: "github", link: "https://github.com/Slaviors-Group/dawg" },
    ],

    search: {
      provider: "local",
    },

    editLink: {
      pattern:
        "https://github.com/Slaviors-Group/dawg/edit/staging/web/:path",
      text: "Edit this page on GitHub",
    },

    footer: {
      message: "Released under the Apache-2.0 License.",
      copyright: "Copyright © 2026 Slaviors-Group",
    },
  },
});
