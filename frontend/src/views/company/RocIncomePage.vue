<script setup>
// Ports company/roc_income.blade.php (route('company.roc')). See
// roc_handler.go's rocIncomeHandler/rocUpdateStatusHandler.
//
// Deviations from ui_spec.md's documented markup, verified against the
// actual blade source and the Go handler (trusted over ui_spec.md's prose):
//  - The summary row is ONE row of TWO cards, not three: "Per Week Total"
//    (with "5% Week Total" alongside it) and "Total Paying ROC" — the
//    "Balance Forward" tile is commented out in the original, omitted here
//    too.
//  - rocLogRow's user_name/created_at fields are sql.Null* scanned straight
//    into JSON (no `.String`/`.Time` flatten in the handler), so they
//    arrive as {String,Valid}/{Time,Valid} objects — nullStr()/nullTime()
//    unwrap that.
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useToast } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const { toastSuccess, toastError } = useToast()

const jobs = ref([])
const selectedJobId = ref('')
const weeklySummary = ref(null)
const logs = ref(null)
const loading = ref(true)
const loadError = ref('')

function nullStr(v) {
  return v && typeof v === 'object' && v.Valid ? v.String : ''
}

function dateOnly(v) {
  if (!v) return ''
  return String(v).slice(0, 10)
}

const jobFilter = ref(route.query.job_id || '')

async function fetchRoc(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (jobFilter.value) params.job_id = jobFilter.value
    const { data } = await api.get('/roc', { params })
    if (data.status === 'error') {
      loadError.value = data.message || 'No ROC job found.'
      jobs.value = []
      weeklySummary.value = null
      logs.value = null
      return
    }
    jobs.value = data.jobs || []
    selectedJobId.value = data.selected_job_id || ''
    weeklySummary.value = data.weekly_summary || null
    logs.value = data.roc_income_logs
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load ROC income.'
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  router.replace({ query: jobFilter.value ? { job_id: jobFilter.value } : {} })
  fetchRoc(1)
}

function clearFilter() {
  jobFilter.value = ''
  router.replace({ query: {} })
  fetchRoc(1)
}

const perWeekTotal = computed(() => Number(weeklySummary.value?.per_week_total || 0))
const fiveWeekTotal = computed(() => perWeekTotal.value * 0.05)
const totalPayingRoc = computed(() =>
  Number(weeklySummary.value?.total_amount || 0) - Number(weeklySummary.value?.balance_forward || 0)
)

async function onToggleStatus(row, event) {
  const checked = event.target.checked
  const newStatus = checked ? 'paid' : 'pending'
  try {
    const { data } = await api.post('/roc/status-update', { id: row.id, status: newStatus })
    if (data.success) {
      row.status = newStatus
      toastSuccess(`Status changed to ${newStatus.charAt(0).toUpperCase() + newStatus.slice(1)}`)
    } else {
      toastError(data.message || 'Failed to update status')
      event.target.checked = !checked
    }
  } catch {
    toastError('Server error. Try again.')
    event.target.checked = !checked
  }
}

onMounted(() => fetchRoc())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">ROC Income</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="mb-3" @submit.prevent="applyFilter">
          <div class="input-group">
            <select v-model="jobFilter" class="form-select">
              <option value="">-- Select Week --</option>
              <option v-for="job in jobs" :key="job.job_id" :value="job.job_id">
                {{ dateOnly(job.week_start) }} to {{ dateOnly(job.week_end) }}
              </option>
            </select>
            <button type="submit" class="btn btn-primary">Filter</button>
            <button v-if="jobFilter" type="button" class="btn btn-outline-secondary" @click="clearFilter">
              Clear
            </button>
          </div>
        </form>

        <div v-if="weeklySummary" class="row g-3 mb-4">
          <div class="col-12 col-md-6">
            <div class="p-3 rounded-4 border bg-light h-100">
              <div class="d-flex align-items-center justify-content-between">
                <div>
                  <div class="text-muted small fw-semibold">Per Week Total</div>
                  <div class="fs-4 fw-bold mt-1">USDT {{ perWeekTotal.toFixed(2) }}</div>
                </div>
                <div>
                  <div class="text-muted small fw-semibold">5% Week Total</div>
                  <div class="fs-4 fw-bold mt-1">USDT {{ fiveWeekTotal.toFixed(2) }}</div>
                </div>
                <div class="rounded-circle bg-white border d-flex align-items-center justify-content-center" style="width:46px;height:46px;">
                  <i class="fas fa-coins fs-5 text-primary"></i>
                </div>
              </div>
              <div class="mt-2 small text-muted">Total for selected week</div>
            </div>
          </div>

          <div class="col-12 col-md-6">
            <div class="p-3 rounded-4 border bg-light h-100">
              <div class="d-flex align-items-center justify-content-between">
                <div>
                  <div class="text-muted small fw-semibold">Total Paying ROC</div>
                  <div class="fs-4 fw-bold mt-1">USDT {{ totalPayingRoc.toFixed(2) }}</div>
                </div>
                <div class="rounded-circle bg-white border d-flex align-items-center justify-content-center" style="width:46px;height:46px;">
                  <i class="fas fa-arrow-trend-down fs-5 text-danger"></i>
                </div>
              </div>
              <div class="mt-2 small text-muted">Selected Week Total Paying</div>
            </div>
          </div>
        </div>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User name</th>
                <th class="border-bottom">SIGNET ID</th>
                <th class="border-bottom">Binance ID</th>
                <th class="border-bottom">WhatsApp</th>
                <th class="border-bottom">Earnings</th>
                <th class="border-bottom">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !logs?.data?.length">
                <td colspan="6" class="text-center text-muted py-4">No users found</td>
              </tr>
              <tr v-for="row in logs?.data" :key="row.id">
                <td>{{ nullStr(row.user_name) || '—' }}</td>
                <td>SIG-00{{ row.user_id }}</td>
                <td>{{ nullStr(row.binance_pay_id) }}</td>
                <td>{{ nullStr(row.whatsapp_number) }}</td>
                <td>{{ Number(row.amount || 0).toFixed(2) }}</td>
                <td>
                  <div class="form-check form-switch">
                    <input
                      type="checkbox"
                      class="form-check-input toggle-status"
                      :checked="row.status === 'paid'"
                      @change="onToggleStatus(row, $event)"
                    />
                    <label class="form-check-label">{{ row.status ? row.status.charAt(0).toUpperCase() + row.status.slice(1) : '' }}</label>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="logs" @change="(p) => fetchRoc(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
