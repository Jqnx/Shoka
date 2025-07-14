import { defineStore } from 'pinia'

export const useReaderSettingsStore = defineStore('readerSettings', {
  state: () => {
    return {
      preload: 3,
      fit: '',
    }
  },
  persist: {
    storage: piniaPluginPersistedstate.cookies(),
  },
})