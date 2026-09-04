<script setup>
// Ports packages/create.blade.php + packages/edit.blade.php, both sharing
// packages/form.blade.php's field set (ui_spec.md's shared-partial note) —
// one Vue component for both routes, `id` prop undefined on create. See
// package_handler.go's packageStoreHandler/packageEditHandler/
// packageUpdateHandler.
//
// Deviations:
//  - models.Package's Rank/TelegramLink are sql.NullString and are
//    returned straight through (not flattened server-side), so a prefill
//    from GET /packages/{id}/edit arrives as {String,Valid} objects —
//    nullStr() unwraps that when populating the form.
//  - ui_spec.md documents Telegram Link as optional, but
//    validatePackagePayload in the Go handler actually rejects a blank/
//    missing telegram_link ("The telegram link field is required.") —
//    trusting the handler over the prose, the field is marked required
//    here too so the form doesn't submit something the backend will 422.
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'

const props = defineProps({
  id: { type: [String, Number], default: undefined },
})

const router = useRouter()
const { run } = useApiAction()

const isEdit = computed(() => props.id !== undefined && props.id !== null && props.id !== '')

function nullStr(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}

const loading = ref(isEdit.value)
const loadError = ref('')
const saving = ref(false)
const errors = ref({})

const form = ref({
  name: '',
  price: '',
  commission: '',
  rank: '',
  telegram_link: '',
  status: 'active',
})

async function loadPackage() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get(`/packages/${props.id}/edit`)
    const pkg = data.package || {}
    form.value = {
      name: pkg.name || '',
      price: pkg.price ?? '',
      commission: pkg.commission ?? '',
      rank: nullStr(pkg.rank),
      telegram_link: nullStr(pkg.telegram_link),
      status: pkg.status || 'active',
    }
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load package.'
  } finally {
    loading.value = false
  }
}

async function submitForm() {
  errors.value = {}
  saving.value = true
  const payload = {
    name: form.value.name,
    price: form.value.price === '' ? null : Number(form.value.price),
    commission: form.value.commission === '' ? null : Number(form.value.commission),
    rank: form.value.rank,
    telegram_link: form.value.telegram_link,
    status: form.value.status,
  }
  const call = isEdit.value
    ? () => api.put(`/packages/${props.id}`, payload)
    : () => api.post('/packages', payload)

  const { ok, error } = await run(call, {
    successMessage: isEdit.value ? 'Package updated successfully.' : 'Package created successfully.',
    showErrorAlert: false,
  })
  saving.value = false
  if (ok) {
    router.push({ name: 'packages.index' })
  } else {
    const respErrors = error?.response?.data?.errors
    if (respErrors) {
      errors.value = respErrors
    } else {
      errors.value = { _general: [error?.response?.data?.message || 'Could not save package.'] }
    }
  }
}

onMounted(() => {
  if (isEdit.value) loadPackage()
})
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="container">
      <h3>{{ isEdit ? 'Edit Package' : 'Create New Package' }}</h3>

      <div v-if="loading" class="text-muted py-4">Loading…</div>
      <form v-else @submit.prevent="submitForm">
        <div v-if="errors._general" class="alert alert-danger py-2">{{ errors._general[0] }}</div>

          <div class="mb-3">
            <label class="form-label">Name</label>
            <input v-model="form.name" type="text" class="form-control" :class="{ 'is-invalid': errors.name }" required />
            <div v-if="errors.name" class="invalid-feedback">{{ errors.name[0] }}</div>
          </div>

          <div class="mb-3">
            <label class="form-label">Price</label>
            <input v-model="form.price" type="number" step="0.01" class="form-control" :class="{ 'is-invalid': errors.price }" required />
            <div v-if="errors.price" class="invalid-feedback">{{ errors.price[0] }}</div>
          </div>

          <div class="mb-3">
            <label class="form-label">Commission</label>
            <input v-model="form.commission" type="number" step="0.01" class="form-control" :class="{ 'is-invalid': errors.commission }" required />
            <div v-if="errors.commission" class="invalid-feedback">{{ errors.commission[0] }}</div>
          </div>

          <div class="mb-3">
            <label class="form-label">Rank</label>
            <input v-model="form.rank" type="text" class="form-control" :class="{ 'is-invalid': errors.rank }" required />
            <div v-if="errors.rank" class="invalid-feedback">{{ errors.rank[0] }}</div>
          </div>

          <div class="mb-3">
            <label class="form-label">Telegram Link</label>
            <input
              v-model="form.telegram_link"
              type="text"
              class="form-control"
              :class="{ 'is-invalid': errors.telegram_link }"
              placeholder="https://t.me/yourchannel"
              required
            />
            <div v-if="errors.telegram_link" class="invalid-feedback">{{ errors.telegram_link[0] }}</div>
          </div>

          <div class="mb-3">
            <label class="form-label">Status</label>
            <select v-model="form.status" class="form-select" :class="{ 'is-invalid': errors.status }" required>
              <option value="active">Active</option>
              <option value="deactive">Deactive</option>
            </select>
            <div v-if="errors.status" class="invalid-feedback">{{ errors.status[0] }}</div>
          </div>

        <button type="submit" class="btn btn-primary" :disabled="saving">
          {{ isEdit ? 'Update' : 'Save' }}
        </button>
      </form>
    </div>
  </DashboardLayout>
</template>
