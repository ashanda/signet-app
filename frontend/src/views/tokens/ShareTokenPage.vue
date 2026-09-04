<script setup>
// Ports tokens/share.blade.php ("Share Token" — route('token.share'), admin/
// user/agent). See token_handler.go's tokenShareBalanceHandler (GET
// /token-shares) and tokenShareSendHandler (POST /token/share).
//
// Note: ui_spec.md flags a malformed/unclosed markup fragment in the
// original's "My Token : {{ $tokens }} USDT" line — rebuilt cleanly here,
// no bug to reproduce (ui_spec.md explicitly says so).
//
// tokenShareSendHandler's JSON body uses camelCase `tokenValue` (not
// `token_value`) alongside snake_case `user_id` — verified against the Go
// struct tag, not guessed.
import { reactive, ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import { useApiAction } from '@/composables/useApiAction'

const { run } = useApiAction()

const loading = ref(true)
const loadError = ref('')
const tokens = ref(0)

const form = reactive({ tokenValue: '', userId: '' })
const errors = ref({})
const submitting = ref(false)

async function fetchBalance() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/token-shares')
    tokens.value = data.tokens || 0
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load token balance.'
  } finally {
    loading.value = false
  }
}

async function submit() {
  errors.value = {}
  submitting.value = true
  const { ok, error } = await run(
    () => api.post('/token/share', { tokenValue: Number(form.tokenValue), user_id: Number(form.userId) }),
    { successMessage: 'Tokens sent successfully!' }
  )
  submitting.value = false
  if (ok) {
    form.tokenValue = ''
    form.userId = ''
    await fetchBalance()
  } else if (error?.response?.data?.errors) {
    errors.value = error.response.data.errors
  }
}

onMounted(fetchBalance)
</script>

<template>
  <DashboardLayout>
    <div class="row justify-content-center">
      <div class="col-12 col-md-6 col-lg-5">
        <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>
        <div class="card shadow-lg rounded-4">
          <div class="card-body p-4 p-lg-5">
            <h1 class="h3 mb-3">Share Token</h1>
            <p class="mb-4">My Token : {{ tokens }} USDT</p>

            <form @submit.prevent="submit">
              <div class="mb-3">
                <label class="form-label">Tokens Value</label>
                <input
                  v-model="form.tokenValue"
                  type="number"
                  class="form-control"
                  :class="{ 'is-invalid': errors.tokenValue }"
                  :max="tokens"
                  min="1"
                  required
                  autofocus
                />
                <div v-if="errors.tokenValue" class="invalid-feedback">{{ errors.tokenValue[0] }}</div>
              </div>
              <div class="mb-4">
                <label class="form-label">User ID</label>
                <input
                  v-model="form.userId"
                  type="number"
                  class="form-control"
                  :class="{ 'is-invalid': errors.user_id }"
                  required
                />
                <div v-if="errors.user_id" class="invalid-feedback">{{ errors.user_id[0] }}</div>
              </div>
              <div class="d-grid">
                <button type="submit" class="btn btn-gray-800" :disabled="submitting || loading">Send Tokens</button>
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
