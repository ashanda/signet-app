import { useAlert } from '@/composables/useToast'

// The Go backend preserves the original's inconsistent success/error
// signaling verbatim (see BACKEND_CONVENTIONS.md "Response shape" +
// api_spec.md's per-endpoint notes): some endpoints answer 200 OK with
// `{status:'error', message:...}` for a *logical* failure (e.g.
// buyPackagesHandler), others answer a non-2xx HTTP status with just
// `{message:...}` (e.g. activePackageHandler's "Not enough tokens."). This
// helper normalizes both into one success/failure result so page code
// doesn't have to special-case each endpoint's convention.
//
// Usage: const { run } = useApiAction(); const ok = await run(() =>
// api.post('/some-endpoint', body), { successMessage: 'Saved!' })
export function useApiAction() {
  const { alertSuccess, alertError } = useAlert()

  async function run(fn, opts = {}) {
    const { successMessage, showSuccessAlert = true, showErrorAlert = true } = opts
    try {
      const { data } = await fn()
      if (data && data.status === 'error') {
        if (showErrorAlert) await alertError(data.message || 'Something went wrong')
        return { ok: false, data }
      }
      if (showSuccessAlert) {
        await alertSuccess((data && data.message) || successMessage || 'Success!')
      }
      return { ok: true, data }
    } catch (err) {
      const message = err?.response?.data?.message || 'Something went wrong'
      if (showErrorAlert) await alertError(message)
      return { ok: false, error: err }
    }
  }

  return { run }
}
