<script setup>
// Ports company/leadership_bonus_log.blade.php. See
// leaderexecutive_handler.go's leadershipBonusLogHandler. The heading shows the effective range (defaulting
// to last month); the inputs only show what was actually requested, and
// "Clear" only appears when a date was given — as in the original GET form.
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const route = useRoute()
const router = useRouter()

const loadError = ref('')
const startInput = ref('')
const endInput = ref('')
const startDate = ref('')
const endDate = ref('')
const pools = ref(null)

async function fetchData() {
  loadError.value = ''
  const q = route.query
  startInput.value = q.start_date || ''
  endInput.value = q.end_date || ''
  try {
    const params = {}
    for (const key of ['page', 'start_date', 'end_date']) if (q[key]) params[key] = q[key]
    const { data } = await api.get('/leadership-bonus-log', { params })
    startDate.value = data.start_date
    endDate.value = data.end_date
    pools.value = data.pools
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load leadership bonus log.'
  }
}

watch(() => route.query, fetchData, { immediate: true })

function applyFilter() {
  router.push({ path: route.path, query: { start_date: startInput.value, end_date: endInput.value } })
}

// ->paginate(20)->withQueryString(): page links keep the filter.
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
          <h1 class="h4">Leadership Bonus Log - {{ startDate }} to {{ endDate }}</h1>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <!-- Search Form -->
        <form method="GET" class="mb-3" @submit.prevent="applyFilter">
          <div class="row g-2">
            <div class="col-md-4">
              <input v-model="startInput" type="date" name="start_date" class="form-control" />
            </div>
            <div class="col-md-4">
              <input v-model="endInput" type="date" name="end_date" class="form-control" />
            </div>
            <div class="col-md-4 d-flex">
              <button type="submit" class="btn btn-primary me-2">
                Filter
              </button>
              <RouterLink v-if="route.query.start_date || route.query.end_date" to="/leadership-bonus-log" class="btn btn-outline-secondary">
                Clear
              </RouterLink>
            </div>
          </div>
        </form>

        <div class="table-responsive mt-4">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">User name</th>
                <th class="border-bottom" scope="col">SIG ID</th>
                <th class="border-bottom" scope="col">Amount</th>
                <th class="border-bottom" scope="col">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="pool in pools?.data" :key="pool.id">
                <td>{{ pool.user_name }}</td>
                <td>{{ 'SIG-00' + pool.user_id }}</td>
                <td>{{ pool.amount }}</td>
                <td>{{ pool.created_at ? String(pool.created_at).slice(0, 10) : '' }}</td>
              </tr>
              <tr v-if="pools && !pools.data?.length">
                <td colspan="4" class="text-center text-muted">No pools data found</td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="pools" @change="goToPage" />
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
