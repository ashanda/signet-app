import api from '@/api/client'
import { useApiAction } from '@/composables/useApiAction'

// Ports the recurring ".active-package" AJAX pattern (ui_spec.md "Active
// Package/activation AJAX pattern"): POST {package_id} to one of two
// endpoints depending on the page, SweetAlert2 success/error, then the
// caller reloads/refetches its list on confirm.
//
// endpoint: '/active-package' (Admin/Agent/User dashboards, Company
// pending-activation) or '/company/new-active-package' (Company
// new-activations page — see dashboard_handler.go's companyNewActivePackageHandler,
// the Go port of the original's `/company/new_active-package`).
export function useActivePackage(endpoint = '/active-package') {
  const { run } = useApiAction()

  async function activate(packageId) {
    const { ok } = await run(() => api.post(endpoint, { package_id: packageId }), {
      successMessage: 'Package activated successfully.',
    })
    return ok
  }

  return { activate }
}
