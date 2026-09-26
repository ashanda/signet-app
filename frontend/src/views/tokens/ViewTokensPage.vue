<script setup>
// Ports tokens/index.blade.php ("Tokens for {name}" — route('view.tokens',
// $userId), company). See token_handler.go's viewTokensHandler (GET
// /tokens/view/{userId}) and generateTokensHandler (POST
// /tokens/generate/{userId}).
//
// As in the original (redirect back with a session flash), the outcome of
// "Generate Tokens" is shown as a success/danger alert above the form.
// generateTokensHandler answers HTTP 200 with {status:'error', message} for
// its logical failures ("Google Authenticator secret not set for the user"
// / "Invalid Google Authenticator code").
//
// The Google Auth Code field is a text input (numeric keypad) rather than
// the original's type="number", which dropped leading zeros from codes.
import { reactive, ref, onMounted, watch } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'

const props = defineProps({
  userId: { type: [String, Number], required: true },
})

const loadError = ref('')
const user = ref(null)
const tokens = ref(null)

const form = reactive({ token_count: '', google_auth_code: '' })
const flashSuccess = ref('')
const generating = ref(false)

async function fetchData(page = 1) {
  loadError.value = ''
  try {
    const { data } = await api.get(`/tokens/view/${props.userId}`, { params: { page } })
    user.value = data.user
    tokens.value = data.tokens
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load tokens.'
  }
}

async function generate() {
  flashSuccess.value = ''
  loadError.value = ''
  generating.value = true
  try {
    const { data } = await api.post(`/tokens/generate/${props.userId}`, {
      token_count: Number(form.token_count),
      google_auth_code: form.google_auth_code,
    })
    if (data.status === 'error') {
      loadError.value = data.message
    } else {
      flashSuccess.value = data.message
      form.token_count = ''
      form.google_auth_code = ''
      await fetchData(1)
    }
  } catch (err) {
    const errs = err?.response?.data?.errors
    loadError.value = (errs && Object.values(errs)[0]?.[0]) || err?.response?.data?.message || 'Something went wrong'
  } finally {
    generating.value = false
  }
}

onMounted(() => fetchData())
watch(() => props.userId, () => fetchData())
</script>

<template>
  <DashboardLayout>
    <h2 class="mb-3">Tokens for {{ user?.name }}</h2>
    <div v-if="flashSuccess" class="alert alert-success">{{ flashSuccess }}</div>
    <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>
    <div class="d-flex justify-content-between w-100 flex-wrap">
      <div class="mb-3 mb-lg-0 col-6">
        <form method="POST" class="mb-3" @submit.prevent="generate">
          <label for="token_count" class="form-label">Enter Number of Tokens:</label>
          <input id="token_count" v-model="form.token_count" type="number" name="token_count" class="form-control mb-2" min="1" max="500" required />
          <label for="token_count" class="form-label">Enter Google Auth Code:</label>
          <input
            v-model="form.google_auth_code"
            type="text"
            inputmode="numeric"
            pattern="[0-9]*"
            autocomplete="one-time-code"
            name="google_auth_code"
            class="form-control mb-2"
            required
          />
          <button type="submit" class="btn btn-primary" :disabled="generating">Generate Tokens</button>
        </form>
      </div>
    </div>
    <table class="table">
      <thead>
        <tr>
          <th>#</th>
          <th>Token</th>
          <th>Status</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(token, index) in tokens?.data" :key="token.id">
          <td>{{ index + 1 }}</td>
          <td>{{ token.token }}</td>
          <td>
            <span class="badge" :class="token.status == 'active' ? 'bg-success' : 'bg-danger'">
              {{ token.status ? token.status.charAt(0).toUpperCase() + token.status.slice(1) : '' }}
            </span>
          </td>
        </tr>
      </tbody>
    </table>
    <!-- Pagination Links -->
    <div class="d-flex justify-content-center">
      <Paginator :pagination="tokens" @change="(p) => fetchData(p)" />
    </div>
  </DashboardLayout>
</template>
