<script setup>
// Ports packages/index.blade.php ("Package List"). See package_handler.go's
// packageIndexHandler/packageDestroyHandler. GET /packages returns a plain
// array (no pagination), as the original's Package::all(). Delete asks with
// the browser's confirm(); create/update/delete land here with the
// original's success flash.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import { useFlashStore } from '@/store/flash'
import { nullStr } from '@/utils/format'

const flashStore = useFlashStore()
const { success } = flashStore.take()

const flashSuccess = ref(success)
const flashError = ref('')
const packages = ref([])

async function fetchData() {
  try {
    const { data } = await api.get('/packages')
    packages.value = data.packages || []
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Could not load packages.'
  }
}

async function deletePackage(pkg) {
  if (!window.confirm('Delete this package?')) return
  flashSuccess.value = ''
  flashError.value = ''
  try {
    const { data } = await api.delete(`/packages/${pkg.id}`)
    flashSuccess.value = data?.message || 'Package deleted successfully.'
    fetchData()
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Something went wrong'
  }
}

const ucfirst = (s) => (s ? s.charAt(0).toUpperCase() + s.slice(1) : '')

onMounted(fetchData)
</script>

<template>
  <DashboardLayout>
    <div class="container">
      <h3 class="mb-4">Package List</h3>
      <RouterLink :to="{ name: 'packages.create' }" class="btn btn-primary mb-3">Add New Package</RouterLink>

      <div v-if="flashSuccess" class="alert alert-success">{{ flashSuccess }}</div>
      <div v-if="flashError" class="alert alert-danger">{{ flashError }}</div>

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
          <tr v-for="(pkg, key) in packages" :key="pkg.id">
            <td>{{ key + 1 }}</td>
            <td>{{ pkg.name }}</td>
            <td>{{ pkg.price }}</td>
            <td>{{ pkg.commission }}</td>
            <td>
              <span class="badge" :class="'bg-' + (pkg.status === 'active' ? 'success' : 'secondary')">
                {{ ucfirst(pkg.status) }}
              </span>
            </td>
            <td>{{ nullStr(pkg.rank) }}</td>
            <td>
              <RouterLink :to="{ name: 'packages.edit', params: { id: pkg.id } }" class="btn btn-sm btn-outline-primary">Edit</RouterLink>
              <form class="d-inline" @submit.prevent="deletePackage(pkg)">
                <button class="btn btn-sm btn-outline-danger">Delete</button>
              </form>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </DashboardLayout>
</template>
