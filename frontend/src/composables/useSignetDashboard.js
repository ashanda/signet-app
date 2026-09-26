import { watch, nextTick, onBeforeUnmount } from 'vue'
import { initSignetDashboard } from './signetDashboard'

// Runs the signet-dashboard effects layer (stat icons, miner scene, clock,
// HUD, count-ups) once `ready` turns truthy — i.e. once the dashboard's data
// has rendered, which is when the original's script ran against the
// server-rendered page — and tears it down when the dashboard unmounts.
// The route must carry meta.bodyClass 'sg-dashboard' (see router/index.js).
export function useSignetDashboard(ready) {
  let cleanup = null

  const stop = watch(
    ready,
    async (value) => {
      if (!value || cleanup) return
      await nextTick()
      cleanup = initSignetDashboard()
      stop()
    },
    { immediate: true },
  )

  onBeforeUnmount(() => {
    stop()
    if (cleanup) cleanup()
    cleanup = null
  })
}
