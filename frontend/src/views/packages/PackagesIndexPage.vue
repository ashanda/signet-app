<script setup>
// Ports packages/index.blade.php ("Package List"). See package_handler.go's
// packageIndexHandler/packageDestroyHandler.
//
// Deviations:
//  - GET /packages returns a plain array (`packages`), not a paginated
//    envelope — matches the Go handler exactly (no Paginator here).
//  - Status badge: Active = bg-success (green), Deactive = bg-secondary
//    (gray) — per ui_spec.md's badge table, NOT bg-danger.
//  - Delete uses useAlert().confirmDanger() (SweetAlert2) instead of the
//    original's native confirm(), per FRONTEND_CONVENTIONS.md's approved
//    consistency deviation.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'
import { useAlert } from '@/composables/useToast'

const { run } = useApiAction()
const { confirmDanger } = useAlert()

const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')
const packages = ref([])

async function fetchData() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/packages')
    packages.value = data.packages || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load packages.'
  } finally {
    loading.value = false
  }
}

async function deletePackage(pkg) {
  const result = await confirmDanger('Are you sure?', 'This package will be deleted permanently.')
  if (!result.isConfirmed) return
  const { ok, data } = await run(() => api.delete(`/packages/${pkg.id}`), {
    successMessage: 'Package deleted successfully.',
  })
  if (ok) {
    flashMessage.value = data?.message || 'Package deleted successfully.'
    fetchData()
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="success" :message="flashMessage" @close="flashMessage = ''" />
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="container">
      <h3 class="mb-4">Package List</h3>
      <router-link :to="{ name: 'packages.create' }" class="btn btn-primary mb-3">Add New Package</router-link>

      <table class="table table-bordered">
        <thead>
          <tr>
            <th>#</th>
            <th>Name</th>
            <th>Price</th>
            <th>Commission</th>
            <th>Status</th>
            <th>Rank</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="!loading && !packages.length">
            <td colspan="7" class="text-center text-muted py-4">No packages found</td>
          </tr>
          <tr v-for="(row, idx) in packages" :key="row.id">
            <td>{{ idx + 1 }}</td>
            <td>{{ row.name }}</td>
            <td>{{ row.price }}</td>
            <td>{{ row.commission }}</td>
            <td>
              <span class="badge" :class="row.status === 'active' ? 'bg-success' : 'bg-secondary'">
                {{ row.status === 'active' ? 'Active' : 'Deactive' }}
              </span>
            </td>
            <td>{{ row.rank?.String ?? row.rank ?? '' }}</td>
            <td>
              <router-link :to="{ name: 'packages.edit', params: { id: row.id } }" class="btn btn-sm btn-outline-primary">
                Edit
              </router-link>
              <button type="button" class="btn btn-sm btn-outline-danger" @click="deletePackage(row)">Delete</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </DashboardLayout>
</template>
