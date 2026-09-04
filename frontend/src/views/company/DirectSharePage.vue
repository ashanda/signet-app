<script setup>
// Ports company/direct_share.blade.php ("Globle Direct Share" — literal
// on-page typo, reproduced verbatim per ui_spec.md). See
// directshare_handler.go's directShareHandler/packagePoolStoreHandler/
// packagePoolUpdateHandler/packagePoolDestroyHandler.
//
// Deviations from ui_spec.md's documented markup, verified against the
// actual Go handler (trusted over prose):
//  - Edit/Delete actions apply only to rows where `pool.user_id == 1`
//    (the handler signals this by simply returning each row's own
//    `user_id` field — no separate "editable" flag — so we compare
//    `row.user_id === 1` client-side). All other rows show the "Auto"
//    badge, matching ui_spec.md.
//  - "Package Value" column shows the joined package name (the handler
//    only returns `package_name`, not a numeric package price) — "-" for
//    user_id 1 rows per ui_spec.md.
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'
import { useAlert } from '@/composables/useToast'

const route = useRoute()
const router = useRouter()
const { run } = useApiAction()
const { confirmDanger } = useAlert()

const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')

const startDate = ref(route.query.start_date || '')
const endDate = ref(route.query.end_date || '')

const companyPool = ref(0)
const salesPool = ref(0)
const totalPool = ref(0)
const pools = ref(null)

function fmt(n) {
  return Number(n || 0).toFixed(2)
}

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const params = { page }
    if (startDate.value) params.start_date = startDate.value
    if (endDate.value) params.end_date = endDate.value
    const { data } = await api.get('/direct-share', { params })
    companyPool.value = data.company_pool
    salesPool.value = data.sales_pool
    totalPool.value = data.total_pool
    pools.value = data.pools
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load direct share data.'
  } finally {
    loading.value = false
  }
}

function applyFilter() {
  router.replace({ query: { ...(startDate.value ? { start_date: startDate.value } : {}), ...(endDate.value ? { end_date: endDate.value } : {}) } })
  fetchData(1)
}

function clearFilter() {
  startDate.value = ''
  endDate.value = ''
  router.replace({ query: {} })
  fetchData(1)
}

// --- Insert Pool Amount form ---
const insertAmount = ref('')
const inserting = ref(false)
async function insertPool() {
  if (insertAmount.value === '' || insertAmount.value === null) return
  inserting.value = true
  const { ok, data } = await run(
    () => api.post('/package-pools', { user_id: 1, pool_amount: Number(insertAmount.value) }),
    { successMessage: 'Pool added successfully.' }
  )
  inserting.value = false
  if (ok) {
    flashMessage.value = data?.message || 'Pool added successfully.'
    insertAmount.value = ''
    fetchData(pools.value?.current_page || 1)
  }
}

// --- Edit Pool modal ---
const showEditModal = ref(false)
const editForm = ref({ id: '', pool_amount: '' })
function openEditModal(row) {
  editForm.value = { id: row.id, pool_amount: row.pool_amount }
  showEditModal.value = true
}
async function submitEdit() {
  const { ok, data } = await run(
    () => api.put(`/package-pools/${editForm.value.id}`, { pool_amount: Number(editForm.value.pool_amount) }),
    { successMessage: 'Pool Values updated successfully.' }
  )
  if (ok) {
    showEditModal.value = false
    flashMessage.value = data?.message || 'Pool Values updated successfully.'
    fetchData(pools.value?.current_page || 1)
  }
}

// --- Delete pool ---
async function deletePool(row) {
  const result = await confirmDanger('Are you sure?', 'This pool will be deleted permanently.')
  if (!result.isConfirmed) return
  const { ok, data } = await run(() => api.delete(`/package-pools/${row.id}`), {
    successMessage: 'Pool Values deleted successfully.',
  })
  if (ok) {
    flashMessage.value = data?.message || 'Pool Values deleted successfully.'
    fetchData(pools.value?.current_page || 1)
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="success" :message="flashMessage" @close="flashMessage = ''" />
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Globle Direct Share</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <form class="row g-2 align-items-end mb-4" @submit.prevent="applyFilter">
          <div class="col-auto">
            <label class="form-label small mb-1">From Date</label>
            <input v-model="startDate" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <label class="form-label small mb-1">To Date</label>
            <input v-model="endDate" type="date" class="form-control" />
          </div>
          <div class="col-auto">
            <button type="submit" class="btn btn-primary">Filter</button>
          </div>
          <div v-if="startDate || endDate" class="col-auto">
            <a href="#" class="btn btn-outline-dark" @click.prevent="clearFilter">Clear</a>
          </div>
        </form>

        <div class="p-3 rounded-4 border bg-light mb-4">
          <div class="d-flex align-items-center justify-content-between flex-wrap gap-3">
            <div>
              <div class="text-muted small fw-semibold">Per Month Total Pool Value</div>
              <div class="fs-4 fw-bold mt-1">USDT {{ fmt(totalPool) }}</div>
            </div>
            <div>
              <div class="text-muted small fw-semibold">Sales Through Get Pool Value</div>
              <div class="fs-4 fw-bold mt-1">USDT {{ fmt(salesPool) }}</div>
            </div>
            <div>
              <div class="text-muted small fw-semibold">Company Included the Pool Value</div>
              <div class="fs-4 fw-bold mt-1">USDT {{ fmt(companyPool) }}</div>
            </div>
            <div class="rounded-circle bg-white border d-flex align-items-center justify-content-center" style="width:46px;height:46px;">
              <i class="fas fa-coins fs-5 text-primary"></i>
            </div>
          </div>
          <div class="mt-2 small text-muted">Total for selected Date Range</div>
        </div>

        <form class="row g-2 align-items-end mb-4" @submit.prevent="insertPool">
          <div class="col-auto">
            <label class="form-label small mb-1">Insert Pool Amount</label>
            <input v-model="insertAmount" type="number" step="0.01" class="form-control" required />
          </div>
          <div class="col-auto">
            <button type="submit" class="btn btn-success" :disabled="inserting">Save</button>
          </div>
        </form>

        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">User name</th>
                <th class="border-bottom">SIG ID</th>
                <th class="border-bottom">Package Value</th>
                <th class="border-bottom">Pool Amount</th>
                <th class="border-bottom">Date</th>
                <th class="border-bottom">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !pools?.data?.length">
                <td colspan="6" class="text-center text-muted py-4">No pools data found</td>
              </tr>
              <tr v-for="row in pools?.data" :key="row.id">
                <td>{{ row.user_name || '—' }}</td>
                <td>SIG-00{{ row.user_id }}</td>
                <td>{{ Number(row.user_id) === 1 ? '-' : row.package_name || '-' }}</td>
                <td>{{ fmt(row.pool_amount) }}</td>
                <td>{{ row.created_at ? String(row.created_at).slice(0, 10) : '' }}</td>
                <td>
                  <template v-if="Number(row.user_id) === 1">
                    <button type="button" class="btn btn-sm btn-warning me-1" @click="openEditModal(row)">
                      <i class="fas fa-pencil-alt"></i>
                    </button>
                    <button type="button" class="btn btn-sm btn-danger" @click="deletePool(row)">
                      <i class="fas fa-trash"></i>
                    </button>
                  </template>
                  <span v-else class="badge bg-secondary">Auto</span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="pools" @change="(p) => fetchData(p)" />
      </div>
    </div>

    <!-- Edit Pool Amount modal -->
    <div v-if="showEditModal" class="modal fade show d-block" tabindex="-1" style="background: rgba(0,0,0,0.5);">
      <div class="modal-dialog">
        <div class="modal-content">
          <form @submit.prevent="submitEdit">
            <div class="modal-header">
              <h5 class="modal-title">Edit Pool Amount</h5>
              <button type="button" class="btn-close" @click="showEditModal = false"></button>
            </div>
            <div class="modal-body">
              <label class="form-label">Pool Amount</label>
              <input v-model="editForm.pool_amount" type="number" step="0.01" class="form-control" required />
            </div>
            <div class="modal-footer">
              <button type="button" class="btn btn-secondary" @click="showEditModal = false">Close</button>
              <button type="submit" class="btn btn-primary">Update</button>
            </div>
          </form>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
