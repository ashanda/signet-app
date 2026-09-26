<script setup>
// Ports earn/index.blade.php ("Earn History" — route('earn.history'), any
// authenticated). See earnlog_handler.go's earnHistoryHandler (GET
// /earn/history?date_from=&date_to=).
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { nullStr, phpDateTime } from '@/utils/format'

// GET filter form + ->paginate(10)->withQueryString(): both live in the URL.
const route = useRoute()
const router = useRouter()

const loadError = ref('')
const dateFrom = ref('')
const dateTo = ref('')
const earns = ref(null)

async function fetchData() {
  loadError.value = ''
  const q = route.query
  dateFrom.value = q.date_from || ''
  dateTo.value = q.date_to || ''
  try {
    const params = {}
    for (const key of ['page', 'date_from', 'date_to']) if (q[key]) params[key] = q[key]
    const { data } = await api.get('/earn/history', { params })
    earns.value = data.earns
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load earn history.'
  }
}

watch(() => route.query, fetchData, { immediate: true })

function applyFilter() {
  router.push({ path: route.path, query: { date_from: dateFrom.value, date_to: dateTo.value } })
}

function goToPage(page) {
  router.push({ path: route.path, query: { ...route.query, page } })
}
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Earn History</h1>
        </div>
      </div>
    </div>
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <!-- Date Range Filter -->
        <form method="GET" class="row g-2 align-items-end mb-3" @submit.prevent="applyFilter">
          <div class="col-auto">
            <label class="form-label mb-1 small text-muted">From</label>
            <input v-model="dateFrom" type="date" name="date_from" class="form-control" />
          </div>
          <div class="col-auto">
            <label class="form-label mb-1 small text-muted">To</label>
            <input v-model="dateTo" type="date" name="date_to" class="form-control" />
          </div>
          <div class="col-auto">
            <button class="btn btn-primary" type="submit">Filter</button>
            <RouterLink v-if="route.query.date_from || route.query.date_to" :to="{ name: 'earn.history' }" class="btn btn-outline-secondary">Clear</RouterLink>
          </div>
        </form>
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">Amount</th>
                <th class="border-bottom" scope="col">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="earn in earns?.data" :key="earn.id">
                <td>
                  <span class="badge bg-success">{{ earn.amount ?? 0 }} USDT {{ nullStr(earn.description) ? ' -' + nullStr(earn.description) : '' }}</span>
                </td>
                <td><span class="badge bg-info">{{ phpDateTime(earn.created_at) }}</span></td>
              </tr>
              <tr v-if="earns && !earns.data?.length">
                <td colspan="2" class="text-center text-muted">No earnings found</td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="earns" @change="goToPage" />
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
