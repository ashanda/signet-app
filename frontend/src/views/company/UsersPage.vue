<script setup>
// Ports company/users.blade.php ("Find Users" — route('company.users'),
// shared by company and admin). See user_handler.go's allUsersHandler /
// userSearchHandler / userSimpleStatusUpdateHandler /
// updateGlobalDirectorShareHandler / updateUserCode / leader-status toggle.
//
// As in the original:
//  - the list search is a GET form, and pagination keeps the search term;
//  - both "Assign Leader" and "Assign Executive" pick from the same
//    active-leaders list;
//  - picking the person who already holds the other role is refused with
//    a "Cannot Assign" warning;
//  - every modal save shows "Updated!" and reloads the page; a logical
//    failure ({success:false, message}) shows it in an error dialog;
//  - the Leader Status switch is company-only (admins see plain text).
// The modals' Close buttons only close (in the original they are untyped
// buttons inside the form, so they also submitted it).
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Swal from 'sweetalert2'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import { useAuthStore } from '@/store/auth'
import { useToast } from '@/composables/useToast'
import { useBsModal } from '@/composables/useBsModal'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const { toastMixin } = useToast()

const statusModal = useBsModal()
const rocModal = useBsModal()
const gdsModal = useBsModal()
const leaderModal = useBsModal()
const executiveModal = useBsModal()

const loadError = ref('')
const allUsers = ref(null)
const leaders = ref([])
const search = ref('')

async function fetchUsers() {
  loadError.value = ''
  const q = route.query
  search.value = q.search || ''
  try {
    const params = {}
    for (const key of ['page', 'search']) if (q[key]) params[key] = q[key]
    const { data } = await api.get('/users', { params })
    allUsers.value = data.users
    leaders.value = data.leaders || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load users.'
  }
}

watch(() => route.query, fetchUsers, { immediate: true })

function submitSearch() {
  router.push({ path: route.path, query: { search: search.value } })
}

// ->appends(['search' => $request->search])
function goToPage(page) {
  router.push({ path: route.path, query: { search: route.query.search, page } })
}

const ucfirst = (s) => (s ? s.charAt(0).toUpperCase() + s.slice(1) : '')

// --- Leader Status switch ---
async function toggleLeaderStatus(user, event) {
  const checkbox = event.target
  const status = checkbox.checked ? 'active' : 'inactive'
  try {
    const { data } = await api.post(`/users/update-leader-status/${user.id}`, { status })
    if (data.success) {
      user.leader_status = status
      toastMixin.fire({ icon: 'success', title: `Leader status changed to ${ucfirst(status)}` })
    } else {
      toastMixin.fire({ icon: 'error', title: data.message ?? 'Failed to update status' })
      checkbox.checked = !checkbox.checked // revert toggle
    }
  } catch {
    toastMixin.fire({ icon: 'error', title: 'Server error. Try again.' })
    checkbox.checked = !checkbox.checked // revert toggle
  }
}

// --- Signet ID lookup ---
const searchSignetId = ref('')
const found = ref(null)

async function searchSignet() {
  const id = searchSignetId.value
  if (id.trim() === '') return
  try {
    const { data } = await api.get(`/users/search/${id}`)
    if (data.success) found.value = data
    else Swal.fire('Not Found', 'No record found for this Signet ID Or Account has Pending Status', 'error')
  } catch {
    Swal.fire('Not Found', 'No record found for this Signet ID Or Account has Pending Status', 'error')
  }
}

function rocBadge(status) {
  if (status === 'active') return 'bg-success'
  if (status === 'inactive') return 'bg-warning'
  return 'bg-danger'
}

// --- modal forms ---
const statusForm = ref({ id: '', status: 'active' })
const rocForm = ref({ id: '', status: 'active' })
const gdsForm = ref({ id: '', value: '', status: '1' })
const leaderForm = ref({ id: '', name: '', leader_code: '', executive_code: '' })
const executiveForm = ref({ id: '', name: '', executive_code: '', leader_code: '' })

function openStatus() {
  statusForm.value = { id: found.value.user.id, status: found.value.user.status }
  statusModal.show()
}
function openRoc() {
  rocForm.value = { id: found.value.user.id, status: found.value.user.roc_status }
  rocModal.show()
}
function openGds() {
  gdsForm.value = { id: found.value.user.id, value: found.value.user.global_director_share, status: found.value.user.global_director_share_status ? '1' : '0' }
  gdsModal.show()
}
function openLeader(user) {
  leaderForm.value = { id: user.id, name: user.name, leader_code: user.leader_code ?? '', executive_code: user.executive_code ?? '' }
  leaderModal.show()
}
function openExecutive(user) {
  executiveForm.value = { id: user.id, name: user.name, executive_code: user.executive_code ?? '', leader_code: user.leader_code ?? '' }
  executiveModal.show()
}

function onLeaderChange() {
  const selected = leaderForm.value.leader_code
  if (selected && String(selected) === String(leaderForm.value.executive_code)) {
    Swal.fire({ icon: 'warning', title: 'Cannot Assign', text: 'This person is already assigned as the Executive.' })
    leaderForm.value.leader_code = ''
  }
}
function onExecutiveChange() {
  const selected = executiveForm.value.executive_code
  if (selected && String(selected) === String(executiveForm.value.leader_code)) {
    Swal.fire({ icon: 'warning', title: 'Cannot Assign', text: 'This person is already assigned as the Leader.' })
    executiveForm.value.executive_code = ''
  }
}

async function save(request, successText) {
  try {
    const { data } = await request()
    if (data.success) {
      Swal.fire('Updated!', successText, 'success').then(() => window.location.reload())
    } else {
      Swal.fire('Error', data.message ?? 'Update failed', 'error')
    }
  } catch (err) {
    Swal.fire('Error', err?.response?.data?.message ?? 'Update failed', 'error')
  }
}

const submitStatus = () => save(() => api.post(`/users/update/${statusForm.value.id}`, { status: statusForm.value.status }), 'User status updated successfully')
const submitRoc = () => save(() => api.post(`/users/update-roc/${rocForm.value.id}`, { status: rocForm.value.status }), 'User ROC status updated successfully')
const submitGds = () =>
  save(
    () => api.post(`/users/update-global-director-share/${gdsForm.value.id}`, { value: Number(gdsForm.value.value) || 0, status: gdsForm.value.status }),
    'User Global Director Share status updated successfully',
  )
const submitLeader = () => save(() => api.post(`/users/update-leader-code/${leaderForm.value.id}`, { leader_code: leaderForm.value.leader_code }), 'Leader assigned successfully')
const submitExecutive = () =>
  save(() => api.post(`/users/update-executive-code/${executiveForm.value.id}`, { executive_code: executiveForm.value.executive_code }), 'Executive assigned successfully')

const PENCIL =
  '<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-pencil-square" viewBox="0 0 16 16"><path d="M15.502 1.94a.5.5 0 0 1 0 .706L14.459 3.69l-2-2L13.502.646a.5.5 0 0 1 .707 0l1.293 1.293zm-1.75 2.456-2-2L4.939 9.21a.5.5 0 0 0-.121.196l-.805 2.414a.25.25 0 0 0 .316.316l2.414-.805a.5.5 0 0 0 .196-.12l6.813-6.814z"/><path fill-rule="evenodd" d="M1 13.5A1.5 1.5 0 0 0 2.5 15h11a1.5 1.5 0 0 0 1.5-1.5v-6a.5.5 0 0 0-1 0v6a.5.5 0 0 1-.5.5h-11a.5.5 0 0 1-.5-.5v-11a.5.5 0 0 1 .5-.5H9a.5.5 0 0 0 0-1H2.5A1.5 1.5 0 0 0 1 2.5z"/></svg>'
</script>

<template>
  <DashboardLayout>
    <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Find Users</h1>
        </div>
        <div>
          <RouterLink :to="{ name: 'leader.code.logs' }" class="btn btn-outline-primary">
            <i class="bi bi-clock-history me-1"></i>Leader Code Logs
          </RouterLink>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <!-- Search Box -->
        <div class="input-group mb-3">
          <input id="search_signet_id" v-model="searchSignetId" type="text" class="form-control" placeholder="Enter Signet ID" @keypress.enter.prevent="searchSignet" />
          <button id="searchBtn" class="btn btn-primary" type="button" @click="searchSignet">Search</button>
        </div>

        <!-- Search Result Table -->
        <div id="resultCard" :class="{ 'd-none': !found }">
          <div v-if="found" class="card shadow-lg border-0 rounded-4 overflow-hidden mb-4 hover-lift" style="transition: transform 0.2s">
            <div class="card-header bg-gradient text-white p-4 sg-card-hero">
              <div class="d-flex align-items-center justify-content-between">
                <div>
                  <h5 class="card-title mb-1">{{ found.user.name }}</h5>
                  <small class="opacity-75">ID: {{ found.user.id }}</small>
                </div>
                <div class="text-end">
                  <span class="badge px-3 py-2 rounded-pill" :class="found.user.status === 'active' ? 'bg-success' : 'bg-danger'">
                    <i class="bi bi-circle-fill me-1" style="font-size: 0.5rem"></i>
                    {{ String(found.user.status).toUpperCase() }}
                  </span>
                  <span class="badge px-3 py-2 rounded-pill" :class="rocBadge(found.user.roc_status)">
                    <i class="bi bi-circle-fill me-1" style="font-size: 0.5rem"></i>
                    ROC: {{ String(found.user.roc_status).toUpperCase() }}
                  </span>
                  <span class="badge px-3 py-2 rounded-pill" :class="found.user.global_director_share_status ? 'bg-success' : 'bg-danger'">
                    <i class="bi bi-circle-fill me-1" style="font-size: 0.5rem"></i>
                    Global Director Share: {{ found.user.global_director_share_status ? 'Active' : 'Inactive' }}
                  </span>
                </div>
              </div>
            </div>

            <div class="card-body p-4">
              <!-- Email Section -->
              <div class="mb-4 pb-3 border-bottom">
                <div class="d-flex align-items-center">
                  <i class="bi bi-envelope-fill text-primary me-2"></i>
                  <span class="text-muted small">Email:</span>
                </div>
                <p class="mb-0 mt-1 fw-medium">{{ found.user.email }}</p>
              </div>

              <!-- Packages Section -->
              <div class="mb-4 pb-3 border-bottom">
                <h6 class="text-uppercase text-muted small mb-3 fw-bold">
                  <i class="bi bi-box-seam-fill text-warning me-2"></i>Packages
                </h6>
                <div class="row g-3">
                  <div class="col-3">
                    <div class="p-3 bg-light rounded-3">
                      <div class="small text-muted mb-1">First Package</div>
                      <div class="fw-bold text-dark">{{ found.packages.first || 'N/A' }}</div>
                    </div>
                  </div>
                  <div class="col-3">
                    <div class="p-3 bg-light rounded-3">
                      <div class="small text-muted mb-1">Current Package</div>
                      <div class="fw-bold text-dark">{{ found.packages.last || 'N/A' }}</div>
                    </div>
                  </div>
                  <div class="col-3">
                    <div class="p-3 bg-light rounded-3">
                      <div class="small text-muted mb-1">Sale Count</div>
                      <div class="fw-bold text-dark">{{ found.sales.total_sales || 'N/A' }}</div>
                    </div>
                  </div>
                  <div class="col-3">
                    <div class="p-3 bg-light rounded-3">
                      <div class="small text-muted mb-1">Direct Sale Count</div>
                      <div class="fw-bold text-dark">{{ found.sales.direct_sales || 'N/A' }}</div>
                    </div>
                  </div>
                  <div class="col-3">
                    <div class="p-3 bg-light rounded-3">
                      <div class="small text-muted mb-1">Wallet Balance</div>
                      <div class="fw-bold text-dark">{{ found.wallet.total_wallet || 'N/A' }} USDT</div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Mining Data Section -->
            </div>

            <div class="card-footer bg-white border-0 p-4 pt-0">
              <button class="btn btn-warning w-100 py-2 fw-bold rounded-3 editBtn shadow-sm" @click="openStatus">
                <i class="bi bi-pencil-square me-2"></i>Change Account Status
              </button>
            </div>
            <div class="card-footer bg-white border-0 p-4 pt-0">
              <button class="btn btn-primary w-100 py-2 fw-bold rounded-3 editRocBtn shadow-sm" @click="openRoc">
                <i class="bi bi-pencil-square me-2"></i>Change ROC Status
              </button>
            </div>
            <div class="card-footer bg-white border-0 p-4 pt-0">
              <button class="btn btn-success w-100 py-2 fw-bold rounded-3 editGlobalDirectorShareBtn shadow-sm text-white" @click="openGds">
                <i class="bi bi-pencil-square me-2"></i>Change Global Director Share
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <!-- Search Form -->
        <form method="GET" class="mb-3" @submit.prevent="submitSearch">
          <div class="input-group">
            <input v-model="search" type="text" name="search" class="form-control" placeholder="Search by name, ID, WhatsApp, or status, Country Code (eg: US, IN, etc.)" />
            <button class="btn btn-primary" type="submit">Search</button>
            <RouterLink v-if="route.query.search" :to="{ name: 'company.users' }" class="btn btn-outline-secondary">Clear</RouterLink>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">User name</th>
                <th class="border-bottom" scope="col">SIGNET ID</th>
                <th class="border-bottom" scope="col">WhatsApp</th>
                <th class="border-bottom" scope="col">Status</th>
                <th class="border-bottom" scope="col">Leader Status</th>
                <th class="border-bottom" scope="col">Leader</th>
                <th class="border-bottom" scope="col">Executive</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="user in allUsers?.data" :key="user.id">
                <td>{{ user.name }}</td>
                <td>{{ 'SIG-00' + user.id }}</td>
                <td>{{ user.whatsapp_number }}</td>
                <td>{{ ucfirst(user.status) }}</td>
                <td>
                  <div v-if="authStore.user?.role === 'company'" class="form-check form-switch">
                    <input type="checkbox" class="form-check-input toggle-leader-status" :checked="user.leader_status == 'active'" @change="toggleLeaderStatus(user, $event)" />
                    <label class="form-check-label">
                      {{ ucfirst(user.leader_status || 'inactive') }}
                    </label>
                  </div>
                  <template v-else>
                    {{ ucfirst(user.leader_status || 'inactive') }}
                  </template>
                </td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <span>{{ user.leader_id ? 'SIG-00' + user.leader_id : '—' }}</span>
                    <button type="button" class="btn btn-sm btn-outline-secondary editLeaderCodeBtn" @click="openLeader(user)" v-html="PENCIL"></button>
                  </div>
                </td>
                <td>
                  <div class="d-flex align-items-center gap-2">
                    <span>{{ user.executive_id ? 'SIG-00' + user.executive_id : '—' }}</span>
                    <button type="button" class="btn btn-sm btn-outline-secondary editExecutiveCodeBtn" @click="openExecutive(user)" v-html="PENCIL"></button>
                  </div>
                </td>
              </tr>
              <tr v-if="allUsers && !allUsers.data?.length">
                <td colspan="6" class="text-center text-muted">No users found</td>
              </tr>
            </tbody>
          </table>

          <div class="d-flex justify-content-center">
            <Paginator :pagination="allUsers" @change="goToPage" />
          </div>
        </div>
      </div>
    </div>

    <!-- Update Modal -->
    <div id="updateModal" :ref="statusModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="updateForm" @submit.prevent="submitStatus">
            <div class="modal-header">
              <h5 class="modal-title">Update User Status</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <input id="update_id" v-model="statusForm.id" type="hidden" name="id" />
              <div class="mb-3">
                <label>Status</label>
                <select id="update_status" v-model="statusForm.status" name="status" class="form-control">
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Update ROC Modal -->
    <div id="updateModalRoc" :ref="rocModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="updateFormRoc" @submit.prevent="submitRoc">
            <div class="modal-header">
              <h5 class="modal-title">Update User ROC Status</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <input id="update_id_roc" v-model="rocForm.id" type="hidden" name="id" />
              <div class="mb-3">
                <label>ROC Status</label>
                <select id="update_status_roc" v-model="rocForm.status" name="status" class="form-control">
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Update Global Director Share Modal -->
    <div id="updateModalGlobalDirectorShare" :ref="gdsModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="updateFormGlobalDirectorShare" @submit.prevent="submitGds">
            <div class="modal-header">
              <h5 class="modal-title">Update User Global Director Share Status</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <input id="update_id_global_director_share" v-model="gdsForm.id" type="hidden" name="id" />
              <div class="mb-3">
                <label>Global Director Share Value</label>
                <input id="update_value_global_director_share" v-model="gdsForm.value" type="number" name="value" class="form-control" />
              </div>
              <div class="mb-3">
                <label>Global Director Share Status</label>
                <select id="update_status_global_director_share" v-model="gdsForm.status" name="status" class="form-control">
                  <option value="1">Active</option>
                  <option value="0">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Update Leader Code Modal -->
    <div id="updateModalLeaderCode" :ref="leaderModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="updateFormLeaderCode" @submit.prevent="submitLeader">
            <div class="modal-header">
              <h5 class="modal-title">
                Assign Leader<span id="leaderCodeUserName" class="text-muted">{{ leaderForm.name ? ' - ' + leaderForm.name : '' }}</span>
              </h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <input id="update_id_leader_code" v-model="leaderForm.id" type="hidden" name="id" />
              <div class="mb-3">
                <label>Leader</label>
                <select id="update_leader_code" v-model="leaderForm.leader_code" name="leader_code" class="form-control" @change="onLeaderChange">
                  <option value="">No Leader</option>
                  <option v-for="leader in leaders" :key="leader.id" :value="String(leader.id)">
                    {{ 'SIG-00' + leader.id }} - {{ leader.name }}
                  </option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <div id="updateModalExecutiveCode" :ref="executiveModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="updateFormExecutiveCode" @submit.prevent="submitExecutive">
            <div class="modal-header">
              <h5 class="modal-title">
                Assign Executive Code<span id="executiveCodeUserName" class="text-muted">{{ executiveForm.name ? ' - ' + executiveForm.name : '' }}</span>
              </h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <input id="update_id_executive_code" v-model="executiveForm.id" type="hidden" name="id" />
              <div class="mb-3">
                <label>Executive</label>
                <select id="update_executive_code" v-model="executiveForm.executive_code" name="executive_code" class="form-control" @change="onExecutiveChange">
                  <option value="">No Executive</option>
                  <option v-for="leader in leaders" :key="leader.id" :value="String(leader.id)">
                    {{ 'SIG-00' + leader.id }} - {{ leader.name }}
                  </option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
