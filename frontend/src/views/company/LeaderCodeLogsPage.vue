<script setup>
// Ports leader_code_logs/index.blade.php ("Leader Code Logs"). See
// user_handler.go's codeLogsHandler("leader_code_logs", ...).
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

const search = ref(route.query.search || '')
const fromDate = ref(route.query.from || '')
const toDate = ref(route.query.to || '')
const logs = ref(null)

function fmtDateTime(v) {
  if (!v) return ''
  const s = String(v).replace('T', ' ')
  return s.slice(0, 16)
}

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (search.value) params.search = search.value
    if (fromDate.value) params.from = fromDate.value
    if (toDate.value) params.to = toDate.value
    const { data } = await api.get('/leader-code-logs', { params })
    logs.value = data.logs
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load leader code logs.'
  } finally {
    loading.value = false
  }
}

function applySearch() {
  router.replace({
    query: {
      ...(search.value ? { search: search.value } : {}),
      ...(fromDate.value ? { from: fromDate.value } : {}),
      ...(toDate.value ? { to: toDate.value } : {}),
    },
  })
  fetchData(1)
}

function clearSearch() {
  search.value = ''
  fromDate.value = ''
  toDate.value = ''
  router.replace({ query: {} })
  fetchData(1)
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4 d-flex align-items-center justify-content-between">
      <h1 class="h4 mb-0">Leader Code Logs</h1>
      <router-link :to="{ name: 'company.users' }" class="btn btn-outline-secondary">
        <i class="fas fa-arrow-left me-1"></i> Back to Find Users
      </router-link>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="row g-2 align-items-end mb-4" @submit.prevent="applySearch">
          <div class="col-auto">
            <label class="form-label small mb-1">Search</label>
            <input v-model="search" type="text" class="form-control" placeholder="Search by name or Signet ID" />
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
          <div v-if="search || fromDate || toDate" class="col-auto">
            <a href="#" class="btn btn-outline-dark" @click.prevent="clearSearch">Clear</a>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User</th>
                <th class="border-bottom">Old Leader</th>
                <th class="border-bottom">New Leader</th>
                <th class="border-bottom">Changed By</th>
                <th class="border-bottom">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !logs?.data?.length">
                <td colspan="5" class="text-center text-muted py-4">No leader changes found</td>
              </tr>
              <tr v-for="row in logs?.data" :key="row.id">
                <td>{{ row.user ? `${row.user.name} (${row.user.signet_id})` : 'N/A' }}</td>
                <td>{{ row.old_name || 'None' }}</td>
                <td>{{ row.new_name || 'None' }}</td>
                <td>{{ row.changed_by || 'N/A' }}</td>
                <td>{{ fmtDateTime(row.created_at) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="logs" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
