<script setup>
// Ports company/leadership_bonus_log.blade.php ("Leadership Bonus Log").
// See leaderexecutive_handler.go's leadershipBonusLogHandler. Same
// date-filter/table skeleton as direct_share_log.blade.php (ui_spec.md
// notes these "log" pages look cloned from one template).
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
    const { data } = await api.get('/leadership-bonus-log', { params })
    startDate.value = data.start_date
    endDate.value = data.end_date
    pools.value = data.pools
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load leadership bonus log.'
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
      <h1 class="h4">Leadership Bonus Log - {{ startDate }} to {{ endDate }}</h1>
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
                <th class="border-bottom">Amount</th>
                <th class="border-bottom">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !pools?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No pools data found</td>
              </tr>
              <tr v-for="row in pools?.data" :key="row.id">
                <td>{{ row.user_name || '—' }}</td>
                <td>SIG-00{{ row.user_id }}</td>
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
