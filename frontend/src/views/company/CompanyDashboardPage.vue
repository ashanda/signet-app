<script setup>
// Ports company/dashboard.blade.php (route('company.dashboard')). Company's
// dashboard is structurally different from Admin/Agent/User's (ui_spec.md):
// no mining widget, no activations table — instead an "All Tokens" table, a
// "Generate Token" user-picker panel, a date-range filter, KPI cards, and a
// per-package activation summary. See dashboard_handler.go's
// companyDashboardHandler for the exact response shape/field names
// (`from_date`/`to_date` are the query params it reads, NOT
// `start_date`/`end_date`).
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAuthStore } from '@/store/auth'
import { useDashboardMetaStore } from '@/store/dashboardMeta'

const router = useRouter()
const authStore = useAuthStore()
const dashboardMeta = useDashboardMetaStore()

const resp = ref(null)
const loading = ref(true)
const loadError = ref('')

const filters = reactive({ from_date: '', to_date: '' })

const users = ref([])
const selectedUserId = ref('')

async function fetchDashboard(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (filters.from_date) params.from_date = filters.from_date
    if (filters.to_date) params.to_date = filters.to_date
    const { data } = await api.get('/company/dashboard', { params })
    resp.value = data
    dashboardMeta.setNewActivations(data.new_activations)
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load dashboard data.'
  } finally {
    loading.value = false
  }
}

async function fetchUsers() {
  try {
    const { data } = await api.get('/users', { params: { per_page: 1000 } })
    users.value = data?.users?.data || []
  } catch {
    users.value = []
  }
}

function userLabel(u) {
  return u.id === 1 ? 'Company Head' : `SIG-00${u.id}`
}

function applyFilter() {
  fetchDashboard(1)
}

function resetFilter() {
  filters.from_date = ''
  filters.to_date = ''
  fetchDashboard(1)
}

function redirectToTokens() {
  if (!selectedUserId.value) {
    window.alert('Please select a user first.')
    return
  }
  router.push(`/view-tokens/${selectedUserId.value}`)
}

function totalTokens(row) {
  return Number(row.active_count || 0) + Number(row.deactive_count || 0)
}

function totalActivations() {
  return (resp.value?.package_wise_counts || []).reduce((sum, r) => sum + Number(r.active_count || 0), 0)
}

function sharePercent(row) {
  const grand = Number(resp.value?.grand_total || 0)
  if (!grand) return 0
  return Math.round((Number(row.total_value || 0) / grand) * 10000) / 100
}

onMounted(() => {
  fetchDashboard()
  fetchUsers()
})
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div v-if="resp" class="card bg-yellow-100 border-0 shadow mb-4">
      <div class="card-header d-sm-flex flex-row align-items-center flex-0">
        <div class="d-block mb-3 mb-sm-0">
          <h2 class="fs-3 fw-extrabold">
            Welcome, <span class="text-capitalize">Company</span> Head - All Users Count({{ resp.all_users }})
          </h2>
        </div>
      </div>
    </div>

    <div v-if="resp" class="row g-3 mb-4">
      <div class="col-12 col-lg-8">
        <div class="card border-0 shadow h-100">
          <div class="card-header">
            <div class="row align-items-center">
              <div class="col">
                <h2 class="fs-5 fw-bold mb-0">All Tokens</h2>
              </div>
              <div class="col text-end">
                <a href="#" class="btn btn-sm btn-primary">See all</a>
              </div>
            </div>
          </div>
          <div class="table-responsive">
            <table class="table align-items-center table-flush mb-0">
              <thead class="thead-light">
                <tr>
                  <th class="border-bottom" scope="col">User name</th>
                  <th class="border-bottom" scope="col">Total Tokens</th>
                  <th class="border-bottom" scope="col">Active Tokens</th>
                  <th class="border-bottom" scope="col">Used Tokens</th>
                </tr>
              </thead>
              <tbody>
                <tr v-if="!resp.token_counts?.data?.length">
                  <td colspan="4" class="text-center text-muted py-4">No tokens found</td>
                </tr>
                <tr v-for="row in resp.token_counts?.data" :key="row.user_id">
                  <td>{{ row.user_name || 'Unknown User' }}</td>
                  <td>{{ totalTokens(row) }} USDT</td>
                  <td><span class="badge bg-success">{{ row.active_count }} USDT</span></td>
                  <td><span class="badge bg-danger">{{ row.deactive_count }} USDT</span></td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="card-body pt-0">
            <Paginator :pagination="resp.token_counts" @change="(p) => fetchDashboard(p)" />
          </div>
        </div>
      </div>

      <div class="col-12 col-lg-4">
        <div class="card border-0 shadow-lg rounded-4 h-100">
          <div class="card-header bg-gradient-primary text-white rounded-top-4">
            <h5 class="fw-bold mb-0">Generate Token</h5>
          </div>
          <div class="card-body">
            <label class="small fw-semibold mb-1">Select User</label>
            <select id="genTokenUser" v-model="selectedUserId" class="form-select mb-3">
              <option value="">Select User</option>
              <option v-for="u in users" :key="u.id" :value="u.id">{{ userLabel(u) }}</option>
            </select>
            <button type="button" class="btn btn-primary w-100" @click="redirectToTokens">View Tokens</button>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm rounded-4 mb-4 bg-light">
      <div class="card-body py-3">
        <form class="row align-items-end g-2" @submit.prevent="applyFilter">
          <div class="col-md-4">
            <label class="small fw-semibold text-muted mb-1" for="fromDate">From Date</label>
            <input id="fromDate" v-model="filters.from_date" type="date" class="form-control" />
          </div>
          <div class="col-md-4">
            <label class="small fw-semibold text-muted mb-1" for="toDate">To Date</label>
            <input id="toDate" v-model="filters.to_date" type="date" class="form-control" />
          </div>
          <div class="col-md-4 d-flex gap-2">
            <button type="submit" class="btn btn-success w-100">Apply Filter</button>
            <button type="button" class="btn btn-outline-dark w-100" @click="resetFilter">Reset</button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="resp" class="row mb-4">
      <div class="col-md-4">
        <div class="card border-0 shadow-sm rounded-4 text-center p-3">
          <small class="text-muted">Total Activations</small>
          <h3 class="fw-bold text-primary mb-0">{{ totalActivations().toLocaleString() }}</h3>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm rounded-4 text-center p-3">
          <small class="text-muted">Total Value</small>
          <h3 class="fw-bold text-success mb-0">{{ Number(resp.grand_total || 0).toLocaleString() }} USDT</h3>
        </div>
      </div>
      <div class="col-md-4">
        <div class="card border-0 shadow-sm rounded-4 text-center p-3">
          <small class="text-muted">Packages Activated</small>
          <h3 class="fw-bold text-dark mb-0">{{ (resp.package_wise_counts || []).length }}</h3>
        </div>
      </div>
    </div>

    <div v-if="resp" class="card border-0 shadow rounded-4">
      <div class="card-header border-0 pb-0 d-flex justify-content-between align-items-center">
        <h6 class="fw-bold text-dark">Package Activation Summary</h6>
        <span v-if="Number(resp.grand_total) > 0" class="badge bg-success fs-6">
          {{ Number(resp.grand_total).toLocaleString() }} USDT
        </span>
      </div>
      <div class="card-body pt-3">
        <div v-if="!resp.package_wise_counts?.length" class="text-center text-muted py-4">
          No activations found
        </div>
        <div v-for="row in resp.package_wise_counts" :key="row.package_id" class="card border-0 shadow-sm rounded-4 mb-3">
         <div class="card-body">
          <div class="d-flex justify-content-between align-items-center mb-2">
            <div><span class="badge bg-primary fs-6 px-3 py-2">{{ row.name || 'Unknown Package' }}</span></div>
            <div class="text-end">
              <div class="fw-bold text-success">{{ Number(row.total_value).toLocaleString() }} USDT</div>
              <small class="text-muted">{{ row.active_count }} Activations</small>
            </div>
          </div>
          <div class="progress rounded-pill" style="height: 8px">
            <div
              class="progress-bar bg-success"
              :style="{ width: sharePercent(row) + '%' }"
            ></div>
          </div>
         </div>
        </div>
      </div>
    </div>

    <div v-if="resp && Number(resp.grand_total) > 0" class="card border-0 shadow-sm bg-dark text-white mt-4 rounded-4">
      <div class="card-body d-flex justify-content-between align-items-center">
        <span>Total Activation Revenue</span>
        <span class="h4 mb-0 text-success">{{ Number(resp.grand_total).toLocaleString() }} USDT</span>
      </div>
    </div>
  </DashboardLayout>
</template>
