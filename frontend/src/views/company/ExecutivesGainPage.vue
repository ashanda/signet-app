<script setup>
// Ports executives/index.blade.php ("Executives Gain") — identical
// structure/behavior to leaders/index.blade.php with "Executive"/
// `executive_id` substituted for "Leader"/`leader_id` throughout, per
// ui_spec.md. See leaderexecutive_handler.go's executivesGainHandler /
// leaderExecCodeGainQuery.
//
// Same deviations as LeadersGainPage.vue: the handler returns no
// `binance_pay_id` (Binance ID column/copy button omitted, flagged for the
// backend team) and no executives options list for a `<select>`
// (implemented as an "Executive ID" number input bound to the
// `executive_id` query param instead).
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const loadError = ref('')

const executiveId = ref(route.query.executive_id || '')
const fromDate = ref(route.query.from || '')
const toDate = ref(route.query.to || '')
const executives = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (executiveId.value) params.executive_id = executiveId.value
    if (fromDate.value) params.from = fromDate.value
    if (toDate.value) params.to = toDate.value
    const { data } = await api.get('/executives/gain', { params })
    fromDate.value = data.from
    toDate.value = data.to
    executives.value = data.executives
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load executives gain.'
  } finally {
    loading.value = false
  }
}

function applySearch() {
  router.replace({
    query: {
      ...(executiveId.value ? { executive_id: executiveId.value } : {}),
      ...(fromDate.value ? { from: fromDate.value } : {}),
      ...(toDate.value ? { to: toDate.value } : {}),
    },
  })
  fetchData(1)
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Executives Gain</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="row g-2 align-items-end mb-4" @submit.prevent="applySearch">
          <div class="col-auto">
            <label class="form-label small mb-1">Executive ID</label>
            <input v-model="executiveId" type="number" min="1" class="form-control" placeholder="All Executives" />
          </div>
          <div class="col-auto">
            <label class="form-label small mb-1">From Date</label>
            <input v-model="fromDate" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <label class="form-label small mb-1">To Date</label>
            <input v-model="toDate" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <button type="submit" class="btn btn-primary">Search</button>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">Executive</th>
                <th class="border-bottom">SIGNET ID</th>
                <th class="border-bottom">Total Package Value</th>
                <th class="border-bottom">5% Gain</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !executives?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No executives found</td>
              </tr>
              <tr v-for="row in executives?.data" :key="row.id">
                <td>{{ row.name }}</td>
                <td>SIG-00{{ row.id }}</td>
                <td>${{ Number(row.total_package || 0).toFixed(2) }}</td>
                <td>${{ (Number(row.total_package || 0) * 0.05).toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="executives" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
