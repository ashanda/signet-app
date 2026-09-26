<script setup>
// Ports packages/buy-packageHistory.blade.php ("Buy Packages History" —
// route('buy.package.history'), any authenticated). See
// package_handler.go's buyPackageHistoryHandler (GET /buy-package-history),
// which joins each row's package name (`$package->userpackage->name`).
// The "User name" column is always the signed-in user — the original's
// query is scoped to auth()->id().
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAuthStore } from '@/store/auth'
import { nullStr, phpDateTime } from '@/utils/format'

const authStore = useAuthStore()

const loadError = ref('')
const packages = ref(null)

async function fetchData(page = 1) {
  loadError.value = ''
  try {
    const { data } = await api.get('/buy-package-history', { params: { page } })
    packages.value = data.packages
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load buy package history.'
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Buy Packages History</h1>
        </div>
      </div>
    </div>
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <RouterLink class="btn btn-primary mb-4" :to="{ name: 'buy.package' }">Buy Package</RouterLink>
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">User name</th>
                <th class="border-bottom" scope="col">Package</th>
                <th class="border-bottom" scope="col">Earn</th>
                <th class="border-bottom" scope="col">Buy Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in packages?.data" :key="row.id">
                <td>{{ authStore.user?.name || 'Unknown User' }}</td>
                <td>{{ nullStr(row.package_name) }}</td>
                <td><span class="badge bg-success">{{ row.earn ?? 0 }} USDT</span></td>
                <td><span class="badge bg-info">{{ phpDateTime(row.created_at) }}</span></td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="packages" @change="(p) => fetchData(p)" />
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
