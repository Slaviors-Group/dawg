const fs = require("node:fs");
const path = require("node:path");

const SITE_URL = "https://dawg.slaviors.id";
const DIST_DIR = path.resolve(__dirname, "..", ".vitepress", "dist");

function fail(message) {
  throw new Error(`SEO check failed: ${message}`);
}

function walk(directory) {
  return fs.readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const target = path.join(directory, entry.name);
    return entry.isDirectory() ? walk(target) : [target];
  });
}

function attribute(html, selector) {
  const match = html.match(selector);
  return match?.[1] || "";
}

function routeOutputPath(route) {
  if (route === "/") return path.join(DIST_DIR, "index.html");
  if (route.endsWith("/")) return path.join(DIST_DIR, route.slice(1), "index.html");
  return path.join(DIST_DIR, `${route.slice(1)}.html`);
}

if (!fs.existsSync(DIST_DIR)) {
  fail("production output is missing; run npm run docs:build first");
}

const htmlFiles = walk(DIST_DIR).filter((file) => file.endsWith(".html"));
const sitemapPath = path.join(DIST_DIR, "sitemap.xml");
const robotsPath = path.join(DIST_DIR, "robots.txt");
if (!fs.existsSync(sitemapPath)) fail("sitemap.xml is missing");
if (!fs.existsSync(robotsPath)) fail("robots.txt is missing");

const sitemap = fs.readFileSync(sitemapPath, "utf8");
const robots = fs.readFileSync(robotsPath, "utf8");
if (!robots.includes(`Sitemap: ${SITE_URL}/sitemap.xml`)) {
  fail("robots.txt does not advertise the canonical sitemap");
}

const descriptions = new Map();
for (const file of htmlFiles) {
  const relative = path.relative(DIST_DIR, file).replaceAll(path.sep, "/");
  const html = fs.readFileSync(file, "utf8");
  const isNotFound = relative === "404.html";
  const robotsValue = attribute(html, /<meta name="robots" content="([^"]+)">/);

  if (isNotFound) {
    if (robotsValue !== "noindex, nofollow") fail("404 page is indexable");
    continue;
  }

  const title = attribute(html, /<title>([^<]+)<\/title>/);
  const description = attribute(html, /<meta name="description" content="([^"]+)">/);
  const canonical = attribute(html, /<link rel="canonical" href="([^"]+)">/);
  const ogUrl = attribute(html, /<meta property="og:url" content="([^"]+)">/);

  if (!title) fail(`${relative} has no title`);
  if (!description) fail(`${relative} has no description`);
  if (!canonical.startsWith(`${SITE_URL}/`)) fail(`${relative} has no canonical URL`);
  if (ogUrl !== canonical) fail(`${relative} Open Graph URL differs from its canonical URL`);
  if (!html.includes('<meta property="og:title"')) fail(`${relative} has no Open Graph title`);
  if (!html.includes('<meta property="og:description"')) fail(`${relative} has no Open Graph description`);
  if (!html.includes('<meta property="og:image"')) fail(`${relative} has no Open Graph image`);
  if (!html.includes('<meta name="twitter:card" content="summary_large_image">')) {
    fail(`${relative} has no large Twitter card`);
  }
  if (!robotsValue.startsWith("index, follow")) fail(`${relative} is not indexable`);
  if ((html.match(/<h1(?:\s|>)/g) || []).length !== 1) fail(`${relative} must contain exactly one h1`);
  if (!sitemap.includes(`<loc>${canonical}</loc>`)) fail(`${relative} canonical URL is absent from sitemap.xml`);

  const duplicate = descriptions.get(description);
  if (duplicate) fail(`${relative} duplicates the description used by ${duplicate}`);
  descriptions.set(description, relative);

  for (const match of html.matchAll(/href="(\/[^"]*)"/g)) {
    const href = match[1];
    const route = href.split(/[?#]/, 1)[0];
    if (!route || route.startsWith("/assets/") || path.extname(route)) continue;
    if (!fs.existsSync(routeOutputPath(route))) {
      fail(`${relative} links to missing internal route ${route}`);
    }
  }
}

const home = fs.readFileSync(path.join(DIST_DIR, "index.html"), "utf8");
if (!home.includes('<script type="application/ld+json">')) {
  fail("homepage structured data is missing");
}

console.log(`SEO check passed for ${htmlFiles.length - 1} indexable pages and one 404 page.`);
