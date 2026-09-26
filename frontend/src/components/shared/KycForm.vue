<script setup>
// Shared body of kyc/create.blade.php ("Submit Your KYC") and
// kyc/edit.blade.php ("Edit Your KYC"). See kyc_handler.go's
// kycStoreHandler / kycEditHandler / kycUpdateHandler (multipart; on edit a
// file field left empty keeps the stored file).
//
// The NIC / passport upload blocks follow the Document Type select, as the
// original's toggleFields() script does. Validation errors are listed in
// the "Whoops!" box; a successful submit lands on the KYC list (kyc.index).
import { computed, reactive, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import { nullStr } from '@/utils/format'

const props = defineProps({
  id: { type: [String, Number], default: undefined },
})

const router = useRouter()
const isEdit = computed(() => props.id !== undefined && props.id !== null && props.id !== '')

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
const existing = reactive({ nic_front: '', nic_back: '', passport_image: '' })
const files = { nic_front: null, nic_back: null, passport_image: null }
const errors = ref([])
const submitting = ref(false)

onMounted(async () => {
  if (!isEdit.value) return
  try {
    const { data } = await api.get(`/kyc/${props.id}/edit`)
    const kyc = data.kyc
    for (const key of Object.keys(form)) form[key] = nullStr(kyc[key])
    for (const key of Object.keys(existing)) existing[key] = nullStr(kyc[key])
  } catch (err) {
    errors.value = [err?.response?.data?.message || 'Could not load KYC record.']
  }
})

function onFile(event, key) {
  files[key] = event.target.files?.[0] || null
}

async function submit() {
  errors.value = []
  submitting.value = true
  const fd = new FormData()
  Object.entries(form).forEach(([key, value]) => fd.append(key, value ?? ''))
  for (const [key, file] of Object.entries(files)) if (file) fd.append(key, file)
  try {
    if (isEdit.value) await api.put(`/kyc/${props.id}`, fd)
    else await api.post('/kyc', fd)
    router.push({ name: 'kyc.index' })
  } catch (err) {
    const errs = err?.response?.data?.errors
    errors.value = errs ? Object.values(errs).flat() : [err?.response?.data?.message || 'Could not save KYC.']
  } finally {
    submitting.value = false
  }
}

const url = (path) => '/storage/' + path
</script>

<template>
  <div class="container">
    <h2 class="mb-4">{{ isEdit ? 'Edit Your KYC' : 'Submit Your KYC' }}</h2>

    <div v-if="errors.length" class="alert alert-danger">
      <template v-if="isEdit"><strong>Whoops!</strong> Please fix the following issues:</template>
      <template v-else><strong>Whoops!</strong> There were problems with your input.<br /><br /></template>
      <ul>
        <li v-for="(error, i) in errors" :key="i">{{ error }}</li>
      </ul>
    </div>

    <form method="POST" enctype="multipart/form-data" @submit.prevent="submit">
      <div class="form-group mb-3">
        <label>Full Name</label>
        <input v-model="form.full_name" type="text" name="full_name" class="form-control" required />
      </div>

      <div class="form-group mb-3">
        <label>Email</label>
        <input v-model="form.email" type="email" name="email" class="form-control" required />
      </div>

      <div class="form-group mb-3">
        <label>{{ isEdit ? 'Contact Number 1 (WhatsApp)' : 'Contact Number 1 ' }}</label>
        <input v-model="form.contact_number1" type="text" name="contact_number1" class="form-control" required />
      </div>

      <div class="form-group mb-3">
        <label>Contact Number 2</label>
        <input v-model="form.contact_number2" type="text" name="contact_number2" class="form-control" />
      </div>

      <div class="form-group mb-3">
        <label>Address</label>
        <input v-model="form.address" type="text" name="address" class="form-control" required />
      </div>

      <div class="form-group mb-3">
        <label>Telegrame User Name</label>
        <input v-model="form.telegram_username" type="text" name="telegram_username" class="form-control" required />
      </div>

      <div class="form-group mb-3">
        <label>Document Type</label>
        <select id="documentType" v-model="form.document_type" name="document_type" class="form-control" required>
          <option value="">-- Select --</option>
          <option value="nic">NIC</option>
          <option value="passport">Passport</option>
        </select>
      </div>

      <div class="form-group mb-3">
        <label>Document Number</label>
        <input v-model="form.document_number" type="text" name="document_number" class="form-control" required />
      </div>

      <!-- NIC uploads -->
      <div id="nicFields" :style="{ display: form.document_type === 'nic' ? 'block' : 'none' }">
        <div class="form-group mb-3">
          <label>NIC Front Image <template v-if="isEdit && existing.nic_front">(<a :href="url(existing.nic_front)" target="_blank">View</a>)</template></label>
          <input type="file" name="nic_front" class="form-control" @change="onFile($event, 'nic_front')" />
        </div>

        <div class="form-group mb-3">
          <label>NIC Back Image <template v-if="isEdit && existing.nic_back">(<a :href="url(existing.nic_back)" target="_blank">View</a>)</template></label>
          <input type="file" name="nic_back" class="form-control" @change="onFile($event, 'nic_back')" />
        </div>
      </div>

      <!-- Passport upload -->
      <div id="passportField" :style="{ display: form.document_type === 'passport' ? 'block' : 'none' }">
        <div class="form-group mb-3">
          <label>Passport Image <template v-if="isEdit && existing.passport_image">(<a :href="url(existing.passport_image)" target="_blank">View</a>)</template></label>
          <input type="file" name="passport_image" class="form-control" @change="onFile($event, 'passport_image')" />
        </div>
      </div>

      <div class="form-group mt-4">
        <button v-if="isEdit" type="submit" class="btn btn-primary" :disabled="submitting">Update KYC</button>
        <button v-else type="submit" class="btn btn-success" :disabled="submitting">Submit KYC</button>
      </div>
    </form>
  </div>
</template>
