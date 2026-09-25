import { defineConfig, type HeadConfig } from "vitepress";
import tailwindcss from "@tailwindcss/vite";

const siteUrl = "https://dawg.slaviors.id";
const socialImage = `${siteUrl}/og-image.png`;
const defaultDescription =
  "Capture, sanitize, package, share, and replay browser bug reproductions as portable OCI-based .dawg artifacts.";

const softwareApplicationSchema = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "WebSite",
      "@id": `${siteUrl}/#website`,
      name: "DAWG",
      alternateName: "Digs Any Web-app Glitch",
      url: `${siteUrl}/`,
      description: defaultDescription,
      inLanguage: "en-US",
    },
    {
      "@type": "SoftwareApplication",
      "@id": `${siteUrl}/#software`,
      name: "DAWG",
      alternateName: "Digs Any Web-app Glitch",
      applicationCategory: "DeveloperApplication",
      operatingSystem: "Windows 10, Windows 11, Linux",
      softwareVersion: "0.3.5-middlechild",
      url: `${siteUrl}/`,
      downloadUrl: [
        "https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_x64-setup.exe",
        "https://github.com/Slaviors-Group/dawg/releases/download/middlechild-5/DAWG_0.3.5_amd64.AppImage",
      ],
      image: socialImage,
      license: "https://github.com/Slaviors-Group/dawg/blob/main/LICENSE",
      codeRepository: "https://github.com/Slaviors-Group/dawg",
      description: defaultDescription,
    },
  ],
};

// https://vitepress.dev/reference/site-config
export default defineConfig({
  lang: "en-US",
  title: "DAWG",
  titleTemplate: ":title | DAWG",
  description: defaultDescription,
  cleanUrls: true,
  sitemap: {
    hostname: siteUrl,
  },

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
    ["meta", { name: "theme-color", content: "#6B21D9" }],
    ["meta", { name: "color-scheme", content: "light" }],
  ],

  transformHead({ pageData }) {
    if (pageData.relativePath === "404.md") {
      return [["meta", { name: "robots", content: "noindex, nofollow" }]];
    }

    const route =
      pageData.relativePath === "index.md"
        ? "/"
        : `/${pageData.relativePath.replace(/\.md$/, "").replace(/\/index$/, "")}`;
    const canonicalUrl = new URL(route, `${siteUrl}/`).href;
    const pageTitle = pageData.title || "DAWG";
    const pageDescription = pageData.description || defaultDescription;
    const isHome = pageData.relativePath === "index.md";
    const socialTitle = isHome ? pageTitle : `${pageTitle} | DAWG`;

    const head: HeadConfig[] = [
      ["link", { rel: "canonical", href: canonicalUrl }],
      [
        "meta",
        {
          name: "robots",
          content: "index, follow, max-image-preview:large, max-snippet:-1, max-video-preview:-1",
        },
      ],
      ["meta", { property: "og:type", content: isHome ? "website" : "article" }],
      ["meta", { property: "og:site_name", content: "DAWG" }],
      ["meta", { property: "og:locale", content: "en_US" }],
      ["meta", { property: "og:title", content: socialTitle }],
      ["meta", { property: "og:description", content: pageDescription }],
      ["meta", { property: "og:url", content: canonicalUrl }],
      ["meta", { property: "og:image", content: socialImage }],
      ["meta", { property: "og:image:type", content: "image/png" }],
      ["meta", { property: "og:image:width", content: "1200" }],
      ["meta", { property: "og:image:height", content: "630" }],
      [
        "meta",
        {
          property: "og:image:alt",
          content: "DAWG developer tool for capturing and replaying web application bugs",
        },
      ],
      ["meta", { name: "twitter:card", content: "summary_large_image" }],
      ["meta", { name: "twitter:title", content: socialTitle }],
      ["meta", { name: "twitter:description", content: pageDescription }],
      ["meta", { name: "twitter:image", content: socialImage }],
      [
        "meta",
        {
          name: "twitter:image:alt",
          content: "DAWG developer tool for capturing and replaying web application bugs",
        },
      ],
    ];

    if (isHome) {
      head.push([
        "script",
        { type: "application/ld+json" },
        JSON.stringify(softwareApplicationSchema),
      ]);
    }

    return head;
  },

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
          { text: "Artifact Anatomy", link: "/docs/artifact-anatomy" },
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
          items: [{ text: "CLI Reference", link: "/docs/cli-reference" }],
        },
        {
          text: "Architecture",
          items: [
            { text: "Overview", link: "/docs/architecture" },
            { text: "Artifact Anatomy", link: "/docs/artifact-anatomy" },
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
            {
              text: "License (GPL-3.0)",
              link: "https://github.com/Slaviors-Group/dawg/blob/main/LICENSE",
            },
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
