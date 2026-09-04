<script setup>
// Ports auth/register_step4.blade.php — on-screen labelled "Step 4: Account
// Status" (the file itself is headed register_step3.blade.php in the
// original source, a copy-paste leftover — see ui_spec.md's file/label
// mismatch note; this view is wired to route.step3.status / /register/status/:id).
// See registerStatusHandler in auth_handler.go.
import { computed, onMounted, ref } from 'vue'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const props = defineProps({
  id: { type: [String, Number], required: true },
})

const loading = ref(true)
const loadError = ref('')
const status = ref('')

async function fetchStatus() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/register/status', { params: { id: props.id } })
    if (data.status === 'success') {
      status.value = data.user?.status || ''
    } else {
      loadError.value = data.message || 'Could not load account status.'
    }
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load account status.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchStatus)

const statusMessage = computed(() => {
  if (status.value === 'pending') return 'Your account is currently pending. Please wait for activation.'
  if (status.value === 'active') return 'Your account is active. You can now login!'
  return 'Your account is inactive. Please contact support.'
})
</script>

<template>
  <AuthCardLayout max-width="500px" back-text="Back to log in" back-to="/login">
    <h1 class="h3 mb-4">Step 4: Account Status</h1>

    <div v-if="loading" class="text-center text-muted py-4">Loading&hellip;</div>

    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <template v-if="!loading && !loadError">
      <p>{{ statusMessage }}</p>
      <p>We'll notify you when your account has been activated.</p>

      <div class="alert alert-info">
        Your current status: <strong>{{ status }}</strong>
      </div>

      <button type="button" class="btn btn-outline-secondary btn-sm mb-3" @click="fetchStatus">
        <i class="fas fa-sync-alt me-1"></i>Refresh status
      </button>
    </template>

    <RouterLink to="/login" class="btn btn-primary d-block">Go to Login</RouterLink>
  </AuthCardLayout>
</template>
