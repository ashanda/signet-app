<script setup>
// Ports the "Mining Community Staking Token" card duplicated across
// Admin/Agent/User dashboards (ui_spec.md "Mining 'Community Staking Token'
// widget"): polls GET /mining/user/{userId} on mount and every 5s, ticks a
// local per-second simulated counter between polls, fires a SweetAlert2
// "Mining Complete!" toast once mining_token >= total_token, and shows a
// live Connecting/Connected/Disconnected badge.
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import api from '@/api/client'
import { useToast } from '@/composables/useToast'

const props = defineProps({
  userId: { type: [Number, String], required: true },
  allUsersCount: { type: Number, default: 0 },
})

const { toastSuccess } = useToast()

const connection = ref('connecting') // 'connecting' | 'connected' | 'disconnected'
const miningToken = ref(0)
const totalToken = ref(0)
const dailyMining = ref(0)
const status = ref('inactive')
const lastUpdated = ref(null)
const completed = ref(false)

let localTicker = null
let syncTicker = null

const progressPercent = computed(() => {
  if (!totalToken.value) return 0
  return Math.min(100, Math.round((miningToken.value / totalToken.value) * 10000) / 100)
})

const ratePerSecond = computed(() => {
  // dailyMining is tokens/day in the source data; derive tokens/second for
  // the local ticking simulation and the "N tokens/second" badge.
  return dailyMining.value > 0 ? dailyMining.value / 86400 : 0
})

async function poll() {
  try {
    const { data } = await api.get(`/mining/user/${props.userId}`)
    if (data.success) {
      connection.value = 'connected'
      miningToken.value = data.data.mining_token
      totalToken.value = data.data.total_token
      dailyMining.value = data.data.daily_mining
      status.value = data.data.status
      lastUpdated.value = data.data.updated_at
      checkComplete()
    } else {
      connection.value = 'disconnected'
    }
  } catch {
    connection.value = 'disconnected'
  }
}

function checkComplete() {
  if (!completed.value && totalToken.value > 0 && miningToken.value >= totalToken.value) {
    completed.value = true
    toastSuccess('Mining Complete!')
  }
}

function tickLocal() {
  if (status.value !== 'active' || completed.value) return
  miningToken.value += ratePerSecond.value
  checkComplete()
}

onMounted(() => {
  poll()
  localTicker = setInterval(tickLocal, 1000)
  syncTicker = setInterval(poll, 5000)
})
onBeforeUnmount(() => {
  clearInterval(localTicker)
  clearInterval(syncTicker)
})
</script>

<template>
  <div class="card shadow-lg mb-4">
    <div class="card-header bg-primary text-white d-flex justify-content-between align-items-center">
      <h5 class="mb-0">⛏️ Mining Community Staking Token - All Users Count({{ allUsersCount }})</h5>
      <span
        class="badge"
        :class="{
          'bg-secondary': connection === 'connecting',
          'bg-success': connection === 'connected',
          'bg-danger': connection === 'disconnected',
        }"
      >
        {{ connection === 'connecting' ? 'Connecting…' : connection === 'connected' ? 'Connected' : 'Disconnected' }}
      </span>
    </div>
    <div class="card-body">
      <div class="row text-center mb-3">
        <div class="col-3">
          <div class="text-muted small">Mining Token</div>
          <div class="fw-bold">{{ miningToken.toFixed(8) }}</div>
        </div>
        <div class="col-3">
          <div class="text-muted small">Total Token</div>
          <div class="fw-bold">{{ totalToken }}</div>
        </div>
        <div class="col-3">
          <div class="text-muted small">Daily Mining</div>
          <div class="fw-bold">{{ dailyMining }}</div>
        </div>
        <div class="col-3">
          <div class="text-muted small">Status</div>
          <span class="badge" :class="status === 'active' ? 'bg-success' : 'bg-danger'">{{ status }}</span>
        </div>
      </div>
      <div class="progress mb-2" style="height: 20px;">
        <div
          class="progress-bar progress-bar-striped progress-bar-animated"
          :style="{ width: progressPercent + '%' }"
        >
          {{ progressPercent }}%
        </div>
      </div>
      <div class="d-flex justify-content-between">
        <span class="badge bg-info text-dark">{{ ratePerSecond.toFixed(8) }} tokens/second</span>
        <small class="text-muted">Last updated: {{ lastUpdated || '—' }}</small>
      </div>
    </div>
  </div>
</template>
