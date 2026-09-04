<script setup>
// Ports kyc/show.blade.php ("KYC" card-grid variant — route('kyc.show'),
// non-company: own KYC). See kyc_handler.go's kycIndexHandler (also
// mounted at GET /kyc/show — same handler, "kyc.show mirrors kyc.index's
// logic exactly" per its own doc comment): for a non-company user this
// returns `{status, alert, message, kycs: [0 or 1 record]}` (`is_verified`
// determines the flashed `alert`/`message` pair, surfaced here via
// FlashAlert as a small enhancement over the bare card-grid ui_spec.md
// describes).
//
// Per ui_spec.md's explicitly-preserved quirk: the "Verified KYC" button
// is rendered regardless of role (only meaningfully clickable for company,
// who wouldn't normally land on this view) — reproduced verbatim, not
// "fixed".
//
// DEVIATION (flagged, not fabricated): ui_spec.md documents a "Telegrame
// Join Link" bonus panel on a verified record with one "Join Telegram"
// button per user package's `telegram_link`. kycIndexHandler's response
// carries no such data (models.Kyc has no telegram_link/user-packages
// join), and no endpoint reachable by a non-company user returns a
// package's telegram_link (GET /packages is company-only; GET
// /buy-package-history's user_packages rows carry only the raw package id,
// unjoined). The compliance warning copy is reproduced as a best-effort
// paraphrase (ui_spec.md describes its existence/tone but not its exact
// wording, so the literal original string isn't available to port
// verbatim) with a note in place of fabricated per-package buttons/links.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAlert } from '@/composables/useToast'

function nsGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}

const { confirmDanger, alertSuccess, alertError } = useAlert()

const loading = ref(true)
const loadError = ref('')
const alertType = ref('')
const alertMessage = ref('')
const kycs = ref([])

async function fetchData() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/kyc/show')
    kycs.value = data.kycs || []
    alertType.value = data.alert === 'success' ? 'success' : data.alert === 'warning' ? 'danger' : 'info'
    alertMessage.value = data.alert === 'info' ? '' : data.message || ''
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load your KYC record.'
  } finally {
    loading.value = false
  }
}

function docUrl(path) {
  return path ? `/storage/${path}` : ''
}

async function deleteKyc(row) {
  const result = await confirmDanger('Are you sure?', 'This KYC record will be deleted permanently.')
  if (!result.isConfirmed) return
  try {
    const { data } = await api.delete(`/kyc/${row.id}`)
    await alertSuccess(data.message || 'KYC deleted.')
    fetchData()
  } catch (err) {
    await alertError(err?.response?.data?.message || 'Could not delete KYC.')
  }
}

onMounted(fetchData)
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />
    <FlashAlert :type="alertType" :message="alertMessage" @close="alertMessage = ''" />

    <div class="py-4 d-flex align-items-center justify-content-between">
      <h1 class="h4 mb-0">KYC</h1>
      <router-link :to="{ name: 'kyc.verified' }" class="btn btn-success">Verified KYC</router-link>
    </div>

    <div v-if="!loading && !kycs.length" class="text-center py-5">
      <p class="text-muted mb-3">You have not submitted your KYC yet.</p>
      <router-link :to="{ name: 'kyc.create' }" class="btn btn-primary">Submit KYC</router-link>
    </div>

    <div v-else class="row g-4">
      <div v-for="row in kycs" :key="row.id" class="col-12 col-md-6">
        <div class="card border-0 shadow h-100">
          <div class="card-header bg-white">
            <h5 class="mb-0">{{ row.full_name }}</h5>
          </div>
          <div class="card-body">
            <div class="alert alert-info mb-3">
              <div><i class="fas fa-envelope me-2"></i>{{ row.email }}</div>
              <div><i class="fas fa-phone me-2"></i>WhatsApp 1: {{ row.contact_number1 }}</div>
              <div><i class="fas fa-phone me-2"></i>WhatsApp 2: {{ nsGet(row.contact_number2) || '-' }}</div>
              <div><i class="fas fa-map-marker-alt me-2"></i>{{ row.address }}</div>
              <div><i class="fab fa-telegram me-2"></i>{{ row.telegram_username || '-' }}</div>
              <div><i class="fas fa-id-card me-2"></i>{{ String(row.document_type || '').toUpperCase() }}: {{ row.document_number }}</div>
            </div>

            <div class="d-flex flex-wrap gap-2 mb-3">
              <template v-if="row.document_type === 'nic'">
                <a v-if="nsGet(row.nic_front)" :href="docUrl(nsGet(row.nic_front))" target="_blank">
                  <img :src="docUrl(nsGet(row.nic_front))" alt="NIC Front" class="img-thumbnail" style="max-width: 140px;" />
                </a>
                <a v-if="nsGet(row.nic_back)" :href="docUrl(nsGet(row.nic_back))" target="_blank">
                  <img :src="docUrl(nsGet(row.nic_back))" alt="NIC Back" class="img-thumbnail" style="max-width: 140px;" />
                </a>
              </template>
              <a v-else-if="row.document_type === 'passport' && nsGet(row.passport_image)" :href="docUrl(nsGet(row.passport_image))" target="_blank">
                <img :src="docUrl(nsGet(row.passport_image))" alt="Passport" class="img-thumbnail" style="max-width: 140px;" />
              </a>
              <span v-if="!nsGet(row.nic_front) && !nsGet(row.nic_back) && !nsGet(row.passport_image)" class="text-muted">No document</span>
            </div>

            <span class="badge mb-3" :class="row.is_verified ? 'bg-success' : 'bg-warning text-dark'">
              {{ row.is_verified ? 'Verified' : 'Pending Verification' }}
            </span>

            <div v-if="!row.is_verified" class="d-flex gap-2">
              <router-link :to="{ name: 'kyc.edit', params: { id: row.id } }" class="btn btn-primary btn-sm">Edit</router-link>
              <button type="button" class="btn btn-danger btn-sm" @click="deleteKyc(row)">Delete</button>
            </div>
            <p v-else class="text-muted mb-0">Already Verified</p>

            <div v-if="row.is_verified" class="card border-0 bg-light mt-3">
              <div class="card-body">
                <h6 class="mb-2"><i class="fab fa-telegram me-2 text-primary"></i>Telegrame Join Link</h6>
                <div class="alert alert-danger py-2 small mb-2">
                  Do not share your Telegram group/channel links publicly. These links are for your
                  personal use only — sharing them outside your account may result in account
                  suspension per our community policy.
                </div>
                <p class="small text-muted mb-0">
                  Package-specific Telegram join links aren't available from this build yet — contact
                  support for your invite link.
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
