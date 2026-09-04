<script setup>
// Ports company/pending-activation.blade.php (route('company.pending.activation')).
// Same structure as new-activations.blade.php but sourced from ALL pending
// user_packages company-wide (not filtered by ref_id) — see
// dashboard_handler.go's companyPendingActivationHandler. The "Active"
// button posts to /active-package (NOT /company/new-active-package — that's
// the New Activations page's endpoint), per useActivePackage's documented
// default.
//
// NOTE: dashActivationRow's user_name/user_whatsapp/package_name fields are
// sql.NullString scanned straight into the JSON response (no `.String`
// flatten in the handler), so they arrive as {String,Valid} objects, not
// plain strings — nullStr() below unwraps that.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useActivePackage } from '@/composables/useActivePackage'

const { activate } = useActivePackage('/active-package')

const activations = ref(null)
const loading = ref(true)
const loadError = ref('')

function nullStr(v) {
  return v && typeof v === 'object' && v.Valid ? v.String : ''
}

async function fetchList(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/company/pending-activation', { params: { page } })
    activations.value = data.activations
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load pending activations.'
  } finally {
    loading.value = false
  }
}

async function onActive(row) {
  const ok = await activate(row.id)
  if (ok) fetchList(activations.value?.current_page || 1)
}

onMounted(() => fetchList())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Pending Activations</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">User name</th>
                <th class="border-bottom" scope="col">Whats app</th>
                <th class="border-bottom" scope="col">Package</th>
                <th class="border-bottom" scope="col">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !activations?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No pending activations found</td>
              </tr>
              <tr v-for="row in activations?.data" :key="row.id">
                <td>{{ nullStr(row.user_name) || 'Unknown User' }}</td>
                <td>{{ nullStr(row.user_whatsapp) }}</td>
                <td>{{ nullStr(row.package_name) || 'Unknown Package' }}</td>
                <td>
                  <button type="button" class="btn btn-primary active-package" @click="onActive(row)">
                    Active
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="activations" @change="(p) => fetchList(p)" />
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
