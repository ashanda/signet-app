<script setup>
// Ports company/user_parent_logs/index.blade.php (route('userparentlogs.index')).
// See userparentmapslog_handler.go's userParentLogsIndexHandler/
// userParentLogsDestroyHandler.
//
// Deviation from the original (flagged for the backend team): the original
// separately looks up `App\Models\User::where('id', $log->parent_id)`.first()
// for the "Activation Arrived User" column (name + SIG id). The Go
// handler's response only carries the raw `parent_id` (no joined name/
// status) — only `user_id`'s row is joined and flattened into a `user`
// object. So the "Activation Arrived User" column here can only render
// "SIG-00{parent_id}" without a name, unlike "New User" which has the full
// joined record.
//
// Also note: `parent_id`/`user_id`/`created_at` are sql.Null* fields
// scanned straight into the JSON response (no `.Int64`/`.Time` flatten in
// the handler), so they arrive as {Int64,Valid}/{Time,Valid} objects, not
// plain values — nullInt()/nullTime() below unwrap that. The nested `user`
// object IS properly flattened (built via a manual map in the handler).
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'
import { useAlert } from '@/composables/useToast'

const { run } = useApiAction()
const { confirmDanger } = useAlert()

const logs = ref(null)
const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')

function nullInt(v) {
  return v && typeof v === 'object' && v.Valid ? v.Int64 : null
}

function nullTime(v) {
  return v && typeof v === 'object' && v.Valid ? v.Time : null
}

function fmtDate(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleString()
}

function timeAgo(iso) {
  if (!iso) return ''
  const then = new Date(iso).getTime()
  if (Number.isNaN(then)) return ''
  const seconds = Math.floor((Date.now() - then) / 1000)
  const units = [
    ['year', 31536000],
    ['month', 2592000],
    ['day', 86400],
    ['hour', 3600],
    ['minute', 60],
  ]
  for (const [name, secs] of units) {
    const n = Math.floor(Math.abs(seconds) / secs)
    if (n >= 1) return seconds >= 0 ? `${n} ${name}${n > 1 ? 's' : ''} ago` : `in ${n} ${name}${n > 1 ? 's' : ''}`
  }
  return 'just now'
}

async function fetchLogs(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/user-parent-logs', { params: { page } })
    logs.value = data.logs
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load logs.'
  } finally {
    loading.value = false
  }
}

async function deleteLog(row) {
  const result = await confirmDanger(
    'Are you sure?',
    'This will delete all UserParent rows linked to this mapping log!'
  )
  if (!result.isConfirmed) return

  const { ok, data } = await run(() => api.delete(`/user-parent-logs/${row.id}`), {
    successMessage: 'Related UserParent records, referral code, user and log were deleted successfully.',
  })
  if (ok) {
    flashMessage.value = data?.message || 'Deleted successfully.'
    fetchLogs(logs.value?.current_page || 1)
  }
}

onMounted(() => fetchLogs())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="success" :message="flashMessage" @close="flashMessage = ''" />
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Fake Account (Deactive Users, 10+ hours old)</h1>
    </div>

    <div class="card border-0 shadow-sm">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th>#</th>
                <th>Activation Arrived User</th>
                <th>New User</th>
                <th>Status</th>
                <th>Created At</th>
                <th class="text-end">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !logs?.data?.length">
                <td colspan="6" class="text-center text-muted py-4">
                  No logs found for deactive users older than 10 hours.
                </td>
              </tr>
              <tr v-for="row in logs?.data" :key="row.id">
                <td>{{ row.id }}</td>
                <td>
                  <template v-if="nullInt(row.parent_id) !== null">
                    <span class="fst-italic text-muted">Name unavailable</span><br />
                    <small class="text-muted">SIG ID: SIG-00{{ nullInt(row.parent_id) }}</small>
                  </template>
                  <span v-else class="text-muted">User not found</span>
                </td>
                <td>
                  <template v-if="row.user">
                    {{ row.user.name || 'N/A' }}<br />
                    <small class="text-muted">SIG ID: {{ row.user.signet_id }}</small>
                  </template>
                  <span v-else class="text-muted">User not found</span>
                </td>
                <td>
                  <span class="badge bg-secondary">{{ row.user?.status || 'unknown' }}</span>
                </td>
                <td>
                  {{ fmtDate(nullTime(row.created_at)) }}
                  <br />
                  <small class="text-muted">{{ timeAgo(nullTime(row.created_at)) }}</small>
                </td>
                <td class="text-end">
                  <button type="button" class="btn btn-sm btn-danger" @click="deleteLog(row)">
                    <i class="fas fa-trash-alt me-1"></i>Delete mappings
                  </button>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="logs" @change="(p) => fetchLogs(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
