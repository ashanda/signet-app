<script setup>
// Ports company/users.blade.php (route('company.users'), shared with admin).
// See user_handler.go's allUsersHandler / userSearchHandler /
// userSimpleStatusUpdateHandler / updateGlobalDirectorShareHandler /
// updateUserCode (leader/executive code assignment).
//
// Key deviations from the original, verified against the real blade source
// (resources/views/company/users.blade.php) and the Go handlers — trusted
// over ui_spec.md's prose:
//  - The "Assign Executive" modal's select is populated from the SAME
//    `$leaders`/`leaders` list as "Assign Leader" (the original never
//    fetches a separate `$executives` list for this page — genuine quirk,
//    reproduced verbatim, not a bug in this port).
//  - The five modal-submit endpoints (`/users/update/{id}`,
//    `/users/update-roc/{id}`, `/users/update-global-director-share/{id}`,
//    `/users/update-leader-code/{id}`, `/users/update-executive-code/{id}`)
//    all decode a strict JSON body server-side (`decodeJSON`), unlike the
//    original's `FormData` fetch — so these submit as JSON here, not
//    multipart.
//  - Those same five endpoints (plus the leader-status toggle) reply
//    `{success:false, message:...}` with HTTP 200 on a *logical* failure
//    (e.g. "A user cannot be their own leader"), which does not fit
//    useApiAction's `status:'error'` convention — so this page handles
//    them with a small local submitAction() helper instead of
//    useApiAction, to avoid mis-firing a success alert on a false success.
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAuthStore } from '@/store/auth'
import { useToast, useAlert } from '@/composables/useToast'
import Swal from 'sweetalert2'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { toastSuccess, toastError } = useToast()
const { alertSuccess, alertError } = useAlert()

const users = ref(null)
const leaders = ref([])
const loading = ref(true)
const loadError = ref('')
const searchTerm = ref(route.query.search || '')

const searchId = ref('')
const lookupResult = ref(null)

async function fetchUsers(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (searchTerm.value) params.search = searchTerm.value
    const { data } = await api.get('/users', { params })
    users.value = data.users
    leaders.value = data.leaders || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load users.'
  } finally {
    loading.value = false
  }
}

function applySearch() {
  router.replace({ query: searchTerm.value ? { search: searchTerm.value } : {} })
  fetchUsers(1)
}

function clearSearch() {
  searchTerm.value = ''
  router.replace({ query: {} })
  fetchUsers(1)
}

async function searchSignet() {
  const id = searchId.value.trim()
  if (!id) return
  try {
    const { data } = await api.get(`/users/search/${id}`)
    if (data.success) {
      lookupResult.value = data
    } else {
      lookupResult.value = null
      await Swal.fire('Not Found', 'No record found for this Signet ID Or Account has Pending Status', 'error')
    }
  } catch {
    lookupResult.value = null
    await Swal.fire('Not Found', 'No record found for this Signet ID Or Account has Pending Status', 'error')
  }
}

async function refreshAfterModal() {
  await fetchUsers(users.value?.current_page || 1)
  if (lookupResult.value) {
    const id = lookupResult.value.user.id
    const { data } = await api.get(`/users/search/${id}`)
    if (data.success) lookupResult.value = data
  }
}

// --- shared submit helper: these endpoints answer 200 OK with
// {success:false,message} on a logical failure, which useApiAction's
// status:'error' convention does not catch — handled manually here.
async function submitAction(fn) {
  try {
    const { data } = await fn()
    if (data.success) return { ok: true, data }
    await alertError('Error', data.message || 'Update failed')
    return { ok: false, data }
  } catch (err) {
    await alertError('Error', err?.response?.data?.message || 'Update failed')
    return { ok: false }
  }
}

// --- Update User Status modal ---
const showStatusModal = ref(false)
const statusForm = ref({ id: '', status: 'active' })
function openStatusModal() {
  if (!lookupResult.value) return
  statusForm.value = { id: lookupResult.value.user.id, status: lookupResult.value.user.status }
  showStatusModal.value = true
}
async function submitStatus() {
  const { ok } = await submitAction(() =>
    api.post(`/users/update/${statusForm.value.id}`, { status: statusForm.value.status })
  )
  if (ok) {
    showStatusModal.value = false
    await alertSuccess('Updated!', 'User status updated successfully')
    await refreshAfterModal()
  }
}

// --- Update User ROC Status modal ---
const showRocModal = ref(false)
const rocForm = ref({ id: '', status: 'active' })
function openRocModal() {
  if (!lookupResult.value) return
  rocForm.value = { id: lookupResult.value.user.id, status: lookupResult.value.user.roc_status || 'active' }
  showRocModal.value = true
}
async function submitRoc() {
  const { ok } = await submitAction(() =>
    api.post(`/users/update-roc/${rocForm.value.id}`, { status: rocForm.value.status })
  )
  if (ok) {
    showRocModal.value = false
    await alertSuccess('Updated!', 'User ROC status updated successfully')
    await refreshAfterModal()
  }
}

// --- Update Global Director Share modal ---
const showGdsModal = ref(false)
const gdsForm = ref({ id: '', value: 0, status: '1' })
function openGdsModal() {
  if (!lookupResult.value) return
  gdsForm.value = {
    id: lookupResult.value.user.id,
    value: lookupResult.value.user.global_director_share || 0,
    status: lookupResult.value.user.global_director_share_status ? '1' : '0',
  }
  showGdsModal.value = true
}
async function submitGds() {
  const { ok } = await submitAction(() =>
    api.post(`/users/update-global-director-share/${gdsForm.value.id}`, {
      value: Number(gdsForm.value.value) || 0,
      status: gdsForm.value.status,
    })
  )
  if (ok) {
    showGdsModal.value = false
    await alertSuccess('Updated!', 'User Global Director Share status updated successfully')
    await refreshAfterModal()
  }
}

// --- Assign Leader modal ---
const showLeaderModal = ref(false)
const leaderForm = ref({ id: '', name: '', leader_code: '' })
const leaderOtherCode = ref('') // row's current executive_code, for the collision guard
function openLeaderModal(row) {
  leaderForm.value = { id: row.id, name: row.name, leader_code: row.leader_code || '' }
  leaderOtherCode.value = row.executive_code || ''
  showLeaderModal.value = true
}
function onLeaderSelectChange() {
  if (leaderForm.value.leader_code && leaderForm.value.leader_code === leaderOtherCode.value) {
    Swal.fire({ icon: 'warning', title: 'Cannot Assign', text: 'This person is already assigned as the Executive.' })
    leaderForm.value.leader_code = ''
  }
}
async function submitLeader() {
  const { ok } = await submitAction(() =>
    api.post(`/users/update-leader-code/${leaderForm.value.id}`, { leader_code: leaderForm.value.leader_code })
  )
  if (ok) {
    showLeaderModal.value = false
    await alertSuccess('Updated!', 'Leader assigned successfully')
    await refreshAfterModal()
  }
}

// --- Assign Executive modal ---
const showExecutiveModal = ref(false)
const executiveForm = ref({ id: '', name: '', executive_code: '' })
const executiveOtherCode = ref('') // row's current leader_code, for the collision guard
function openExecutiveModal(row) {
  executiveForm.value = { id: row.id, name: row.name, executive_code: row.executive_code || '' }
  executiveOtherCode.value = row.leader_code || ''
  showExecutiveModal.value = true
}
function onExecutiveSelectChange() {
  if (executiveForm.value.executive_code && executiveForm.value.executive_code === executiveOtherCode.value) {
    Swal.fire({ icon: 'warning', title: 'Cannot Assign', text: 'This person is already assigned as the Leader.' })
    executiveForm.value.executive_code = ''
  }
}
async function submitExecutive() {
  const { ok } = await submitAction(() =>
    api.post(`/users/update-executive-code/${executiveForm.value.id}`, {
      executive_code: executiveForm.value.executive_code,
    })
  )
  if (ok) {
    showExecutiveModal.value = false
    await alertSuccess('Updated!', 'Executive assigned successfully')
    await refreshAfterModal()
  }
}

// --- Leader Status row toggle ---
async function onToggleLeaderStatus(row, event) {
  const checked = event.target.checked
  const newStatus = checked ? 'active' : 'inactive'
  try {
    const { data } = await api.post(`/users/update-leader-status/${row.id}`, { status: newStatus })
    if (data.success) {
      row.leader_status = newStatus
      toastSuccess(`Leader status changed to ${newStatus.charAt(0).toUpperCase() + newStatus.slice(1)}`)
    } else {
      toastError(data.message || 'Failed to update status')
      event.target.checked = !checked
    }
  } catch {
    toastError('Server error. Try again.')
    event.target.checked = !checked
  }
}

onMounted(() => fetchUsers())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Find Users</h1>
        </div>
        <div>
          <router-link to="/leader-code-logs" class="btn btn-outline-primary">
            <i class="fas fa-history me-1"></i>Leader Code Logs
          </router-link>
        </div>
      </div>
    </div>

    <!-- Quick lookup card -->
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="input-group mb-3">
          <input
            v-model="searchId"
            type="text"
            class="form-control"
            placeholder="Enter Signet ID"
            @keypress.enter.prevent="searchSignet"
          />
          <button class="btn btn-primary" type="button" @click="searchSignet">Search</button>
        </div>

        <div v-if="lookupResult" class="card shadow-lg border-0 rounded-4 overflow-hidden mb-0">
          <div class="card-header text-white p-4 users-gradient-header">
            <div class="d-flex align-items-center justify-content-between flex-wrap gap-2">
              <div>
                <h5 class="card-title mb-1 text-dark">{{ lookupResult.user.name }}</h5>
                <small class="opacity-75">ID: {{ lookupResult.user.id }}</small>
              </div>
              <div class="d-flex flex-wrap gap-2">
                <span class="badge px-3 py-2 rounded-pill" :class="lookupResult.user.status === 'active' ? 'bg-success' : 'bg-danger'">
                  {{ lookupResult.user.status?.toUpperCase() }}
                </span>
                <span
                  class="badge px-3 py-2 rounded-pill"
                  :class="{
                    'bg-success': lookupResult.user.roc_status === 'active',
                    'bg-warning': lookupResult.user.roc_status === 'inactive',
                    'bg-danger': !['active', 'inactive'].includes(lookupResult.user.roc_status),
                  }"
                >
                  ROC: {{ (lookupResult.user.roc_status || '').toUpperCase() }}
                </span>
                <span class="badge px-3 py-2 rounded-pill" :class="lookupResult.user.global_director_share_status ? 'bg-success' : 'bg-danger'">
                  Global Director Share: {{ lookupResult.user.global_director_share_status ? 'Active' : 'Inactive' }}
                </span>
              </div>
            </div>
          </div>

          <div class="card-body p-4">
            <div class="mb-4 pb-3 border-bottom">
              <div class="d-flex align-items-center">
                <i class="fas fa-envelope text-primary me-2"></i>
                <span class="text-muted small">Email:</span>
              </div>
              <p class="mb-0 mt-1 fw-medium">{{ lookupResult.user.email }}</p>
            </div>

            <div class="mb-1">
              <h6 class="text-uppercase text-muted small mb-3 fw-bold">
                <i class="fas fa-box text-warning me-2"></i>Packages
              </h6>
              <div class="row g-3">
                <div class="col-6 col-md-3">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">First Package</div>
                    <div class="fw-bold text-dark">{{ lookupResult.packages.first || 'N/A' }}</div>
                  </div>
                </div>
                <div class="col-6 col-md-3">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">Current Package</div>
                    <div class="fw-bold text-dark">{{ lookupResult.packages.last || 'N/A' }}</div>
                  </div>
                </div>
                <div class="col-6 col-md-2">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">Sale Count</div>
                    <div class="fw-bold text-dark">{{ lookupResult.sales.total_sales ?? 'N/A' }}</div>
                  </div>
                </div>
                <div class="col-6 col-md-2">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">Direct Sale Count</div>
                    <div class="fw-bold text-dark">{{ lookupResult.sales.direct_sales ?? 'N/A' }}</div>
                  </div>
                </div>
                <div class="col-6 col-md-2">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">Wallet Balance</div>
                    <div class="fw-bold text-dark">{{ lookupResult.wallet.total_wallet ?? 'N/A' }} USDT</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="card-footer bg-white border-0 p-4 pt-0">
            <button class="btn btn-warning w-100 py-2 fw-bold rounded-3 shadow-sm" @click="openStatusModal">
              <i class="fas fa-pencil-alt me-2"></i>Change Account Status
            </button>
          </div>
          <div class="card-footer bg-white border-0 p-4 pt-0">
            <button class="btn btn-primary w-100 py-2 fw-bold rounded-3 shadow-sm" @click="openRocModal">
              <i class="fas fa-pencil-alt me-2"></i>Change ROC Status
            </button>
          </div>
          <div class="card-footer bg-white border-0 p-4 pt-0">
            <button class="btn btn-success w-100 py-2 fw-bold rounded-3 shadow-sm text-white" @click="openGdsModal">
              <i class="fas fa-pencil-alt me-2"></i>Change Global Director Share
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Search/filter + table card -->
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="mb-3" @submit.prevent="applySearch">
          <div class="input-group">
            <input
              v-model="searchTerm"
              type="text"
              class="form-control"
              placeholder="Search by name, ID, WhatsApp, or status, Country Code (eg: US, IN, etc.)"
            />
            <button class="btn btn-primary" type="submit">Search</button>
            <button v-if="searchTerm" type="button" class="btn btn-outline-secondary" @click="clearSearch">Clear</button>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User name</th>
                <th class="border-bottom">SIGNET ID</th>
                <th class="border-bottom">WhatsApp</th>
                <th class="border-bottom">Status</th>
                <th class="border-bottom">Leader Status</th>
                <th class="border-bottom">Leader</th>
                <th class="border-bottom">Executive</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !users?.data?.length">
                <td colspan="7" class="text-center text-muted py-4">No users found</td>
              </tr>
              <tr v-for="row in users?.data" :key="row.id">
                <td>{{ row.name }}</td>
                <td>{{ row.signet_id }}</td>
                <td>{{ row.whatsapp_number || '—' }}</td>
                <td>{{ row.status ? row.status.charAt(0).toUpperCase() + row.status.slice(1) : '—' }}</td>
                <td>
                  <div v-if="authStore.isCompany" class="form-check form-switch">
                    <input
                      type="checkbox"
                      class="form-check-input"
                      :checked="row.leader_status === 'active'"
                      @change="onToggleLeaderStatus(row, $event)"
                    />
                    <label class="form-check-label">
                      {{ (row.leader_status || 'inactive').charAt(0).toUpperCase() + (row.leader_status || 'inactive').slice(1) }}
                    </label>
                  </div>
                  <span v-else>{{ (row.leader_status || 'inactive').charAt(0).toUpperCase() + (row.leader_status || 'inactive').slice(1) }}</span>
                </td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <span>{{ row.leader_code ? `SIG-00${row.leader_code}` : '—' }}</span>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="openLeaderModal(row)">
                      <i class="fas fa-pencil-alt"></i>
                    </button>
                  </div>
                </td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <span>{{ row.executive_code ? `SIG-00${row.executive_code}` : '—' }}</span>
                    <button type="button" class="btn btn-sm btn-outline-secondary" @click="openExecutiveModal(row)">
                      <i class="fas fa-pencil-alt"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="users" @change="(p) => fetchUsers(p)" />
      </div>
    </div>

    <!-- Update User Status modal -->
    <div v-if="showStatusModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitStatus">
            <div class="modal-header">
              <h5 class="modal-title">Update User Status</h5>
              <button type="button" class="btn-close" @click="showStatusModal = false"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">Status</label>
                <select v-model="statusForm.status" class="form-select">
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showStatusModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Update User ROC Status modal -->
    <div v-if="showRocModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitRoc">
            <div class="modal-header">
              <h5 class="modal-title">Update User ROC Status</h5>
              <button type="button" class="btn-close" @click="showRocModal = false"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">ROC Status</label>
                <select v-model="rocForm.status" class="form-select">
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showRocModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Update Global Director Share modal -->
    <div v-if="showGdsModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitGds">
            <div class="modal-header">
              <h5 class="modal-title">Update User Global Director Share Status</h5>
              <button type="button" class="btn-close" @click="showGdsModal = false"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">Global Director Share Value</label>
                <input v-model="gdsForm.value" type="number" class="form-control" />
              </div>
              <div class="mb-3">
                <label class="form-label">Global Director Share Status</label>
                <select v-model="gdsForm.status" class="form-select">
                  <option value="1">Active</option>
                  <option value="0">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showGdsModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Assign Leader modal -->
    <div v-if="showLeaderModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitLeader">
            <div class="modal-header">
              <h5 class="modal-title">Assign Leader <span class="text-muted">- {{ leaderForm.name }}</span></h5>
              <button type="button" class="btn-close" @click="showLeaderModal = false"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">Leader</label>
                <select v-model="leaderForm.leader_code" class="form-select" @change="onLeaderSelectChange">
                  <option value="">No Leader</option>
                  <option v-for="l in leaders" :key="l.id" :value="String(l.id)">{{ l.signet_id }} - {{ l.name }}</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showLeaderModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Assign Executive modal -->
    <div v-if="showExecutiveModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitExecutive">
            <div class="modal-header">
              <h5 class="modal-title">Assign Executive Code <span class="text-muted">- {{ executiveForm.name }}</span></h5>
              <button type="button" class="btn-close" @click="showExecutiveModal = false"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">Executive</label>
                <select v-model="executiveForm.executive_code" class="form-select" @change="onExecutiveSelectChange">
                  <option value="">No Executive</option>
                  <option v-for="l in leaders" :key="l.id" :value="String(l.id)">{{ l.signet_id }} - {{ l.name }}</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showExecutiveModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>

<style scoped>
.users-gradient-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
</style>
