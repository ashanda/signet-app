<script setup>
// Ports company/user_parent_logs/index.blade.php (route('userparentlogs.index')).
// See userparentmapslog_handler.go's userParentLogsIndexHandler/
// userParentLogsDestroyHandler.
//
// "Activation Arrived User" is the user with id = parent_id (joined
// server-side as `parent_user`), "User not found" when there is none.
//
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'
import { useAlert } from '@/composables/useToast'
import { phpDateTime, diffForHumans } from '@/utils/format'

const { run } = useApiAction()
const { confirmDanger } = useAlert()

const logs = ref(null)
const loading = ref(true)
const loadError = ref('')
const flashMessage = ref('')

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
    'This will delete all UserParent rows linked to this mapping log!',
  )
  if (!result.isConfirmed) return

  // The original submits a DELETE form and lands back on the list with a
  // session flash — no success dialog.
  const { ok, data } = await run(() => api.delete(`/user-parent-logs/${row.id}`), {
    showSuccessAlert: false,
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
                  <template v-if="row.parent_user">
                    {{ row.parent_user.name || 'N/A' }}<br />
                    <small class="text-muted">SIG ID: {{ 'SIG-00' + row.parent_user.id }}</small>
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
                  {{ phpDateTime(row.created_at) }}
                  <br />
                  <small class="text-muted">{{ diffForHumans(row.created_at) }}</small>
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
