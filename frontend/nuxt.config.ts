import tailwindcss from "@tailwindcss/vite";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2024-11-01",
  devtools: {
    enabled: true,

    timeline: {
      enabled: true,
    },
  },
  runtimeConfig: {
    apiUrl: "",
    assetsUrl: "",
    public: {
      wsUrl: "",
    },
  },
  devServer: {
    port: 3000,
  },
  modules: [
    "@vueuse/nuxt",
    "@nuxt/eslint",
    "@nuxtjs/color-mode",
    "shadcn-nuxt",
    "@nuxt/image",
    "@pinia/nuxt",
    "pinia-plugin-persistedstate/nuxt",
    "@vee-validate/nuxt",
    "vue-sonner/nuxt",
    "motion-v/nuxt",
    "@nuxt/fonts",
  ],
  css: ["./app/assets/css/tailwind.css"],
  vite: {
    plugins: [tailwindcss()],
  },
  colorMode: {
    preference: "system",
    fallback: "dark",
  },
  shadcn: {
    prefix: "",
    componentDir: "./app/components/ui",
  },
  image: {
    domains: ["localhost"],
    alias: {
      archive: `${process.env.NUXT_API_URL}/a`,
      assets: `${process.env.NUXT_ASSETS_URL}`,
    },
  },
  fonts: {
    defaults: {
      weights: ["100 900"],
    },
  },
});

