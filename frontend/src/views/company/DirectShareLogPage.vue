<script setup>
// Ports company/direct_share_log.blade.php ("Globle Direct Share Log" —
// typo reproduced verbatim per ui_spec.md). See
// directshare_handler.go's directShareLogHandler.
//
// Deviation from ui_spec.md's documented markup, verified against the
// actual Go handler (trusted over prose): ui_spec.md documents a "Binance
// Pay ID" column (with a copy-to-clipboard icon button). The handler's SQL
// only selects `gswl.user_id, gswl.amount, gswl.description,
// gswl.created_at` plus the joined `user_name` — no `binance_pay_id` is
// returned, so that column/button is omitted here (rendered as "—") rather
// than fabricated. Flagged for the backend team.
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

const startDate = ref(route.query.start_date || '')
const endDate = ref(route.query.end_date || '')
const pools = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (startDate.value) params.start_date = startDate.value
    if (endDate.value) params.end_date = endDate.value
    const { data } = await api.get('/direct-share-log', { params })
    startDate.value = data.start_date
    endDate.value = data.end_date
    pools.value = data.pools
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load direct share log.'
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  router.replace({ query: { ...(startDate.value ? { start_date: startDate.value } : {}), ...(endDate.value ? { end_date: endDate.value } : {}) } })
  fetchData(1)
}

function clearFilter() {
  startDate.value = ''
  endDate.value = ''
  router.replace({ query: {} })
  fetchData(1)
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Globle Direct Share Log - {{ startDate }} to {{ endDate }}</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="row g-2 align-items-end mb-4" @submit.prevent="applyFilter">
          <div class="col-auto">
            <label class="form-label small mb-1">From Date</label>
            <input v-model="startDate" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <label class="form-label small mb-1">To Date</label>
            <input v-model="endDate" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <button type="submit" class="btn btn-primary">Filter</button>
          </div>
          <div v-if="startDate || endDate" class="col-auto">
            <a href="#" class="btn btn-outline-dark" @click.prevent="clearFilter">Clear</a>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User name</th>
                <th class="border-bottom">SIG ID</th>
                <th class="border-bottom">Binance Pay ID</th>
                <th class="border-bottom">Amount</th>
                <th class="border-bottom">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !pools?.data?.length">
                <td colspan="5" class="text-center text-muted py-4">No pools data found</td>
              </tr>
              <tr v-for="row in pools?.data" :key="row.id">
                <td>{{ row.user_name || '—' }}</td>
                <td>SIG-00{{ row.user_id }}</td>
                <td>—</td>
                <td>{{ Number(row.amount || 0).toFixed(2) }}</td>
                <td>{{ row.created_at ? String(row.created_at).slice(0, 10) : '' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="pools" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
