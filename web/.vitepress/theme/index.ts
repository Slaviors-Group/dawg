import DefaultTheme from "vitepress/theme";
import Layout from "./Layout.vue";
import type { Theme } from "vitepress";
import "./style.css";

export default {
  extends: DefaultTheme,
  Layout,
  enhanceApp({ app, router, siteData }) {
    // ...
  },
} satisfies Theme;
