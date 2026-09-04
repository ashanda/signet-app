<script setup>
// Ports kyc/create.blade.php ("Submit Your KYC" — route('kyc.create') /
// route('kyc.store')). See kyc_handler.go's kycStoreHandler (POST /kyc,
// multipart/form-data): exact field names/document-type enum verified
// against the Go handler, not guessed.
import { computed, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'

const router = useRouter()

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
const nicFront = ref(null)
const nicBack = ref(null)
const passportImage = ref(null)
const errors = ref({})
const submitting = ref(false)

const errorMessages = computed(() => Object.values(errors.value).flat())

function onFileChange(e, target) {
  target.value = e.target.files?.[0] || null
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
    const { data } = await api.post('/kyc', fd)
    if (data.status === 'success') {
      router.push({ name: 'kyc.show' })
    }
  } catch (err) {
    if (err?.response?.status === 422 && err.response.data?.errors) {
      errors.value = err.response.data.errors
    } else {
      errors.value = { _general: [err?.response?.data?.message || 'Could not submit KYC.'] }
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <DashboardLayout>
    <div class="row justify-content-center">
      <div class="col-12 col-lg-8">
        <h1 class="h4 mb-4">Submit Your KYC</h1>

        <div v-if="errorMessages.length" class="alert alert-danger">
          <ul class="mb-0 ps-3">
            <li v-for="(msg, idx) in errorMessages" :key="idx">{{ msg }}</li>
          </ul>
        </div>

        <div class="card border-0 shadow">
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
                    <label class="form-label">NIC Front Image</label>
                    <input type="file" class="form-control" accept="image/*" @change="onFileChange($event, nicFront)" />
                  </div>
                  <div class="col-md-6">
                    <label class="form-label">NIC Back Image</label>
                    <input type="file" class="form-control" accept="image/*" @change="onFileChange($event, nicBack)" />
                  </div>
                </template>
                <template v-else-if="form.document_type === 'passport'">
                  <div class="col-md-6">
                    <label class="form-label">Passport Image</label>
                    <input type="file" class="form-control" accept="image/*" @change="onFileChange($event, passportImage)" />
                  </div>
                </template>
              </div>

              <button type="submit" class="btn btn-success mt-4" :disabled="submitting">Submit KYC</button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
