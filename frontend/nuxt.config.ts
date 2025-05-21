import tailwindcss from "@tailwindcss/vite";

const baseApi = 'http://localhost:8081/api'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },
  runtimeConfig: {
    baseApi: baseApi,
  },
  devServer: {
    port: 8080,
  },
  modules: [
    '@nuxt/ui',
    '@vueuse/nuxt',
    '@nuxt/eslint',
    '@nuxt/fonts',
    '@nuxtjs/color-mode',
    'shadcn-nuxt',
    '@nuxt/image'
  ],
  css: ['~/assets/css/tailwind.css'],
  vite: {
    plugins: [
      tailwindcss(),
    ],
  },
  ui: {
    theme: {
      colors: ['primary', 'secondary', 'tertiary', 'info', 'success', 'warning', 'error']
    }
  },
  colorMode: {
    preference: 'dark',
  },
  shadcn: {
    prefix: '',
    componentDir: './components/ui'
  },
  image: {
    domains: ['localhost'],
    alias: {
      archive: `${baseApi}/a`,
    } 
  },
})