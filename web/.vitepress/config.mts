import { defineConfig } from "vitepress";
import tailwindcss from "@tailwindcss/vite";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "DAWG - Digs Any Web-app Glitch",
  description:
    "A Portable, Versioned Bug-Reproduction Artifact Platform for Deterministic Full-Stack Web Application Debugging",

  locales: {
    root: {
      label: 'English',
      lang: 'en'
    },
    id: {
      label: 'Bahasa Indonesia',
      lang: 'id',
      description: "Platform Artefak Reproduksi Bug Web yang Portabel dan Berversi",
      themeConfig: {
        nav: [
          { text: "Beranda", link: "/id/" },
          {
            text: "Panduan",
            items: [
              { text: "Mulai", link: "/id/docs/getting-started" },
              { text: "Instalasi", link: "/id/docs/installation" },
              { text: "Referensi CLI", link: "/id/docs/cli-reference" },
            ],
          },
          {
            text: "Arsitektur",
            items: [
              { text: "Ikhtisar", link: "/id/docs/architecture" },
              { text: "Kebijakan Sanitasi", link: "/id/docs/sanitizer-policy" },
            ],
          },
          { text: "Catatan Perubahan", link: "/id/docs/changelog" },
          {
            text: "v0.1.2-alpha",
            items: [
              {
                text: "Rilis",
                link: "https://github.com/Slaviors-Group/dawg/releases",
              },
              {
                text: "Berkontribusi",
                link: "/id/docs/contributing",
              },
            ],
          },
        ],
        sidebar: {
          "/id/docs/": [
            {
              text: "Pengenalan",
              items: [
                { text: "Mulai", link: "/id/docs/getting-started" },
                { text: "Instalasi", link: "/id/docs/installation" },
              ],
            },
            {
              text: "Panduan",
              items: [
                { text: "Referensi CLI", link: "/id/docs/cli-reference" },
                { text: "Arsitektur", link: "/id/docs/architecture" },
                { text: "Kebijakan Sanitasi", link: "/id/docs/sanitizer-policy" },
              ],
            },
            {
              text: "Komunitas",
              items: [
                { text: "Berkontribusi", link: "/id/docs/contributing" },
                { text: "Catatan Perubahan", link: "/id/docs/changelog" },
              ],
            },
            {
              text: "Legal",
              items: [
                { text: "Kebijakan Pengguna", link: "/id/docs/user-policy" },
                { text: "Ketentuan Layanan", link: "/id/docs/terms-of-service" },
              ],
            },
          ],
        },
        footer: {
          message: "Dirilis di bawah Lisensi Apache-2.0.",
          copyright: "Hak Cipta © 2026 Slaviors-Group",
        },
        editLink: {
          pattern: "https://github.com/Slaviors-Group/dawg/edit/staging/web/:path",
          text: "Edit halaman ini di GitHub",
        },
      }
    }
  },

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
        {
          text: "Legal",
          items: [
            { text: "User Policy", link: "/docs/user-policy" },
            { text: "Terms of Service", link: "/docs/terms-of-service" },
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
