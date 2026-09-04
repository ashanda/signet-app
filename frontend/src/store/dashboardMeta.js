import { defineStore } from 'pinia'

// Small shared-state store for cross-page widget values that originate from
// whichever dashboard/list payload last carried them. Every
// Admin/Agent/Company/User dashboard response includes `new_activations`
// (see backend/internal/handlers/dashboard_handler.go, tree.NewActivations)
// — the sidebar's "Waiting Activations" badge (ui_spec.md sidebar section)
// reads it from here instead of polling a dedicated endpoint, since the
// original only ever computed it as part of a full dashboard render too.
export const useDashboardMetaStore = defineStore('dashboardMeta', {
  state: () => ({
    newActivations: 0,
  }),
  actions: {
    setNewActivations(n) {
      this.newActivations = n ?? 0
    },
  },
})
