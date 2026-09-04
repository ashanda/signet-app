<script setup>
// Ports kyc/index.blade.php ("KYC" — route('kyc.index'), company: all
// records). See kyc_handler.go's kycIndexHandler (GET /kyc): for a company
// user this returns `{status, kycs: <paginated, unverified only>}` — the
// router (meta.roles: ['company']) already restricts this view to company,
// so only that branch is implemented here (the non-company own-record
// shape is KycShowPage.vue's job).
//
// `kycs.data` rows are RAW models.Kyc structs — contact_number2/nic_front/
// nic_back/passport_image are sql.NullString (no custom MarshalJSON), may
// arrive as {String,Valid} objects; nsGet handles that defensively, same
// pattern as PackagesIndexPage.vue's established
// `row.rank?.String ?? row.rank ?? ''`.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'

function nsGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}

const { run } = useApiAction()

const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')
const kycs = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/kyc', { params: { page } })
    kycs.value = data.kycs
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load KYC records.'
  } finally {
    loading.value = false
  }
}

function docUrl(path) {
  return path ? `/storage/${path}` : ''
}

async function verify(row) {
  const { ok, data } = await run(() => api.post(`/kyc/${row.id}/verify`), { successMessage: 'KYC verified successfully.' })
  if (ok) {
    flashMessage.value = data?.message || 'KYC verified successfully.'
    fetchData(kycs.value?.current_page || 1)
  }
}

async function unverify(row) {
  const { ok, data } = await run(() => api.post(`/kyc/${row.id}/unverify`), { successMessage: 'KYC unverified successfully.' })
  if (ok) {
    flashMessage.value = data?.message || 'KYC unverified successfully.'
    fetchData(kycs.value?.current_page || 1)
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="success" :message="flashMessage" @close="flashMessage = ''" />
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4 d-flex align-items-center justify-content-between">
      <h1 class="h4 mb-0">KYC</h1>
      <router-link :to="{ name: 'kyc.verified' }" class="btn btn-success">Verified KYC</router-link>
    </div>

    <div v-if="!loading && !kycs?.data?.length" class="alert alert-info">No KYC records found.</div>

    <div v-else class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">Full Name</th>
                <th class="border-bottom">Email</th>
                <th class="border-bottom">WhatsApp 1</th>
                <th class="border-bottom">WhatsApp 2</th>
                <th class="border-bottom">Address</th>
                <th class="border-bottom">Telegrame Username</th>
                <th class="border-bottom">Document Type</th>
                <th class="border-bottom">Document No</th>
                <th class="border-bottom">Documents</th>
                <th class="border-bottom">Verified</th>
                <th class="border-bottom">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in kycs?.data" :key="row.id">
                <td>{{ row.full_name }}</td>
                <td>{{ row.email }}</td>
                <td>{{ row.contact_number1 }}</td>
                <td>{{ nsGet(row.contact_number2) || '-' }}</td>
                <td>{{ row.address }}</td>
                <td>{{ row.telegram_username || '-' }}</td>
                <td>{{ String(row.document_type || '').toUpperCase() }}</td>
                <td>{{ row.document_number }}</td>
                <td>
                  <template v-if="row.document_type === 'nic' && (nsGet(row.nic_front) || nsGet(row.nic_back))">
                    <a v-if="nsGet(row.nic_front)" :href="docUrl(nsGet(row.nic_front))" target="_blank" class="me-2" title="NIC Front">
                      <i class="fas fa-id-card"></i>
                    </a>
                    <a v-if="nsGet(row.nic_back)" :href="docUrl(nsGet(row.nic_back))" target="_blank" title="NIC Back">
                      <i class="fas fa-id-card"></i>
                    </a>
                  </template>
                  <a v-else-if="row.document_type === 'passport' && nsGet(row.passport_image)" :href="docUrl(nsGet(row.passport_image))" target="_blank" title="Passport">
                    <i class="fas fa-passport"></i>
                  </a>
                  <span v-else class="text-muted">No document</span>
                </td>
                <td>
                  <span :class="row.is_verified ? 'badge-success' : 'badge-warning'">
                    {{ row.is_verified ? 'Verified' : 'Pending' }}
                  </span>
                </td>
                <td>
                  <button v-if="!row.is_verified" type="button" class="btn btn-sm btn-success" @click="verify(row)">Verify</button>
                  <button v-else type="button" class="btn btn-sm btn-warning" @click="unverify(row)">Unverify</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="kycs" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
