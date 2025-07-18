import { defineStore } from 'pinia'

export const usePageStore = defineStore('page', {
  state: () => {
    return {
      currentPage: 1,
      pageSize: 30
    }
  },
  persist: {
    storage: piniaPluginPersistedstate.cookies(),
  },
})
