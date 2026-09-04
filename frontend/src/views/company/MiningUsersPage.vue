<script setup>
// Ports company/mining.blade.php (route('mining.users')). Purely
// search-driven — MiningController@index (GET /mining/users) passes no
// data to the view in the original, and mining_handler.go's
// miningUsersHandler mirrors that (just {success:true}), so there's
// nothing to fetch on mount. See mining_handler.go's miningSearchHandler /
// miningUpdateHandler for the exact response/request shapes.
//
// NOTE: unlike users/search's response, mining/search's `user.id` is
// already `models.SignetID`-formatted (e.g. "SIG-005"), not a raw numeric
// id — rendered as-is, and re-used as-is in the update URL (the backend's
// parseUintParam normalizes "SIG-00N" back to N via signetIDToNumeric).
import { ref } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAlert } from '@/composables/useToast'

const { alertSuccess, alertError } = useAlert()

const searchId = ref('')
const result = ref(null)
const loadError = ref('')

const showModal = ref(false)
const form = ref({ id: '', daily_mining: 0, total_token: 0, status: 'active' })

async function searchSignet() {
  const id = searchId.value.trim()
  if (!id) return
  loadError.value = ''
  try {
    const { data } = await api.get(`/mining/search/${id}`)
    if (data.success) {
      result.value = data
    } else {
      result.value = null
      await alertError('Not Found', data.message || 'No record found for this Signet ID')
    }
  } catch (err) {
    result.value = null
    await alertError('Not Found', err?.response?.data?.message || 'No record found for this Signet ID')
  }
}

function openEditModal() {
  if (!result.value) return
  form.value = {
    id: result.value.user.id,
    daily_mining: result.value.mining.daily_mining,
    total_token: result.value.mining.total_token,
    status: result.value.mining.status,
  }
  showModal.value = true
}

async function submitUpdate() {
  try {
    const { data } = await api.post(`/mining/update/${form.value.id}`, {
      daily_mining: Number(form.value.daily_mining) || 0,
      total_token: Number(form.value.total_token) || 0,
      status: form.value.status,
    })
    if (data.success) {
      showModal.value = false
      await alertSuccess('Updated!', 'Mining data updated successfully')
      await searchSignet()
    } else {
      await alertError('Error', data.message || 'Update failed')
    }
  } catch (err) {
    await alertError('Error', err?.response?.data?.message || 'Update failed')
  }
}
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Mining Token</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="input-group mb-3">
          <input
            v-model="searchId"
            type="text"
            class="form-control"
            placeholder="Enter Signet ID"
            @keypress.enter.prevent="searchSignet"
          />
          <button class="btn btn-primary" type="button" @click="searchSignet">Search</button>
        </div>

        <div v-if="result" class="card shadow-lg border-0 rounded-4 overflow-hidden mb-0">
          <div class="card-header text-white p-4 mining-gradient-header">
            <div class="d-flex align-items-center justify-content-between">
              <div>
                <h5 class="card-title mb-1 text-dark">{{ result.user.name }}</h5>
                <small class="opacity-75">ID: {{ result.user.id }}</small>
              </div>
              <span
                class="badge px-3 py-2 rounded-pill"
                :class="result.mining.status === 'active' ? 'bg-success' : 'bg-danger'"
              >
                {{ result.mining.status.toUpperCase() }}
              </span>
            </div>
          </div>

          <div class="card-body p-4">
            <div class="mb-4 pb-3 border-bottom">
              <div class="d-flex align-items-center">
                <i class="fas fa-envelope text-primary me-2"></i>
                <span class="text-muted small">Email:</span>
              </div>
              <p class="mb-0 mt-1 fw-medium">{{ result.user.email }}</p>
            </div>

            <div class="mb-4 pb-3 border-bottom">
              <h6 class="text-uppercase text-muted small mb-3 fw-bold">
                <i class="fas fa-box text-warning me-2"></i>Packages
              </h6>
              <div class="row g-3">
                <div class="col-4">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">First Package</div>
                    <div class="fw-bold text-dark">{{ result.packages.first || 'N/A' }}</div>
                  </div>
                </div>
                <div class="col-4">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">Current Package</div>
                    <div class="fw-bold text-dark">{{ result.packages.last || 'N/A' }}</div>
                  </div>
                </div>
                <div class="col-4">
                  <div class="p-3 bg-light rounded-3">
                    <div class="small text-muted mb-1">Sale Count</div>
                    <div class="fw-bold text-dark">{{ result.sales.total_sales ?? 'N/A' }}</div>
                  </div>
                </div>
              </div>
            </div>

            <div class="mb-3">
              <h6 class="text-uppercase text-muted small mb-3 fw-bold">
                <i class="fas fa-coins text-info me-2"></i>Mining Data
              </h6>
              <div class="row g-3">
                <div class="col-md-4">
                  <div class="text-center p-3 bg-primary bg-opacity-10 rounded-3">
                    <div class="small text-muted mb-1">Total Token</div>
                    <div class="h5 mb-0 fw-bold text-primary">{{ result.mining.total_token }}</div>
                  </div>
                </div>
                <div class="col-md-4">
                  <div class="text-center p-3 bg-success bg-opacity-10 rounded-3">
                    <div class="small text-muted mb-1">Mining Token</div>
                    <div class="h5 mb-0 fw-bold text-success">{{ result.mining.mining_token }}</div>
                  </div>
                </div>
                <div class="col-md-4">
                  <div class="text-center p-3 bg-info bg-opacity-10 rounded-3">
                    <div class="small text-muted mb-1">Daily Mining</div>
                    <div class="h5 mb-0 fw-bold text-info">{{ result.mining.daily_mining }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div class="card-footer bg-white border-0 p-4 pt-0">
            <button class="btn btn-warning w-100 py-2 fw-bold rounded-3 shadow-sm" @click="openEditModal">
              <i class="fas fa-pencil-alt me-2"></i>Edit Mining Data
            </button>
          </div>
        </div>
      </div>
    </div>

    <div v-if="showModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitUpdate">
            <div class="modal-header">
              <h5 class="modal-title">Update Daily Mining Token</h5>
              <button type="button" class="btn-close" @click="showModal = false"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">Daily Mining Token</label>
                <input v-model="form.daily_mining" type="text" class="form-control" />
              </div>
              <div class="mb-3">
                <label class="form-label">Total Issued Token</label>
                <input v-model="form.total_token" type="text" class="form-control" />
              </div>
              <div class="mb-3">
                <label class="form-label">Status</label>
                <select v-model="form.status" class="form-select">
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>

<style scoped>
.mining-gradient-header {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
</style>
