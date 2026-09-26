import Swal from 'sweetalert2'
import api from '@/api/client'

// Ports the recurring `$('.active-package').on('click', ...)` jQuery handler:
// POST {package_id}, then a SweetAlert2 success dialog whose OK reloads the
// page, or an error dialog.
//
// endpoint: '/active-package' (dashboards, Company pending-activation) or
// '/company/new-active-package' (Company new-activations — the Go port of
// the original's `/company/new_active-package`).
//
// The views word their alerts differently; by default this matches the
// activation pages (server message when present, else the fixed text):
//   successFromResponse: text = response.message || 'Package updated successfully!'
//   errorFromResponse:   text = responseJSON.message || 'Error updating package.'
export function useActivePackage(endpoint = '/active-package', { successFromResponse = true, errorFromResponse = true } = {}) {
  async function activate(packageId) {
    try {
      const { data: response } = await api.post(endpoint, { package_id: packageId })
      const result = await Swal.fire({
        icon: 'success',
        title: 'Success!',
        text: (successFromResponse && response?.message) || 'Package updated successfully!',
      })
      if (result.isConfirmed) window.location.reload()
      return true
    } catch (err) {
      const serverMessage = err?.response?.data?.message
      await Swal.fire({
        icon: 'error',
        title: 'Error',
        text: (errorFromResponse && serverMessage) || 'Error updating package.',
      })
      return false
    }
  }

  return { activate }
}
