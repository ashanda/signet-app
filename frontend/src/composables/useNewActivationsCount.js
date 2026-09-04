import { storeToRefs } from 'pinia'
import { useDashboardMetaStore } from '@/store/dashboardMeta'

// Thin read-only wrapper around dashboardMeta for the sidebar badge — see
// store/dashboardMeta.js for why this isn't its own polling endpoint.
export function useNewActivationsCount() {
  const store = useDashboardMetaStore()
  const { newActivations: count } = storeToRefs(store)
  return { count }
}
