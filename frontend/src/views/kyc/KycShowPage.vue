<script setup>
// Ports kyc/show.blade.php (route('kyc.show') — the sidebar's KYC link for
// non-company users). See kyc_handler.go's kycIndexHandler(d, true): the
// user's own KYC record, the page-load SweetAlert, and — for the "Telegrame
// Join Link" panel shown once verified — one Telegram link per active
// package. The Verify/Unverify branch only applies to company users, who
// normally land on the KYC list instead.
import { ref, onMounted } from 'vue'
import Swal from 'sweetalert2'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { nullStr } from '@/utils/format'

const auth = useAuthStore()
const router = useRouter()

const flashSuccess = ref('')
const flashError = ref('')
const kycs = ref(null)
const telegramLinks = ref([])

const rows = () => (Array.isArray(kycs.value) ? kycs.value : kycs.value?.data || [])

async function fetchData() {
  try {
    const { data } = await api.get('/kyc/show')
    kycs.value = data.kycs
    telegramLinks.value = data.telegram_links || []
    if (data.alert) Swal.fire({ icon: data.alert.icon, title: data.alert.title, text: data.alert.text })
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Could not load your KYC record.'
  }
}

onMounted(fetchData)

// destroy redirects to the KYC list; verify/unverify redirect back here
// with a success flash.
async function destroy(kyc) {
  if (!window.confirm('Are you sure?')) return
  try {
    await api.delete(`/kyc/${kyc.id}`)
    router.push({ name: 'kyc.index' })
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Something went wrong'
  }
}

async function setVerified(kyc, verified) {
  flashSuccess.value = ''
  try {
    await api.post(`/kyc/${kyc.id}/${verified ? 'verify' : 'unverify'}`)
    flashSuccess.value = verified ? 'KYC verified successfully.' : 'KYC unverified successfully.'
    fetchData()
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Something went wrong'
  }
}

const docUrl = (path) => '/storage/' + path
</script>

<template>
  <DashboardLayout>
    <div v-if="flashSuccess" class="alert alert-success">{{ flashSuccess }}</div>
    <div v-if="flashError" class="alert alert-danger">{{ flashError }}</div>
    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">KYC</h1>
        </div>
      </div>
    </div>
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="row">
          <template v-for="kyc in rows()" :key="kyc.id">
            <div class="col-md-6 col-lg-6 mb-6">
              <div class="card border-0">
                <div class="card-body d-flex flex-column justify-content-between">
                  <!-- Header Section -->
                  <h5 class="card-title mb-3 text-primary fw-semibold">
                    <i class="fas fa-user me-2"></i>{{ kyc.full_name }}
                  </h5>
                  <div class="alert alert-info">
                    <!-- Personal Info -->
                    <ul class="list-unstyled mb-3 small">
                      <li><i class="fas fa-envelope me-2 text-secondary"></i><strong>Email:</strong> {{ kyc.email }}</li>
                      <li><i class="fab fa-whatsapp me-2 text-success"></i><strong>WhatsApp 1:</strong> {{ kyc.contact_number1 }}</li>
                      <li><i class="fab fa-whatsapp me-2 text-muted"></i><strong>WhatsApp 2:</strong> {{ nullStr(kyc.contact_number2) || '—' }}</li>
                      <li><i class="fas fa-map-marker-alt me-2 text-danger"></i><strong>Address:</strong> {{ kyc.address }}</li>
                      <li><i class="fab fa-telegram me-2 text-info"></i><strong>Telegram:</strong> {{ nullStr(kyc.telegram_username) || '—' }}</li>
                      <li>
                        <i class="fas fa-id-card me-2 text-dark"></i>
                        <strong>{{ String(kyc.document_type || '').toUpperCase() }}:</strong> {{ kyc.document_number }}
                      </li>
                    </ul>

                    <!-- Document Previews -->
                    <div class="mb-3">
                      <strong class="d-block mb-1 text-muted">Documents:</strong>
                      <div v-if="kyc.document_type === 'nic'" class="d-flex align-items-center gap-2">
                        <a v-if="nullStr(kyc.nic_front)" :href="docUrl(nullStr(kyc.nic_front))" target="_blank" data-bs-toggle="tooltip" title="NIC Front">
                          <img :src="docUrl(nullStr(kyc.nic_front))" width="50" class="rounded border" />
                        </a>
                        <a v-if="nullStr(kyc.nic_back)" :href="docUrl(nullStr(kyc.nic_back))" target="_blank" data-bs-toggle="tooltip" title="NIC Back">
                          <img :src="docUrl(nullStr(kyc.nic_back))" width="50" class="rounded border" />
                        </a>
                      </div>
                      <template v-else-if="kyc.document_type === 'passport'">
                        <a v-if="nullStr(kyc.passport_image)" :href="docUrl(nullStr(kyc.passport_image))" target="_blank" data-bs-toggle="tooltip" title="Passport">
                          <img :src="docUrl(nullStr(kyc.passport_image))" width="60" class="rounded border" />
                        </a>
                      </template>
                      <span v-else class="text-muted">No document uploaded</span>
                    </div>

                    <!-- Verification Status -->
                    <p class="mb-3">
                      <span class="badge" :class="kyc.is_verified ? 'bg-success' : 'bg-warning text-dark'">
                        {{ kyc.is_verified ? 'Verified' : 'Pending Verification' }}
                      </span>
                    </p>

                    <!-- Action Buttons -->
                    <div class="d-flex gap-2 mt-auto">
                      <template v-if="auth.role !== 'company'">
                        <template v-if="!kyc.is_verified">
                          <RouterLink :to="{ name: 'kyc.edit', params: { id: kyc.id } }" class="btn btn-sm btn-outline-primary" data-bs-toggle="tooltip" title="Edit">
                            <i class="fas fa-edit"></i>
                          </RouterLink>
                          <form @submit.prevent="destroy(kyc)">
                            <button class="btn btn-sm btn-outline-danger" data-bs-toggle="tooltip" title="Delete">
                              <i class="fas fa-trash"></i>
                            </button>
                          </form>
                        </template>
                        <span v-else class="text-success small">Already Verified</span>
                      </template>
                      <template v-else>
                        <form v-if="!kyc.is_verified" @submit.prevent="setVerified(kyc, true)">
                          <button type="submit" class="btn btn-sm btn-success" data-bs-toggle="tooltip" title="Verify">
                            <i class="fas fa-check-circle"></i>
                          </button>
                        </form>
                        <form v-else @submit.prevent="setVerified(kyc, false)">
                          <button type="submit" class="btn btn-sm btn-warning text-white" data-bs-toggle="tooltip" title="Unverify">
                            <i class="fas fa-times-circle"></i>
                          </button>
                        </form>
                      </template>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <div class="col-md-6">
              <div v-if="kyc.is_verified" class="card border-0">
                <div class="card-body">
                  <h5 class="card-title">
                    <i class="fab fa-telegram me-1"></i> Telegrame Join Link
                  </h5>
                  <div class="alert alert-danger d-flex align-items-start" role="alert">
                    <i class="fas fa-exclamation-triangle fa-lg me-2 mt-1"></i>
                    <div>
                      <strong>Important Warning:</strong> For your safety and to maintain the integrity of our platform,
                      <u>do not share your Telegram link (e.g., <code>t.me/username</code>) anywhere in the system</u>, including in your profile, KYC
                      form, or chat.

                      <br /><br />

                      Sharing direct Telegram links can lead to:
                      <ul class="mb-1 mt-1">
                        <li>Phishing or scam activities through unsolicited contact</li>
                        <li>Violation of our platform’s security and anti-spam policy</li>
                        <li>Increased risk of identity fraud or impersonation</li>
                      </ul>

                      <strong
                        >📌 If you are found sharing your Telegram link or using the platform to promote external channels, your account will be
                        <span class="text-uppercase text-info">immediately suspended without notice</span>.</strong
                      >
                      This action is irreversible, and no appeals will be accepted.

                      <br /><br />
                      We are committed to keeping all users safe. Please use the system responsibly.
                    </div>
                  </div>
                  <p class="card-text">
                    <a v-for="(link, i) in telegramLinks" :key="i" :href="'https://t.me/' + link" target="_blank" class="btn btn-primary">
                      <i class="fab fa-telegram"></i> Join Telegram
                    </a>
                  </p>
                </div>
              </div>
            </div>
          </template>
          <div v-if="kycs && !rows().length" class="col-12">
            <div v-if="auth.role !== 'company'" class="text-center my-4">
              <p class="text-muted">You have not submitted your KYC yet.</p>
              <RouterLink :to="{ name: 'kyc.create' }" class="btn btn-success">Submit KYC</RouterLink>
            </div>
            <div v-else class="alert alert-info">No KYC records found.</div>
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
