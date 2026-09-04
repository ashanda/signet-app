<script setup>
// Ports countries/index.blade.php ("Countries List"). See
// countries_handler.go's countriesIndexHandler/countriesStoreHandler/
// countriesUpdateHandler/countriesDestroyHandler.
//
// Deviation from ui_spec.md (an approved, documented one, per
// FRONTEND_CONVENTIONS.md): the original uses a native browser confirm()
// for delete; this rebuild uses useAlert().confirmDanger() (SweetAlert2)
// consistently with the rest of the app.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'
import { useAlert } from '@/composables/useToast'

const { run } = useApiAction()
const { confirmDanger } = useAlert()

const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')

const countries = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/countries', { params: { page } })
    countries.value = data.countries
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load countries.'
  } finally {
    loading.value = false
  }
}

// --- Add/Edit Country modal ---
const showModal = ref(false)
const modalMode = ref('add') // 'add' | 'edit'
const form = ref({ id: '', code: '', name: '' })
const errors = ref({})
const saving = ref(false)

function openAddModal() {
  modalMode.value = 'add'
  form.value = { id: '', code: '', name: '' }
  errors.value = {}
  showModal.value = true
}

function editCountry(country) {
  modalMode.value = 'edit'
  form.value = { id: country.id, code: country.code, name: country.name }
  errors.value = {}
  showModal.value = true
}

function closeModal() {
  showModal.value = false
}

async function submitForm() {
  errors.value = {}
  saving.value = true
  const payload = { code: form.value.code, name: form.value.name }
  const call =
    modalMode.value === 'add'
      ? () => api.post('/countries', payload)
      : () => api.put(`/countries/${form.value.id}`, payload)

  const { ok, data, error } = await run(call, {
    successMessage: modalMode.value === 'add' ? 'Country Create Success' : 'Country updated successfully.',
    showErrorAlert: false,
  })
  saving.value = false
  if (ok) {
    showModal.value = false
    flashMessage.value = data?.message || 'Saved.'
    fetchData(countries.value?.current_page || 1)
  } else {
    const respErrors = error?.response?.data?.errors
    if (respErrors) {
      errors.value = respErrors
    } else {
      errors.value = { _general: [error?.response?.data?.message || 'Could not save country.'] }
    }
  }
}

async function deleteCountry(country) {
  const result = await confirmDanger('Are you sure?', 'This country will be deleted permanently.')
  if (!result.isConfirmed) return
  const { ok, data } = await run(() => api.delete(`/countries/${country.id}`), {
    successMessage: 'Country deleted successfully.',
  })
  if (ok) {
    flashMessage.value = data?.message || 'Country deleted successfully.'
    fetchData(countries.value?.current_page || 1)
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="success" :message="flashMessage" @close="flashMessage = ''" />
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4 d-flex align-items-center justify-content-between">
      <h3 class="mb-0">Countries List</h3>
      <button type="button" class="btn btn-primary" @click="openAddModal">Add New Country</button>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table table-bordered table-striped align-middle">
            <thead class="thead-light">
              <tr>
                <th>#</th>
                <th>Country Code</th>
                <th>Country Name</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !countries?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No countries found</td>
              </tr>
              <tr v-for="(row, idx) in countries?.data" :key="row.id">
                <td>{{ (countries.from || 1) + idx }}</td>
                <td>{{ row.code }}</td>
                <td>{{ row.name }}</td>
                <td>
                  <button type="button" class="btn btn-sm btn-outline-primary me-1" @click="editCountry(row)">Edit</button>
                  <button type="button" class="btn btn-sm btn-outline-danger" @click="deleteCountry(row)">Delete</button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="countries" @change="(p) => fetchData(p)" />
      </div>
    </div>

    <!-- Add/Edit Country modal -->
    <div v-if="showModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitForm">
            <div class="modal-header">
              <h5 class="modal-title">{{ modalMode === 'add' ? 'Add Country' : 'Edit Country' }}</h5>
              <button type="button" class="btn-close" @click="closeModal"></button>
            </div>
            <div class="modal-body">
              <div v-if="errors._general" class="alert alert-danger py-2">{{ errors._general[0] }}</div>
              <div class="mb-3">
                <label class="form-label">Country Code</label>
                <input v-model="form.code" type="text" class="form-control" :class="{ 'is-invalid': errors.code }" required />
                <div v-if="errors.code" class="invalid-feedback">{{ errors.code[0] }}</div>
              </div>
              <div class="mb-3">
                <label class="form-label">Country Name</label>
                <input v-model="form.name" type="text" class="form-control" :class="{ 'is-invalid': errors.name }" required />
                <div v-if="errors.name" class="invalid-feedback">{{ errors.name[0] }}</div>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="closeModal">Close</button>
              <button type="submit" class="btn btn-primary" :disabled="saving">Save</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
