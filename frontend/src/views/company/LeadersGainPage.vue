<script setup>
// Ports leaders/index.blade.php ("Leaders Gain"). See
// leaderexecutive_handler.go's leadersGainHandler /
// leaderExecCodeGainQuery.
//
// Deviations from ui_spec.md's documented markup, verified against the
// actual Go handler (trusted over prose):
//  - The handler's row query only selects `u.id, u.name, u.email` plus the
//    two computed totals — no `binance_pay_id` is returned, so the
//    "Binance ID" column + copy-to-clipboard button ui_spec.md documents
//    cannot be rendered from real data and is omitted here (flagged for
//    the backend team) rather than the useToast()/copy behavior being
//    faked against a value that doesn't exist.
//  - The Leader filter is documented as a `<select>` populated from
//    "$leaders", but the handler returns no such options list in its
//    response (only the already-paginated report rows) — implemented here
//    as a plain "Leader ID" number input bound to the `leader_id` query
//    param the handler actually reads, instead of fabricating a select
//    with no backing data.
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

const leaderId = ref(route.query.leader_id || '')
const fromDate = ref(route.query.from || '')
const toDate = ref(route.query.to || '')
const leaders = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (leaderId.value) params.leader_id = leaderId.value
    if (fromDate.value) params.from = fromDate.value
    if (toDate.value) params.to = toDate.value
    const { data } = await api.get('/leaders/gain', { params })
    fromDate.value = data.from
    toDate.value = data.to
    leaders.value = data.leaders
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load leaders gain.'
  } finally {
    loading.value = false
  }
}

function applySearch() {
  router.replace({
    query: {
      ...(leaderId.value ? { leader_id: leaderId.value } : {}),
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
      <h1 class="h4">Leaders Gain</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="row g-2 align-items-end mb-4" @submit.prevent="applySearch">
          <div class="col-auto">
            <label class="form-label small mb-1">Leader ID</label>
            <input v-model="leaderId" type="number" min="1" class="form-control" placeholder="All Leaders" />
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
                <th class="border-bottom">Leader</th>
                <th class="border-bottom">SIGNET ID</th>
                <th class="border-bottom">Total Package Value</th>
                <th class="border-bottom">5% Gain</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !leaders?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No leaders found</td>
              </tr>
              <tr v-for="row in leaders?.data" :key="row.id">
                <td>{{ row.name }}</td>
                <td>SIG-00{{ row.id }}</td>
                <td>${{ Number(row.total_package || 0).toFixed(2) }}</td>
                <td>${{ (Number(row.total_package || 0) * 0.05).toFixed(2) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="leaders" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
