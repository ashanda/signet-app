<script setup>
// Ports salaries/index.blade.php ("Salaries"). See salary_handler.go's
// salariesIndexHandler/salariesSearchUsersHandler/salariesStoreHandler.
//
// Deviations from ui_spec.md's documented markup, verified against the
// actual Go handler (trusted over prose):
//  - The original's "Add Salary" modal uses Select2's AJAX-backed
//    autocomplete (route('salaries.searchUsers')). Per FRONTEND_CONVENTIONS.md
//    this is implemented here as a plain debounced text input + dropdown
//    list of results instead of pulling in jQuery/Select2.
//  - The commented-out "Actions"/delete column and its paired `.deleteBtn`
//    script are dead code in the original (no button ever rendered) —
//    preserved as an omission here too, even though the Go backend does
//    implement DELETE /salaries/{id}.
//  - `salaries.*`'s sql.Null* fields (remarks, created_at, user_name,
//    user_whatsapp) are embedded straight into JSON as {String,Valid}/
//    {Time,Valid} objects (no server-side flatten) — nullStr()/nullTime()
//    unwrap that, matching RocIncomePage.vue's established pattern.
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'

const route = useRoute()
const router = useRouter()
const { run } = useApiAction()

function nullStr(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}

function dateOnly(v) {
  const s = nullStr(v) || v
  if (!s) return ''
  return String(s).slice(0, 10)
}

const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')

const dateFrom = ref(route.query.date_from || '')
const dateTo = ref(route.query.date_to || '')
const salaries = ref(null)
const totalAmount = ref(0)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (dateFrom.value) params.date_from = dateFrom.value
    if (dateTo.value) params.date_to = dateTo.value
    const { data } = await api.get('/salaries', { params })
    salaries.value = data.salaries
    totalAmount.value = data.total_amount
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load salaries.'
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  router.replace({ query: { ...(dateFrom.value ? { date_from: dateFrom.value } : {}), ...(dateTo.value ? { date_to: dateTo.value } : {}) } })
  fetchData(1)
}

function clearFilter() {
  dateFrom.value = ''
  dateTo.value = ''
  router.replace({ query: {} })
  fetchData(1)
}

// --- Add Salary modal ---
const showAddModal = ref(false)
const addForm = ref({ user_id: '', amount: '', salary_date: '', remarks: '' })
const userSearchTerm = ref('')
const userResults = ref([])
const userSearching = ref(false)
const selectedUserLabel = ref('')
const showUserDropdown = ref(false)
let searchTimer = null

watch(userSearchTerm, (term) => {
  showUserDropdown.value = true
  clearTimeout(searchTimer)
  searchTimer = setTimeout(async () => {
    userSearching.value = true
    try {
      const { data } = await api.get('/salaries/search-users', { params: { term } })
      userResults.value = data || []
    } catch {
      userResults.value = []
    } finally {
      userSearching.value = false
    }
  }, 300)
})

function pickUser(u) {
  addForm.value.user_id = u.id
  selectedUserLabel.value = u.text
  userSearchTerm.value = u.text
  showUserDropdown.value = false
}

function openAddModal() {
  addForm.value = { user_id: '', amount: '', salary_date: '', remarks: '' }
  userSearchTerm.value = ''
  selectedUserLabel.value = ''
  userResults.value = []
  errors.value = {}
  showAddModal.value = true
}

function closeAddModal() {
  showAddModal.value = false
}

const errors = ref({})
const saving = ref(false)

async function submitAdd() {
  errors.value = {}
  saving.value = true
  const { ok, data, error } = await run(
    () =>
      api.post('/salaries', {
        user_id: Number(addForm.value.user_id) || 0,
        amount: Number(addForm.value.amount) || 0,
        salary_date: addForm.value.salary_date,
        remarks: addForm.value.remarks,
      }),
    { successMessage: 'Salary added successfully', showErrorAlert: false }
  )
  saving.value = false
  if (ok) {
    showAddModal.value = false
    flashMessage.value = data?.message || 'Salary added successfully'
    fetchData(1)
  } else {
    const respErrors = error?.response?.data?.errors
    if (respErrors) {
      errors.value = respErrors
    } else {
      errors.value = { _general: [error?.response?.data?.message || 'Could not add salary.'] }
    }
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="success" :message="flashMessage" @close="flashMessage = ''" />
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4 d-flex align-items-center justify-content-between">
      <h1 class="h4 mb-0">Salaries</h1>
      <button type="button" class="btn btn-primary" @click="openAddModal">
        <i class="fas fa-plus me-1"></i> Add Salary
      </button>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="row g-2 align-items-end mb-4" @submit.prevent="applyFilter">
          <div class="col-auto">
            <label class="form-label small mb-1">From Date</label>
            <input v-model="dateFrom" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <label class="form-label small mb-1">To Date</label>
            <input v-model="dateTo" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <button type="submit" class="btn btn-primary">Filter</button>
          </div>
          <div v-if="dateFrom || dateTo" class="col-auto">
            <a href="#" class="btn btn-outline-dark" @click.prevent="clearFilter">Clear</a>
          </div>
          <div class="col-auto ms-auto">
            <span class="fw-semibold">Total in range: {{ Number(totalAmount || 0).toFixed(2) }}</span>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User</th>
                <th class="border-bottom">SIGNET ID</th>
                <th class="border-bottom">Amount</th>
                <th class="border-bottom">Date</th>
                <th class="border-bottom">Remarks</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !salaries?.data?.length">
                <td colspan="5" class="text-center text-muted py-4">No salary records found</td>
              </tr>
              <tr v-for="row in salaries?.data" :key="row.id">
                <td>{{ nullStr(row.user_name) || '—' }}</td>
                <td>SIG-00{{ row.user_id }}</td>
                <td>{{ Number(row.amount || 0).toFixed(2) }}</td>
                <td>{{ dateOnly(row.salary_date) }}</td>
                <td>{{ nullStr(row.remarks) || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="salaries" @change="(p) => fetchData(p)" />
      </div>
    </div>

    <!-- Add Salary modal -->
    <div v-if="showAddModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitAdd">
            <div class="modal-header">
              <h5 class="modal-title">Add Salary</h5>
              <button type="button" class="btn-close" @click="closeAddModal"></button>
            </div>
            <div class="modal-body">
              <div v-if="errors._general" class="alert alert-danger py-2">{{ errors._general[0] }}</div>
              <div class="mb-3 position-relative">
                <label class="form-label">User</label>
                <input
                  v-model="userSearchTerm"
                  type="text"
                  class="form-control"
                  :class="{ 'is-invalid': errors.user_id }"
                  placeholder="Search by name, Signet ID, or WhatsApp"
                  autocomplete="off"
                  @focus="showUserDropdown = true"
                />
                <div v-if="errors.user_id" class="invalid-feedback">{{ errors.user_id[0] }}</div>
                <ul
                  v-if="showUserDropdown && userSearchTerm && userResults.length"
                  class="list-group position-absolute w-100 shadow"
                  style="z-index: 1060; max-height: 220px; overflow-y: auto;"
                >
                  <li
                    v-for="u in userResults"
                    :key="u.id"
                    class="list-group-item list-group-item-action"
                    style="cursor: pointer;"
                    @click="pickUser(u)"
                  >
                    {{ u.text }}
                  </li>
                </ul>
                <div v-if="showUserDropdown && userSearchTerm && !userSearching && !userResults.length" class="form-text">
                  No matching users.
                </div>
              </div>
              <div class="mb-3">
                <label class="form-label">Amount</label>
                <input v-model="addForm.amount" type="number" step="0.01" min="0" class="form-control" :class="{ 'is-invalid': errors.amount }" required />
                <div v-if="errors.amount" class="invalid-feedback">{{ errors.amount[0] }}</div>
              </div>
              <div class="mb-3">
                <label class="form-label">Date</label>
                <input v-model="addForm.salary_date" type="date" class="form-control" :class="{ 'is-invalid': errors.salary_date }" required />
                <div v-if="errors.salary_date" class="invalid-feedback">{{ errors.salary_date[0] }}</div>
              </div>
              <div class="mb-3">
                <label class="form-label">Remarks</label>
                <input v-model="addForm.remarks" type="text" class="form-control" :class="{ 'is-invalid': errors.remarks }" />
                <div v-if="errors.remarks" class="invalid-feedback">{{ errors.remarks[0] }}</div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeAddModal">Close</button>
              <button type="submit" class="btn btn-primary" :disabled="saving">Save</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
