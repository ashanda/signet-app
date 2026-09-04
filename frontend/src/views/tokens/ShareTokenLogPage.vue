<script setup>
// Ports tokens/share-log.blade.php ("Tokens Share Log" — route
// ('token.share.log'), admin/user/agent). See token_handler.go's
// tokenShareLogHandler (GET /token/share/logs): rows shaped
// {id, amount, created_at, receiver, receiver_whatsapp} — "Unknown User"/
// "N/A" fallbacks are already applied server-side.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const loading = ref(true)
const loadError = ref('')
const logs = ref(null)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/token/share/logs', { params: { page } })
    logs.value = data.logs
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load token share log.'
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">Tokens Share Log</h1>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">Tokens</th>
                <th class="border-bottom">Receiver</th>
                <th class="border-bottom">Receiver Whats App</th>
                <th class="border-bottom">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !logs?.data?.length">
                <td colspan="4" class="text-center text-muted py-4">No logs found</td>
              </tr>
              <tr v-for="row in logs?.data" :key="row.id">
                <td><span class="badge bg-success">{{ row.amount }} USDT</span></td>
                <td><span class="badge bg-info">{{ row.receiver || 'Unknown User' }}</span></td>
                <td><span class="badge bg-info">{{ row.receiver_whatsapp || 'N/A' }}</span></td>
                <td><span class="badge bg-info">{{ row.created_at }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="logs" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
