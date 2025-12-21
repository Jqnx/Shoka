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
    "@nuxt/fonts",
    "@nuxtjs/color-mode",
    "shadcn-nuxt",
    "@nuxt/image",
    "@pinia/nuxt",
    "pinia-plugin-persistedstate/nuxt",
    "@vee-validate/nuxt",
    "vue-sonner/nuxt",
    "motion-v/nuxt",
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
  //auth: {
  //  isEnabled: true,
  //  provider: {
  //    type: "local",
  //    endpoints: {
  //      signIn: { path: "/login", method: "post" },
  //      signUp: { path: "/register", method: "post" },
  //      signOut: { path: "/logout", method: "post" },
  //      getSession: { path: "/session", method: "get" },
  //    },
  //    token: {
  //      signInResponseTokenPointer: "/token",
  //      type: "Bearer",
  //      cookieName: "auth.token",
  //      headerName: "Authorization",
  //      maxAgeInSeconds: 259200,
  //      sameSiteAttribute: "lax",
  //      cookieDomain: "",
  //      secureCookieAttribute: false,
  //      httpOnlyCookieAttribute: false,
  //    },
  //    session: {
  //      dataType: {
  //        id: "number",
  //        username: "string",
  //        created_at: "string,",
  //      },
  //    },
  //  },
  //  globalAppMiddleware: true,
  //},
});
