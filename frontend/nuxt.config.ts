import tailwindcss from "@tailwindcss/vite";

const baseApi = 'http://localhost:8081/api'
const baseWs = 'ws://localhost:8081/ws'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: {
    enabled: true,

    timeline: {
      enabled: true,
    },
  },
  runtimeConfig: {
    baseApi: baseApi,
    public: {
      ws: baseWs,
    }
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
    '@nuxt/image',
    '@pinia/nuxt',
    'pinia-plugin-persistedstate/nuxt',
    '@vee-validate/nuxt',
    'vue-sonner/nuxt',
    '@sidebase/nuxt-auth',
  ],
  css: ['./app/assets/css/tailwind.css'],
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
    componentDir: './app/components/ui'
  },
  image: {
    domains: ['localhost'],
    alias: {
      archive: `${baseApi}/a`,
    }
  },
  auth: {
    isEnabled: true,
    provider: {
      type: 'local',
      endpoints: {
        signIn: { path: '/login', method: 'post'},
        signUp: { path: '/register', method: 'post'},
        signOut: { path: '/logout', method: 'post'},
        getSession: { path: '/session', method: 'get'}
      },
      token: {
        signInResponseTokenPointer: '/token',
        type: 'Bearer',
        cookieName: 'auth.token',
        headerName: 'Authorization',
        maxAgeInSeconds: 259200,
        sameSiteAttribute: 'lax',
        cookieDomain: '',
        secureCookieAttribute: false,
        httpOnlyCookieAttribute: false,
      },
      session: {
        dataType: {
          id: 'number',
          username: 'string',
          created_at: 'string,'
        }
      }
    },
    globalAppMiddleware: true,
  }
})
