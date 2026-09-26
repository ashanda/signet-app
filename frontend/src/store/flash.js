import { defineStore } from 'pinia'

// Laravel's `redirect()->route(...)->with('success', '...')`: a message set
// just before navigating, shown once by the page navigated to.
export const useFlashStore = defineStore('flash', {
  state: () => ({ success: '', error: '' }),
  actions: {
    flash(type, message) {
      this[type] = message
    },
    // Returns and clears the pending messages.
    take() {
      const out = { success: this.success, error: this.error }
      this.success = ''
      this.error = ''
      return out
    },
  },
})
