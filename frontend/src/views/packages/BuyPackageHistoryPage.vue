<script setup>
// Ports packages/buy-packageHistory.blade.php ("Buy Packages History" —
// route('buy.package.history'), any authenticated). See
// package_handler.go's buyPackageHistoryHandler (GET /buy-package-history).
//
// Deviation from ui_spec.md, verified against the actual Go handler
// (trusted over prose): ui_spec.md documents a "Package" column showing
// the package NAME. The handler's query is `SELECT * FROM user_packages
// WHERE user_id = ?` with no join to `packages` — `row.package` is only
// the raw package id (string). Shown as "Package #{id}" here rather than
// fabricating a name; flagged for the backend team (would need a
// package-name join, or a public packages lookup endpoint non-company
// users can call).
//
// `packages` rows are RAW models.UserPackage structs — created_at is
// sql.NullTime (no custom MarshalJSON), may arrive as a {Time,Valid}
// object; handled defensively like PackagesIndexPage.vue's established
// `row.rank?.String ?? row.rank ?? ''` pattern.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAuthStore } from '@/store/auth'

function ntGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.Time : ''
  return v
}

const authStore = useAuthStore()

const loading = ref(true)
const loadError = ref('')
const packages = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/buy-package-history', { params: { page } })
    packages.value = data.packages
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load buy package history.'
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4 d-flex align-items-center justify-content-between">
      <h1 class="h4 mb-0">Buy Packages History</h1>
      <router-link :to="{ name: 'buy.package' }" class="btn btn-primary">Buy Package</router-link>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User name</th>
                <th class="border-bottom">Package</th>
                <th class="border-bottom">Earn</th>
                <th class="border-bottom">Buy Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !packages?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No records found</td>
              </tr>
              <tr v-for="row in packages?.data" :key="row.id">
                <td>{{ authStore.user?.name }}</td>
                <td>Package #{{ row.package }}</td>
                <td><span class="badge bg-success">{{ Number(row.earn || 0).toFixed(2) }} USDT</span></td>
                <td><span class="badge bg-info">{{ ntGet(row.created_at) }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="packages" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
