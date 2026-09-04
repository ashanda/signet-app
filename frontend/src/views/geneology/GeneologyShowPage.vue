<script setup>
// Ports geneology/show.blade.php (individual downline member detail —
// route('geneology.show', $userId)). See geneology_handler.go's
// viewGeneologyHandler (GET /geneology/{userId}).
//
// `user_package` rows are a hand-built map, but `activated_at`/
// `created_at` are assigned the RAW sql.NullTime value (`p.ActivatedAt`,
// not `.Time`) — no custom MarshalJSON on that stdlib type, so they may
// arrive as {Time,Valid} objects; ntGet handles both shapes.
import { ref, onMounted, watch } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const props = defineProps({
  userId: { type: [String, Number], required: true },
})

function ntGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.Time : ''
  return v
}

function ymd(v) {
  const t = ntGet(v)
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

    <div v-if="userdata" class="card border-0 shadow mb-4">
      <div class="card-header bg-primary text-white">
        <h5 class="mb-0">{{ userdata.name }}</h5>
      </div>
      <div class="card-body">
        <h6 class="mb-3">User Information</h6>
        <div class="table-responsive mb-4">
          <table class="table table-bordered">
            <tbody>
              <tr>
                <th style="width: 200px;">Name</th>
                <td>{{ userdata.name }}</td>
              </tr>
              <tr>
                <th>Signet ID</th>
                <td>{{ userdata.signet_id }}</td>
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
                <td>{{ ymd(userdata.created_at) || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <h6 class="mb-3">Referred By (Parent)</h6>
        <div v-if="parentData" class="table-responsive mb-4">
          <table class="table table-bordered">
            <tbody>
              <tr>
                <th style="width: 200px;">Name</th>
                <td>{{ parentData.name }}</td>
              </tr>
              <tr>
                <th>Email</th>
                <td>{{ parentData.email || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="fst-italic text-muted mb-4">No parent user found.</p>

        <h6 class="mb-3">User Packages</h6>
        <div v-if="userPackages.length" class="table-responsive">
          <table class="table table-bordered table-striped">
            <thead>
              <tr>
                <th>#</th>
                <th>Package Name</th>
                <th>Earnings</th>
                <th>Activated On</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(pkg, idx) in userPackages" :key="pkg.id">
                <td>{{ idx + 1 }}</td>
                <td>{{ pkg.package_name || 'N/A' }}</td>
                <td>{{ Number(pkg.earn || 0).toFixed(2) }}</td>
                <td>{{ ymd(pkg.activated_at) || '—' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else class="fst-italic text-muted mb-0">No packages assigned to this user.</p>
      </div>
    </div>
  </DashboardLayout>
</template>
