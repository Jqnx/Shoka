import { defineStore } from 'pinia'

export const usePageStore = defineStore('page', {
  state: () => {
    return {
      currentPage: 1,
    }
  },
  persist: {
    storage: piniaPluginPersistedstate.cookies(),
  },
})
