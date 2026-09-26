<script setup>
// Shared body of kyc/index.blade.php ("KYC") and kyc/verified.blade.php
// ("Verified KYC"), which differ only in the title and the "Verified KYC"
// button. See kyc_handler.go's kycIndexHandler/kycVerifiedHandler.
//
// Company users get the paginated list with Verify/Unverify; anyone else
// sees their own single record with Edit/Delete while it's unverified. Each
// load raises the page's SweetAlert (`alert` in the response), as the
// original's Alert::…() calls do. Verify/unverify/delete land back here
// with the original's success flash.
//
// The Verified column uses `badge-success`/`badge-warning`, classes nothing
// defines, so it renders as plain text — as in the original.
import { ref, onMounted } from 'vue'
import Swal from 'sweetalert2'
import api from '@/api/client'
import Paginator from '@/components/shared/Paginator.vue'
import { useAuthStore } from '@/store/auth'
import { useFlashStore } from '@/store/flash'
import { nullStr } from '@/utils/format'

const props = defineProps({
  verified: { type: Boolean, default: false },
})

const auth = useAuthStore()
const flashStore = useFlashStore()

const flashSuccess = ref(flashStore.take().success)
const flashError = ref('')
const kycs = ref(null)

// Company responses are paginated; a user's own KYC comes back as a plain array.
const rows = () => (Array.isArray(kycs.value) ? kycs.value : kycs.value?.data || [])

async function fetchData(page = 1) {
  try {
    const { data } = await api.get(props.verified ? '/kyc/verified' : '/kyc', { params: { page } })
    kycs.value = data.kycs
    if (data.alert) Swal.fire({ icon: data.alert.icon, title: data.alert.title, text: data.alert.text })
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Could not load KYC records.'
  }
}

onMounted(() => fetchData())

async function act(request, successMessage) {
  flashSuccess.value = ''
  flashError.value = ''
  try {
    const { data } = await request()
    flashSuccess.value = successMessage === undefined ? data?.message || '' : successMessage
    await fetchData(kycs.value?.current_page || 1)
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Something went wrong'
  }
}

const verify = (kyc) => act(() => api.post(`/kyc/${kyc.id}/verify`), 'KYC verified successfully.')
const unverify = (kyc) => act(() => api.post(`/kyc/${kyc.id}/unverify`), 'KYC unverified successfully.')

// destroy redirects to kyc.index with only a SweetAlert (no session flash),
// which kyc.index's own alert then replaces.
function destroy(kyc) {
  if (!window.confirm('Are you sure?')) return
  act(() => api.delete(`/kyc/${kyc.id}`), '')
}

const docUrl = (path) => '/storage/' + path
</script>

<template>
  <div v-if="flashSuccess" class="alert alert-success">{{ flashSuccess }}</div>
  <div v-if="flashError" class="alert alert-danger">{{ flashError }}</div>
  <div class="py-4">
    <div class="d-flex justify-content-between w-100 flex-wrap">
      <div class="mb-3 mb-lg-0">
        <h1 class="h4">{{ verified ? 'Verified KYC' : 'KYC' }}</h1>
      </div>
    </div>
  </div>
  <div class="card border-0 shadow mb-4">
    <div class="card-body">
      <div v-if="!verified" class="d-flex justify-content-end mb-3">
        <RouterLink :to="{ name: 'kyc.verified' }" class="btn btn-success">
          Verified KYC
        </RouterLink>
      </div>
      <div class="table-responsive">
        <template v-if="rows().length">
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
              <tr v-for="kyc in rows()" :key="kyc.id">
                <td>{{ kyc.full_name }}</td>
                <td>{{ kyc.email }}</td>
                <td>{{ kyc.contact_number1 }}</td>
                <td>{{ nullStr(kyc.contact_number2) || '-' }}</td>
                <td>{{ kyc.address }}</td>
                <td>{{ nullStr(kyc.telegram_username) || '-' }}</td>
                <td>{{ String(kyc.document_type || '').toUpperCase() }}</td>
                <td>{{ kyc.document_number }}</td>
                <td>
                  <template v-if="kyc.document_type === 'nic'">
                    <a v-if="nullStr(kyc.nic_front)" :href="docUrl(nullStr(kyc.nic_front))" target="_blank" title="NIC Front">
                      <i class="fas fa-id-card fa-lg text-primary me-2"></i>
                    </a>
                    <a v-if="nullStr(kyc.nic_back)" :href="docUrl(nullStr(kyc.nic_back))" target="_blank" title="NIC Back">
                      <i class="fas fa-id-card fa-lg text-primary me-2"></i>
                    </a>
                  </template>
                  <template v-else-if="kyc.document_type === 'passport'">
                    <a v-if="nullStr(kyc.passport_image)" :href="docUrl(nullStr(kyc.passport_image))" target="_blank" title="Passport">
                      <i class="fas fa-passport fa-lg text-info"></i>
                    </a>
                  </template>
                  <span v-else>No document</span>
                </td>
                <td>
                  <span v-if="kyc.is_verified" class="badge-success">Verified</span>
                  <span v-else class="badge-warning">Pending</span>
                </td>
                <td>
                  <template v-if="auth.role !== 'company'">
                    <template v-if="!kyc.is_verified">
                      <!-- Normal user: show Edit and Delete -->
                      <RouterLink :to="{ name: 'kyc.edit', params: { id: kyc.id } }" class="btn btn-sm btn-primary">Edit</RouterLink>
                      <form class="d-inline" @submit.prevent="destroy(kyc)">
                        <button class="btn btn-sm btn-danger">Delete</button>
                      </form>
                    </template>
                    <span v-else class="badge-success">N/A</span>
                  </template>
                  <template v-else>
                    <!-- Company user: show Verify / Unverify buttons -->
                    <form v-if="!kyc.is_verified" class="d-inline" @submit.prevent="verify(kyc)">
                      <button type="submit" class="btn btn-sm btn-success">Verify</button>
                    </form>
                    <form v-else class="d-inline" @submit.prevent="unverify(kyc)">
                      <button type="submit" class="btn btn-sm btn-warning">Unverify</button>
                    </form>
                  </template>
                </td>
              </tr>
            </tbody>
          </table>

          <!-- Only show pagination if it's a company user -->
          <div v-if="auth.role === 'company'" class="d-flex justify-content-center mt-3">
            <Paginator :pagination="kycs" @change="(p) => fetchData(p)" />
          </div>
        </template>
        <template v-else-if="kycs">
          <div v-if="auth.role !== 'company'" class="text-center my-4">
            <p class="text-muted">You have not submitted your KYC yet.</p>
            <RouterLink :to="{ name: 'kyc.create' }" class="btn btn-success">Submit KYC</RouterLink>
          </div>
          <div v-else class="alert alert-info">No KYC records found.</div>
        </template>
      </div>
    </div>
  </div>
</template>
