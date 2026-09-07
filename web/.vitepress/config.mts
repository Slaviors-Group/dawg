import { defineConfig } from "vitepress";
import tailwindcss from "@tailwindcss/vite";

// https://vitepress.dev/reference/site-config
export default defineConfig({
  title: "DAWG - Digs Any Web-app Glitch",
  description:
    "A Portable, Versioned Bug-Reproduction Artifact Platform for Deterministic Full-Stack Web Application Debugging",
  vite: {
    plugins: [tailwindcss()],
  },
});
