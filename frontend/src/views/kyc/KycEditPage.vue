<script setup>
// Ports kyc/edit.blade.php ("Edit Your KYC" — route('kyc.edit', $kyc) /
// route('kyc.update', $kyc), PUT). See kyc_handler.go's kycEditHandler (GET
// /kyc/{id}/edit) and kycUpdateHandler (PUT /kyc/{id}, multipart/form-data
// — files optional, old file kept server-side when a field is omitted).
//
// GET /kyc/{id}/edit returns a RAW models.Kyc struct — contact_number2/
// nic_front/nic_back/passport_image are sql.NullString (no custom
// MarshalJSON), so they may arrive as {String,Valid} objects; nsGet below
// handles that, same pattern as PackagesIndexPage.vue's established
// `row.rank?.String ?? row.rank ?? ''`.
import { computed, reactive, ref, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'

const props = defineProps({
  id: { type: [String, Number], required: true },
})

const router = useRouter()

function nsGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}

const loading = ref(true)
const loadError = ref('')

const form = reactive({
  full_name: '',
  email: '',
  contact_number1: '',
  contact_number2: '',
  address: '',
  telegram_username: '',
  document_type: '',
  document_number: '',
})
const existingFiles = reactive({ nic_front: '', nic_back: '', passport_image: '' })
const nicFront = ref(null)
const nicBack = ref(null)
const passportImage = ref(null)
const errors = ref({})
const submitting = ref(false)

const errorMessages = computed(() => Object.values(errors.value).flat())

function fileUrl(path) {
  return path ? `/storage/${path}` : ''
}

function onFileChange(e, target) {
  target.value = e.target.files?.[0] || null
}

async function load() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get(`/kyc/${props.id}/edit`)
    const kyc = data.kyc
    form.full_name = kyc.full_name || ''
    form.email = kyc.email || ''
    form.contact_number1 = kyc.contact_number1 || ''
    form.contact_number2 = nsGet(kyc.contact_number2)
    form.address = kyc.address || ''
    form.telegram_username = kyc.telegram_username || ''
    form.document_type = kyc.document_type || ''
    form.document_number = kyc.document_number || ''
    existingFiles.nic_front = nsGet(kyc.nic_front)
    existingFiles.nic_back = nsGet(kyc.nic_back)
    existingFiles.passport_image = nsGet(kyc.passport_image)
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load KYC record.'
  } finally {
    loading.value = false
  }
}

async function submit() {
  errors.value = {}
  submitting.value = true
  const fd = new FormData()
  Object.entries(form).forEach(([key, value]) => fd.append(key, value ?? ''))
  if (form.document_type === 'nic') {
    if (nicFront.value) fd.append('nic_front', nicFront.value)
    if (nicBack.value) fd.append('nic_back', nicBack.value)
  } else if (form.document_type === 'passport') {
    if (passportImage.value) fd.append('passport_image', passportImage.value)
  }

  try {
    const { data } = await api.put(`/kyc/${props.id}`, fd)
    if (data.status === 'success') {
      router.push({ name: 'kyc.show' })
    }
  } catch (err) {
    if (err?.response?.status === 422 && err.response.data?.errors) {
      errors.value = err.response.data.errors
    } else {
      errors.value = { _general: [err?.response?.data?.message || 'Could not update KYC.'] }
    }
  } finally {
    submitting.value = false
  }
}

onMounted(load)
watch(() => props.id, load)
</script>

<template>
  <DashboardLayout>
    <div class="row justify-content-center">
      <div class="col-12 col-lg-8">
        <h1 class="h4 mb-4">Edit Your KYC</h1>

        <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>
        <div v-if="errorMessages.length" class="alert alert-danger">
          <ul class="mb-0 ps-3">
            <li v-for="(msg, idx) in errorMessages" :key="idx">{{ msg }}</li>
          </ul>
        </div>

        <div v-if="!loading" class="card border-0 shadow">
          <div class="card-body">
            <form enctype="multipart/form-data" @submit.prevent="submit">
              <div class="row g-3">
                <div class="col-md-6">
                  <label class="form-label">Full Name</label>
                  <input v-model="form.full_name" type="text" class="form-control" required />
                </div>
                <div class="col-md-6">
                  <label class="form-label">Email</label>
                  <input v-model="form.email" type="email" class="form-control" required />
                </div>
                <div class="col-md-6">
                  <label class="form-label">Contact Number 1</label>
                  <input v-model="form.contact_number1" type="text" class="form-control" required />
                </div>
                <div class="col-md-6">
                  <label class="form-label">Contact Number 2</label>
                  <input v-model="form.contact_number2" type="text" class="form-control" />
                </div>
                <div class="col-12">
                  <label class="form-label">Address</label>
                  <input v-model="form.address" type="text" class="form-control" required />
                </div>
                <div class="col-md-6">
                  <label class="form-label">Telegrame User Name</label>
                  <input v-model="form.telegram_username" type="text" class="form-control" required />
                </div>
                <div class="col-md-6">
                  <label class="form-label">Document Type</label>
                  <select v-model="form.document_type" class="form-select" required>
                    <option value="" disabled>Select document type</option>
                    <option value="nic">NIC</option>
                    <option value="passport">Passport</option>
                  </select>
                </div>
                <div class="col-md-6">
                  <label class="form-label">Document Number</label>
                  <input v-model="form.document_number" type="text" class="form-control" required />
                </div>

                <template v-if="form.document_type === 'nic'">
                  <div class="col-md-6">
                    <label class="form-label">
                      NIC Front Image
                      <a v-if="existingFiles.nic_front" :href="fileUrl(existingFiles.nic_front)" target="_blank" class="ms-1">(View)</a>
                    </label>
                    <input type="file" class="form-control" accept="image/*" @change="onFileChange($event, nicFront)" />
                  </div>
                  <div class="col-md-6">
                    <label class="form-label">
                      NIC Back Image
                      <a v-if="existingFiles.nic_back" :href="fileUrl(existingFiles.nic_back)" target="_blank" class="ms-1">(View)</a>
                    </label>
                    <input type="file" class="form-control" accept="image/*" @change="onFileChange($event, nicBack)" />
                  </div>
                </template>
                <template v-else-if="form.document_type === 'passport'">
                  <div class="col-md-6">
                    <label class="form-label">
                      Passport Image
                      <a v-if="existingFiles.passport_image" :href="fileUrl(existingFiles.passport_image)" target="_blank" class="ms-1">(View)</a>
                    </label>
                    <input type="file" class="form-control" accept="image/*" @change="onFileChange($event, passportImage)" />
                  </div>
                </template>
              </div>

              <button type="submit" class="btn btn-primary mt-4" :disabled="submitting">Update KYC</button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
