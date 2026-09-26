<script setup>
// Ports geneology/show.blade.php (individual downline member detail —
// route('geneology.show', $userId)). See geneology_handler.go's
// viewGeneologyHandler (GET /geneology/{userId}).
//
// "Activated On" shows each package row's created_at, as the original does.
import { ref, onMounted, watch } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { nullTime, numberFormat } from '@/utils/format'

const props = defineProps({
  userId: { type: [String, Number], required: true },
})

function ymd(v) {
  const t = nullTime(v)
  return t ? String(t).slice(0, 10) : ''
}

const loading = ref(true)
const loadError = ref('')
const userdata = ref(null)
const userPackages = ref([])
const parentData = ref(null)

async function fetchData() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get(`/geneology/${props.userId}`)
    userdata.value = data.userdata
    userPackages.value = data.user_package || []
    parentData.value = data.parent_data || null
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load this user.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
watch(() => props.userId, fetchData)
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />
    <div v-if="userdata" class="container">
      <div class="card shadow mt-4">
        <div class="card-header bg-primary text-white">
          <h4>{{ userdata.name }}</h4>
        </div>
        <div class="card-body">
          <!-- User Info -->
          <h5 class="mb-3">User Information</h5>
          <table class="table table-bordered">
            <tbody>
              <tr>
                <th>Name</th>
                <td>{{ userdata.name }}</td>
              </tr>
              <tr>
                <th>Signet ID</th>
                <td>{{ 'SIG-00' + userdata.id }}</td>
              </tr>
              <tr>
                <th>Email</th>
                <td>{{ userdata.email }}</td>
              </tr>
              <tr>
                <th>Mobile</th>
                <td>{{ userdata.whatsapp_number || 'N/A' }}</td>
              </tr>
              <tr>
                <th>Registered At</th>
                <td>{{ ymd(userdata.created_at) }}</td>
              </tr>
            </tbody>
          </table>

          <!-- Parent Info -->
          <template v-if="parentData">
            <h5 class="mt-4">Referred By (Parent)</h5>
            <table class="table table-bordered">
              <tbody>
                <tr>
                  <th>Name</th>
                  <td>{{ parentData.name }}</td>
                </tr>
                <tr>
                  <th>Email</th>
                  <td>{{ parentData.email }}</td>
                </tr>
              </tbody>
            </table>
          </template>
          <p v-else class="text-muted mt-3"><em>No parent user found.</em></p>

          <!-- Package Info -->
          <h5 class="mt-4">User Packages</h5>
          <table v-if="userPackages.length" class="table table-bordered table-striped">
            <thead>
              <tr>
                <th>#</th>
                <th>Package Name</th>
                <th>Earnings</th>
                <th>Activated On</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(pkg, key) in userPackages" :key="pkg.id">
                <td>{{ key + 1 }}</td>
                <td>{{ pkg.package_name || 'N/A' }}</td>
                <td>{{ numberFormat(pkg.earn, 2) }}</td>
                <td>{{ ymd(pkg.created_at) }}</td>
              </tr>
            </tbody>
          </table>
          <p v-else class="text-muted"><em>No packages assigned to this user.</em></p>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
