<script setup>
// Ports tokens/index.blade.php ("Tokens for {name}" — route('view.tokens',
// $userId), company). See token_handler.go's viewTokensHandler (GET
// /tokens/view/{userId}) and generateTokensHandler (POST
// /tokens/generate/{userId}).
//
// generateTokensHandler answers HTTP 200 for BOTH success and the two
// logical failures ("Google Authenticator secret not set for the user" /
// "Invalid Google Authenticator code") via {status:'error',...} — handled
// automatically by useApiAction. Field-level 422s (token_count out of
// 1..500 range, missing google_auth_code) are shown inline.
import { reactive, ref, onMounted, watch } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'

const props = defineProps({
  userId: { type: [String, Number], required: true },
})

const { run } = useApiAction()

const loading = ref(true)
const loadError = ref('')
const user = ref(null)
const tokens = ref(null)

const form = reactive({ token_count: '', google_auth_code: '' })
const errors = ref({})
const generating = ref(false)

async function fetchData(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get(`/tokens/view/${props.userId}`, { params: { page } })
    user.value = data.user
    tokens.value = data.tokens
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load tokens.'
  } finally {
    loading.value = false
  }
}

async function generate() {
  errors.value = {}
  generating.value = true
  const { ok, error } = await run(
    () =>
      api.post(`/tokens/generate/${props.userId}`, {
        token_count: Number(form.token_count),
        google_auth_code: form.google_auth_code,
      }),
    { showSuccessAlert: true }
  )
  generating.value = false
  if (ok) {
    form.token_count = ''
    form.google_auth_code = ''
    await fetchData(1)
  } else if (error?.response?.data?.errors) {
    errors.value = error.response.data.errors
  }
}

onMounted(() => fetchData())
watch(() => props.userId, () => fetchData())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h2 class="h4">Tokens for {{ user?.name }}</h2>
    </div>

    <div class="row mb-4">
      <div class="col-12 col-md-6">
        <div class="card border-0 shadow">
          <div class="card-body">
            <h5 class="card-title">Generate Tokens</h5>
            <form @submit.prevent="generate">
              <div class="mb-3">
                <label class="form-label">Number of Tokens</label>
                <input
                  v-model="form.token_count"
                  type="number"
                  min="1"
                  max="500"
                  class="form-control"
                  :class="{ 'is-invalid': errors.token_count }"
                  required
                />
                <div v-if="errors.token_count" class="invalid-feedback">{{ errors.token_count[0] }}</div>
              </div>
              <div class="mb-3">
                <label class="form-label">Google Auth Code</label>
                <input
                  v-model="form.google_auth_code"
                  type="text"
                  inputmode="numeric"
                  pattern="[0-9]*"
                  autocomplete="one-time-code"
                  class="form-control"
                  :class="{ 'is-invalid': errors.google_auth_code }"
                  required
                />
                <div v-if="errors.google_auth_code" class="invalid-feedback">{{ errors.google_auth_code[0] }}</div>
              </div>
              <button type="submit" class="btn btn-primary" :disabled="generating">Generate Tokens</button>
            </form>
          </div>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom">#</th>
                <th class="border-bottom">Token</th>
                <th class="border-bottom">Status</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="!loading && !tokens?.data?.length">
                <td colspan="3" class="text-center text-muted py-4">No tokens found</td>
              </tr>
              <tr v-for="(row, idx) in tokens?.data" :key="row.id">
                <td>{{ (tokens.from || 1) + idx }}</td>
                <td>{{ row.token }}</td>
                <td>
                  <span class="badge" :class="row.status === 'active' ? 'bg-success' : 'bg-danger'">
                    {{ row.status === 'active' ? 'Active' : 'Inactive' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <Paginator :pagination="tokens" @change="(p) => fetchData(p)" />
      </div>
    </div>
  </DashboardLayout>
</template>
