<script setup>
// Ports earn/index.blade.php ("Earn History" — route('earn.history'), any
// authenticated). See earnlog_handler.go's earnHistoryHandler (GET
// /earn/history?date_from=&date_to=): rows are RAW models.EarnLog structs
// (not hand-built maps), so `description`/`created_at` (sql.NullString/
// sql.NullTime, no custom MarshalJSON in the Go stdlib) serialize as
// {String,Valid}/{Time,Valid} objects, not plain values — same gotcha
// already handled defensively in PackagesIndexPage.vue's `row.rank?.String
// ?? row.rank ?? ''` pattern; nsGet/ntGet below generalize that.
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

function nsGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}
function ntGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.Time : ''
  return v
}

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const loadError = ref('')

const dateFrom = ref(route.query.date_from || '')
const dateTo = ref(route.query.date_to || '')
const earns = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (dateFrom.value) params.date_from = dateFrom.value
    if (dateTo.value) params.date_to = dateTo.value
    const { data } = await api.get('/earn/history', { params })
    earns.value = data.earns
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load earn history.'
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

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Earn History</h1>
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
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">Amount</th>
                <th class="border-bottom">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !earns?.data?.length">
                <td colspan="2" class="text-center text-muted py-4">No earnings found</td>
              </tr>
              <tr v-for="row in earns?.data" :key="row.id">
                <td>
                  <span class="badge bg-success">{{ row.amount }} USDT{{ nsGet(row.description) ? ' -' + nsGet(row.description) : '' }}</span>
                </td>
                <td><span class="badge bg-info">{{ ntGet(row.created_at) }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="earns" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
