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
import { phpDateTime } from '@/utils/format'

const loadError = ref('')
const logs = ref(null)

async function fetchData(page = 1) {
  loadError.value = ''
  try {
    const { data } = await api.get('/token/share/logs', { params: { page } })
    logs.value = data.logs
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load token share log.'
  }
}

onMounted(() => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">Tokens Share Log</h1>
        </div>
      </div>
    </div>
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">Tokens</th>
                <th class="border-bottom" scope="col">Receiver</th>
                <th class="border-bottom" scope="col">Receiver Whats App</th>
                <th class="border-bottom" scope="col">Date</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="token in logs?.data" :key="token.id">
                <td><span class="badge bg-success">{{ token.amount ?? 0 }} USDT</span></td>
                <td><span class="badge bg-info">{{ token.receiver || 'Unknown User' }}</span></td>
                <td><span class="badge bg-info">{{ token.receiver_whatsapp || 'N/A' }}</span></td>
                <td><span class="badge bg-info">{{ phpDateTime(token.created_at) }}</span></td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="logs" @change="(p) => fetchData(p)" />
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
