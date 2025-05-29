import { defineStore } from 'pinia'

export const usePageStore = defineStore('pageStore', {
  state: () => {
    return {
      currentPage: 1,
    }
  },
  persist: {
    storage: piniaPluginPersistedstate.cookies(),
  },
})
