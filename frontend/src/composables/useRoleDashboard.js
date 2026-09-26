import { ref, onMounted } from 'vue'
import Swal from 'sweetalert2'
import api from '@/api/client'
import { useAuthStore } from '@/store/auth'
import { useDashboardMetaStore } from '@/store/dashboardMeta'
import { useActivePackage } from '@/composables/useActivePackage'

// Shared behaviour behind admin/agent/user dashboard.blade.php (each page
// keeps its own template, since the three Blade views differ in layout).
// Response shape: dashboard_handler.go's admin/agent/user handlers —
// {ref_link, my_tokens, total_value, wallet_balance, my_package,
// fee_percentage, pool_amount, total_poolshare_value, my_share_value,
// my_globle_director_share, activations:{..paginated..}, rank, roc,
// all_users, new_activations}.
//
// activeMessages: the three views' `.active-package` handlers word their
// alerts differently —
//   admin: fixed success + fixed error text
//   user:  fixed success, error from responseJSON.message
//   agent: success from response.message, error from responseJSON.message
export function useRoleDashboard(endpoint, activeMessages = {}) {
  const { successFromResponse = false, errorFromResponse = false } = activeMessages
  const { activate } = useActivePackage('/active-package', { successFromResponse, errorFromResponse })
  const authStore = useAuthStore()
  const dashboardMeta = useDashboardMetaStore()

  const resp = ref(null)
  const loadError = ref('')
  const onVacation = ref(!!authStore.user?.on_vacation)
  const urlDisplay = ref('')

  async function fetchDashboard(page = 1) {
    loadError.value = ''
    try {
      const { data } = await api.get(endpoint, { params: { page } })
      resp.value = data
      dashboardMeta.setNewActivations(data.new_activations)
    } catch (err) {
      loadError.value = err?.response?.data?.message || 'Could not load dashboard data.'
    }
  }

  onMounted(() => fetchDashboard())

  // $('#vacationSwitch').on('change', ...) — the label only changes on
  // success; on failure the checkbox is left as the user flipped it.
  async function toggleVacation() {
    try {
      const { data: res } = await api.post('/toggle-vacation')
      onVacation.value = !!res.on_vacation
      if (authStore.user) authStore.user.on_vacation = onVacation.value
      Swal.fire({
        icon: 'success',
        title: 'Updated!',
        text: 'Vacation mode has been turned ' + (res.on_vacation ? 'ON' : 'OFF') + '.',
        timer: 2000,
        showConfirmButton: false,
      })
    } catch {
      Swal.fire({
        icon: 'error',
        title: 'Error',
        text: 'Failed to update vacation status.',
        confirmButtonColor: '#d33',
      })
    }
  }

  // custome.js's #copyButton handler.
  function copyLink() {
    const url = resp.value?.ref_link ?? ''
    navigator.clipboard
      .writeText(url)
      .then(() => {
        urlDisplay.value = 'Copied URL: ' + url
      })
      .catch(() => {
        urlDisplay.value = 'Failed to copy URL'
      })
  }

  function needTokens(row) {
    const price = Number(row.package_price || 0)
    const fee = Number(resp.value?.fee_percentage || 0)
    return price - price * (fee / 100)
  }

  return { authStore, resp, loadError, onVacation, urlDisplay, fetchDashboard, toggleVacation, copyLink, needTokens, activate }
}
