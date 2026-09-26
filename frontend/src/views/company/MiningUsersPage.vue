<script setup>
// Ports company/mining.blade.php (route('mining.users')). Purely
// search-driven — MiningController@index passes no data to the view, so
// there's nothing to fetch on mount. See mining_handler.go's
// miningSearchHandler / miningUpdateHandler.
//
// mining/search's `user.id` is already SignetID-formatted (e.g. "SIG-005");
// it's shown as-is and reused in the update URL (the backend's
// parseUintParam maps "SIG-00N" back to N). As in the original, a
// successful update reloads the page.
import { ref } from 'vue'
import Swal from 'sweetalert2'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import { useBsModal } from '@/composables/useBsModal'

const updateModal = useBsModal()

const searchId = ref('')
const result = ref(null)
const form = ref({ id: '', daily_mining: '', total_token: '', status: 'active' })

async function searchSignet() {
  const id = searchId.value
  if (id.trim() === '') return
  try {
    const { data } = await api.get(`/mining/search/${id}`)
    if (data.success) result.value = data
    else Swal.fire('Not Found', 'No record found for this Signet ID', 'error')
  } catch {
    Swal.fire('Not Found', 'No record found for this Signet ID', 'error')
  }
}

function openEdit() {
  form.value = {
    id: result.value.user.id,
    daily_mining: result.value.mining.daily_mining,
    total_token: result.value.mining.total_token,
    status: result.value.mining.status,
  }
  updateModal.show()
}

async function submitUpdate() {
  try {
    const { data } = await api.post(`/mining/update/${form.value.id}`, {
      daily_mining: Number(form.value.daily_mining) || 0,
      total_token: Number(form.value.total_token) || 0,
      status: form.value.status,
    })
    if (data.success) {
      Swal.fire('Updated!', 'Mining data updated successfully', 'success').then(() => {
        window.location.reload()
      })
    } else {
      Swal.fire('Error', data.message ?? 'Update failed', 'error')
    }
  } catch (err) {
    Swal.fire('Error', err?.response?.data?.message ?? 'Update failed', 'error')
  }
}
</script>

<template>
  <DashboardLayout>
    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Mining Token</h1>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <!-- Search Box -->
        <div class="input-group mb-3">
          <input id="search_signet_id" v-model="searchId" type="text" class="form-control" placeholder="Enter Signet ID" @keypress.enter.prevent="searchSignet" />
          <button id="searchBtn" class="btn btn-primary" type="button" @click="searchSignet">Search</button>
        </div>

        <!-- Search Result Table -->
        <div id="resultCard" :class="{ 'd-none': !result }">
          <div v-if="result" class="card shadow-lg border-0 rounded-4 overflow-hidden mb-4 hover-lift" style="transition: transform 0.2s">
            <div class="card-header bg-gradient text-white p-4 sg-card-hero">
              <div class="d-flex align-items-center justify-content-between">
                <div>
                  <h5 class="card-title mb-1">{{ result.user.name }}</h5>
                  <small class="opacity-75">ID: {{ result.user.id }}</small>
                </div>
                <div class="text-end">
                  <span class="badge px-3 py-2 rounded-pill" :class="result.mining.status === 'active' ? 'bg-success' : 'bg-danger'">
                    <i class="bi bi-circle-fill me-1" style="font-size: 0.5rem"></i>
                    {{ result.mining.status.toUpperCase() }}
                  </span>
                </div>
              </div>
            </div>

            <div class="card-body p-4">
              <!-- Email Section -->
              <div class="mb-4 pb-3 border-bottom">
                <div class="d-flex align-items-center">
                  <i class="bi bi-envelope-fill text-primary me-2"></i>
                  <span class="text-muted small">Email:</span>
                </div>
                <p class="mb-0 mt-1 fw-medium">{{ result.user.email }}</p>
              </div>

              <!-- Packages Section -->
              <div class="mb-4 pb-3 border-bottom">
                <h6 class="text-uppercase text-muted small mb-3 fw-bold">
                  <i class="bi bi-box-seam-fill text-warning me-2"></i>Packages
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
                      <div class="fw-bold text-dark">{{ result.sales.total_sales || 'N/A' }}</div>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Mining Data Section -->
              <div class="mb-3">
                <h6 class="text-uppercase text-muted small mb-3 fw-bold">
                  <i class="bi bi-minecart-loaded text-info me-2"></i>Mining Data
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
              <button class="btn btn-warning w-100 py-2 fw-bold rounded-3 editBtn shadow-sm" @click="openEdit">
                <i class="bi bi-pencil-square me-2"></i>Edit Mining Data
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Update Modal -->
    <div id="updateModal" :ref="updateModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <div class="modal-content">
          <form id="updateForm" @submit.prevent="submitUpdate">
            <div class="modal-header">
              <h5 class="modal-title">Update Daily Mining Token</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <input id="update_id" v-model="form.id" type="hidden" name="id" />
              <div class="mb-3">
                <label>Daily Mining Token</label>
                <input id="update_token" v-model="form.daily_mining" type="text" name="daily_mining" class="form-control" />
              </div>
              <div class="mb-3">
                <label>Total Issued Token</label>
                <input id="update_total_token" v-model="form.total_token" type="text" name="total_token" class="form-control" />
              </div>
              <div class="mb-3">
                <label>Status</label>
                <select id="update_status" v-model="form.status" name="status" class="form-control">
                  <option value="active">Active</option>
                  <option value="inactive">Inactive</option>
                </select>
              </div>
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
