<script setup>
// Ports countries/index.blade.php ("Countries List"). See
// countries_handler.go's countriesIndexHandler/countriesStoreHandler/
// countriesUpdateHandler/countriesDestroyHandler.
//
// Store/update/destroy redirect back with a success flash in the original;
// here the list is re-fetched and the same flash shown. Delete asks with the
// browser's confirm(), as the original does.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import { useBsModal } from '@/composables/useBsModal'

const countryModal = useBsModal()

const flashSuccess = ref('')
const flashError = ref('')
const countries = ref(null)

async function fetchData(page = countries.value?.current_page || 1) {
  try {
    const { data } = await api.get('/countries', { params: { page } })
    countries.value = data.countries
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Could not load countries.'
  }
}

onMounted(() => fetchData(1))

const modalTitle = ref('Add Country')
const form = ref({ id: '', code: '', name: '' })

function createCountry() {
  modalTitle.value = 'Add Country'
  form.value = { id: '', code: '', name: '' }
}

function editCountry(country) {
  modalTitle.value = 'Edit Country'
  form.value = { id: country.id, code: country.code, name: country.name }
}

async function write(request) {
  flashSuccess.value = ''
  flashError.value = ''
  try {
    const { data } = await request()
    if (data?.status === 'error') flashError.value = data.message
    else flashSuccess.value = data?.message || ''
    await fetchData()
    return true
  } catch (err) {
    const errs = err?.response?.data?.errors
    flashError.value = (errs && Object.values(errs)[0]?.[0]) || err?.response?.data?.message || 'Something went wrong'
    return false
  }
}

async function submitForm() {
  const payload = { code: form.value.code, name: form.value.name }
  const ok = await write(() => (form.value.id ? api.put(`/countries/${form.value.id}`, payload) : api.post('/countries', payload)))
  if (ok) countryModal.hide()
}

function deleteCountry(country) {
  if (!window.confirm('Delete this country?')) return
  write(() => api.delete(`/countries/${country.id}`))
}
</script>

<template>
  <DashboardLayout>
    <div class="container">
      <h3 class="mb-4">Countries List</h3>

      <button class="btn btn-primary mb-3" data-bs-toggle="modal" data-bs-target="#countryModal" @click="createCountry">
        Add New Country
      </button>

      <div v-if="flashSuccess" class="alert alert-success">
        {{ flashSuccess }}
      </div>
      <div v-if="flashError" class="alert alert-danger">
        {{ flashError }}
      </div>

      <table class="table table-bordered table-striped">
        <thead>
          <tr>
            <th>#</th>
            <th>Country Code</th>
            <th>Country Name</th>
            <th>Actions</th>
          </tr>
        </thead>

        <tbody>
          <tr v-for="(country, key) in countries?.data" :key="country.id">
            <td>{{ (countries.from || 1) + key }}</td>
            <td>{{ country.code }}</td>
            <td>{{ country.name }}</td>
            <td>
              <button class="btn btn-sm btn-outline-primary" data-bs-toggle="modal" data-bs-target="#countryModal" @click="editCountry(country)">
                Edit
              </button>

              <form class="d-inline" @submit.prevent="deleteCountry(country)">
                <button class="btn btn-sm btn-outline-danger">
                  Delete
                </button>
              </form>
            </td>
          </tr>
        </tbody>
      </table>

      <Paginator :pagination="countries" @change="(p) => fetchData(p)" />
    </div>

    <!-- Country Modal -->
    <div id="countryModal" :ref="countryModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="countryForm" method="POST" @submit.prevent="submitForm">
            <div class="modal-header">
              <h5 id="modalTitle" class="modal-title">
                {{ modalTitle }}
              </h5>

              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>

            <div class="modal-body">
              <div class="mb-3">
                <label>Country Code</label>
                <input id="code" v-model="form.code" type="text" name="code" class="form-control" required />
              </div>

              <div class="mb-3">
                <label>Country Name</label>
                <input id="name" v-model="form.name" type="text" name="name" class="form-control" required />
              </div>
            </div>

            <div class="modal-footer">
              <button type="submit" class="btn btn-primary">
                Save
              </button>

              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">
                Close
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
