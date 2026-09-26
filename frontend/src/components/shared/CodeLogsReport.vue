<script setup>
// Shared body of leader_code_logs/index.blade.php and
// executive_code_logs/index.blade.php, identical apart from the
// Leader/Executive wording. See user_handler.go's codeLogsHandler.
// The search form is a GET form and pagination appends the query
// (->appends($request->query())), so both drive the URL query here.
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { phpDateTime } from '@/utils/format'

const props = defineProps({
  kind: { type: String, required: true }, // 'leader' | 'executive'
})

const Label = props.kind === 'leader' ? 'Leader' : 'Executive'
const endpoint = props.kind === 'leader' ? '/leader-code-logs' : '/executive-code-logs'

const route = useRoute()
const router = useRouter()

const loadError = ref('')
const search = ref('')
const fromDate = ref('')
const toDate = ref('')
const logs = ref(null)

async function fetchData() {
  loadError.value = ''
  const q = route.query
  search.value = q.search || ''
  fromDate.value = q.from || ''
  toDate.value = q.to || ''
  try {
    const params = {}
    for (const key of ['page', 'search', 'from', 'to']) if (q[key]) params[key] = q[key]
    const { data } = await api.get(endpoint, { params })
    logs.value = data.logs
  } catch (err) {
    loadError.value = err?.response?.data?.message || `Could not load ${Label.toLowerCase()} code logs.`
  }
}

watch(() => route.query, fetchData, { immediate: true })

function submit() {
  router.push({ path: route.path, query: { search: search.value, from: fromDate.value, to: toDate.value } })
}

function goToPage(page) {
  router.push({ path: route.path, query: { ...route.query, page } })
}
</script>

<template>
  <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

  <div class="py-4">
    <div class="d-flex justify-content-between w-100 flex-wrap">
      <div class="mb-3 mb-lg-0">
        <h1 class="h4">{{ Label }} Code Logs</h1>
      </div>
      <div>
        <RouterLink :to="{ name: 'company.users' }" class="btn btn-outline-secondary">
          <i class="bi bi-arrow-left me-1"></i>Back to Find Users
        </RouterLink>
      </div>
    </div>
  </div>

  <div class="card border-0 shadow mb-4">
    <div class="card-body">
      <form method="GET" class="mb-3" @submit.prevent="submit">
        <div class="row g-2">
          <div class="col-12 col-md-4">
            <label>User</label>
            <input v-model="search" type="text" name="search" class="form-control" placeholder="Search by name or Signet ID" />
          </div>
          <div class="col-6 col-md-3">
            <label>From</label>
            <input v-model="fromDate" type="date" name="from" class="form-control" />
          </div>
          <div class="col-6 col-md-3">
            <label>To</label>
            <input v-model="toDate" type="date" name="to" class="form-control" />
          </div>
          <div class="col-12 col-md-2 d-flex align-items-end">
            <button class="btn btn-primary w-100" type="submit">Search</button>
          </div>
        </div>
      </form>

      <div class="table-responsive">
        <table class="table align-items-center table-flush">
          <thead class="thead-light">
            <tr>
              <th class="border-bottom" scope="col">User</th>
              <th class="border-bottom" scope="col">Old {{ Label }}</th>
              <th class="border-bottom" scope="col">New {{ Label }}</th>
              <th class="border-bottom" scope="col">Changed By</th>
              <th class="border-bottom" scope="col">Date</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="log in logs?.data" :key="log.id">
              <td>{{ log.user?.name ?? 'N/A' }} ({{ log.user ? 'SIG-00' + log.user.id : 'N/A' }})</td>
              <td>{{ log.old_name || 'None' }}</td>
              <td>{{ log.new_name || 'None' }}</td>
              <td>{{ log.changed_by || 'N/A' }}</td>
              <td>{{ phpDateTime(log.created_at, 16) }}</td>
            </tr>
            <tr v-if="logs && !logs.data?.length">
              <td colspan="5" class="text-center text-muted">No {{ Label.toLowerCase() }} changes found</td>
            </tr>
          </tbody>
        </table>

        <div class="d-flex justify-content-center">
          <Paginator :pagination="logs" @change="goToPage" />
        </div>
      </div>
    </div>
  </div>
</template>
