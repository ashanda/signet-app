<script setup>
// Ports company/direct_share.blade.php ("Globle Direct Share" — literal
// on-page typo, reproduced verbatim). See directshare_handler.go's
// directShareHandler/packagePoolStoreHandler/packagePoolUpdateHandler/
// packagePoolDestroyHandler.
//
// Edit/Delete apply only to company rows (user_id 1); every other row shows
// the "Auto" badge. "Package Value" is the pool's package price ("N/A" if
// the package is gone), "-" for company rows. Store/update/destroy redirect
// back with a session flash in the original; here the list is re-fetched
// and the same alert shown. The summary's coin icon is a Phosphor class the
// original never loads, so it renders as an empty circle there too.
import { ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Swal from 'sweetalert2'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import { useBsModal } from '@/composables/useBsModal'
import { numberFormat, phpDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const editModal = useBsModal()

const flashSuccess = ref('')
const flashError = ref('')
const startInput = ref('')
const endInput = ref('')
const companyPool = ref(0)
const salesPool = ref(0)
const totalPool = ref(0)
const pools = ref(null)

async function fetchData() {
  const q = route.query
  startInput.value = q.start_date || ''
  endInput.value = q.end_date || ''
  try {
    const params = {}
    for (const key of ['page', 'start_date', 'end_date']) if (q[key]) params[key] = q[key]
    const { data } = await api.get('/direct-share', { params })
    companyPool.value = data.company_pool
    salesPool.value = data.sales_pool
    totalPool.value = data.total_pool
    pools.value = data.pools
  } catch (err) {
    flashError.value = err?.response?.data?.message || 'Could not load direct share data.'
  }
}

watch(() => route.query, fetchData, { immediate: true })

function applyFilter() {
  router.push({ path: route.path, query: { start_date: startInput.value, end_date: endInput.value } })
}

// ->paginate(20)->withQueryString()
function goToPage(page) {
  router.push({ path: route.path, query: { ...route.query, page } })
}

async function write(request, fallback) {
  flashSuccess.value = ''
  flashError.value = ''
  try {
    const { data } = await request()
    if (data?.status === 'error') flashError.value = data.message
    else flashSuccess.value = data?.message || fallback
    await fetchData()
    return true
  } catch (err) {
    const errs = err?.response?.data?.errors
    flashError.value = (errs && Object.values(errs)[0]?.[0]) || err?.response?.data?.message || 'Something went wrong'
    return false
  }
}

const insertAmount = ref('')
async function insertPool() {
  const ok = await write(() => api.post('/package-pools', { user_id: 1, pool_amount: Number(insertAmount.value) }), 'Pool added successfully.')
  if (ok) insertAmount.value = ''
}

const editForm = ref({ id: '', pool_amount: '' })
function openEdit(pool) {
  editForm.value = { id: pool.id, pool_amount: pool.pool_amount }
  editModal.show()
}
async function submitEdit() {
  const ok = await write(
    () => api.put(`/package-pools/${editForm.value.id}`, { pool_amount: Number(editForm.value.pool_amount) }),
    'Pool Values updated successfully.',
  )
  if (ok) editModal.hide()
}

function deletePool(id) {
  Swal.fire({
    title: 'Are you sure?',
    text: 'This pool will be deleted permanently.',
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#d33',
    cancelButtonColor: '#6c757d',
    confirmButtonText: 'Yes, Delete',
  }).then((result) => {
    if (result.isConfirmed) write(() => api.delete(`/package-pools/${id}`), 'Pool Values deleted successfully.')
  })
}
</script>

<template>
  <DashboardLayout>
    <div v-if="flashSuccess" class="alert alert-success">{{ flashSuccess }}</div>
    <div v-if="flashError" class="alert alert-danger">{{ flashError }}</div>

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Globle Direct Share</h1>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <!-- Search Form -->
        <form method="GET" class="mb-3" @submit.prevent="applyFilter">
          <div class="row g-2">
            <div class="col-md-4">
              <input v-model="startInput" type="date" name="start_date" class="form-control" />
            </div>
            <div class="col-md-4">
              <input v-model="endInput" type="date" name="end_date" class="form-control" />
            </div>
            <div class="col-md-4 d-flex">
              <button type="submit" class="btn btn-primary me-2">
                Filter
              </button>
              <RouterLink v-if="route.query.start_date || route.query.end_date" to="/direct-share" class="btn btn-outline-secondary">
                Clear
              </RouterLink>
            </div>
          </div>
        </form>

        <div class="row g-3 mb-4">
          <div class="col-12 col-md-12">
            <div class="p-3 rounded-4 border bg-light h-100">
              <div class="d-flex align-items-center justify-content-between">
                <div>
                  <div class="text-muted small fw-semibold">Per Month Total Pool Value</div>
                  <div class="fs-4 fw-bold mt-1">USDT {{ numberFormat(totalPool, 2) }}</div>
                </div>
                <div>
                  <div class="text-muted small fw-semibold">Sales Through Get Pool Value</div>
                  <div class="fs-4 fw-bold mt-1">USDT {{ numberFormat(salesPool, 2) }}</div>
                </div>
                <div>
                  <div class="text-muted small fw-semibold">Company Included the Pool Value</div>
                  <div class="fs-4 fw-bold mt-1">USDT {{ numberFormat(companyPool, 2) }}</div>
                </div>
                <div class="rounded-circle bg-white border d-flex align-items-center justify-content-center" style="width: 46px; height: 46px">
                  <i class="ph-bold ph-coins fs-4 text-primary"></i>
                </div>
              </div>
              <div class="mt-2 small text-muted">Total for selected Date Range</div>
            </div>
          </div>
        </div>

        <div class="row align-items-end">
          <form method="POST" class="row g-2" @submit.prevent="insertPool">
            <input type="hidden" name="user_id" value="1" />
            <label class="form-label">Insert Pool Amount</label>
            <div class="col-md-10">
              <input v-model="insertAmount" type="number" name="pool_amount" step="0.01" class="form-control" placeholder="Enter Pool Amount" required />
            </div>
            <div class="col-md-2">
              <button type="submit" class="btn btn-primary w-100">
                Save
              </button>
            </div>
          </form>
        </div>

        <div class="table-responsive mt-4">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">User name</th>
                <th class="border-bottom" scope="col">SIG ID</th>
                <th class="border-bottom" scope="col">Package Value</th>
                <th class="border-bottom" scope="col">Pool Amount</th>
                <th class="border-bottom" scope="col">Date</th>
                <th class="border-bottom" scope="col">Actions</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="pool in pools?.data" :key="pool.id">
                <td>{{ pool.user_name }}</td>
                <td>{{ 'SIG-00' + pool.user_id }}</td>
                <td>{{ pool.user_id == 1 ? '-' : (pool.package_price ?? 'N/A') }}</td>
                <td>{{ Number(pool.pool_amount || 0).toFixed(2) }}</td>
                <td>{{ phpDateTime(pool.created_at, 10) }}</td>
                <td>
                  <template v-if="pool.user_id == 1">
                    <button type="button" class="btn btn-sm btn-warning" @click="openEdit(pool)">
                      <i class="fas fa-edit"></i>
                    </button>
                    <form class="d-inline" @submit.prevent>
                      <button type="button" class="btn btn-sm btn-danger" @click="deletePool(pool.id)">
                        <i class="fas fa-trash"></i>
                      </button>
                    </form>
                  </template>
                  <span v-else class="badge bg-secondary">Auto</span>
                </td>
              </tr>
              <tr v-if="pools && !pools.data?.length">
                <td colspan="4" class="text-center text-muted">No pools data found</td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="pools" @change="goToPage" />
          </div>
        </div>
      </div>
    </div>

    <div :ref="editModal.el" class="modal fade" tabindex="-1">
      <div class="modal-dialog">
        <form method="POST" @submit.prevent="submitEdit">
          <div class="modal-content">
            <div class="modal-header">
              <h5 class="modal-title">Edit Pool Amount</h5>
              <button type="button" class="btn-close" data-bs-dismiss="modal"></button>
            </div>
            <div class="modal-body">
              <div class="mb-3">
                <label class="form-label">Pool Amount</label>
                <input v-model="editForm.pool_amount" type="number" name="pool_amount" step="0.01" class="form-control" required />
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn btn-primary">
                Update
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  </DashboardLayout>
</template>
