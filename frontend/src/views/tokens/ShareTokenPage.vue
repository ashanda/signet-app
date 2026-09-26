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
import Swal from 'sweetalert2'

const loading = ref(true)
const loadError = ref('')
const tokens = ref(0)

const form = reactive({ tokenValue: '', userId: '' })
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

// Alert::success('Success', ...) / Alert::error('Error', ...) + redirect back.
async function submit() {
  submitting.value = true
  try {
    const { data } = await api.post('/token/share', { tokenValue: Number(form.tokenValue), user_id: Number(form.userId) })
    if (data.status === 'error') {
      await Swal.fire({ icon: 'error', title: 'Error', text: data.message })
    } else {
      await Swal.fire({ icon: 'success', title: 'Success', text: 'Tokens sent successfully!' })
      form.tokenValue = ''
      form.userId = ''
      await fetchBalance()
    }
  } catch (err) {
    const errs = err?.response?.data?.errors
    const message = (errs && Object.values(errs)[0]?.[0]) || err?.response?.data?.message || 'Something went wrong'
    await Swal.fire({ icon: 'error', title: 'Error', text: message })
  } finally {
    submitting.value = false
  }
}

onMounted(fetchBalance)
</script>

<template>
  <DashboardLayout>
    <!-- Section -->
    <section class="vh-lg-100 mt-5 mt-lg-0 bg-soft d-flex align-items-center">
      <div class="container">
        <div class="row justify-content-center form-bg-image">
          <div class="col-12 d-flex align-items-center justify-content-center">
            <div class="bg-white shadow border-0 rounded border-light p-4 p-lg-5 w-100 fmxw-500">
              <!-- The original's unterminated "</a" swallows this block's
                   closing </div>, so the form renders inside the centred
                   heading block. Kept, since that is what users see. -->
              <div class="text-center text-md-center mb-4 mt-md-0">
                <h1 class="mb-0 h3">Share Token</h1>
                <p>My Token : {{ tokens }} USDT</p>
                <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>
                <form class="mt-4" method="POST" @submit.prevent="submit">
                  <!-- Form -->
                  <div class="form-group mb-4">
                    <label for="text">Tokens Value</label>
                    <div class="input-group">
                      <input id="tokenValue" v-model="form.tokenValue" type="number" :max="tokens" min="1" class="form-control" placeholder="10" name="tokenValue" autofocus required />
                    </div>
                  </div>
                  <!-- End of Form -->
                  <div class="form-group">
                    <!-- Form -->
                    <div class="form-group mb-4">
                      <label for="user_id">User ID</label>
                      <div class="input-group">
                        <input id="user_id" v-model="form.userId" type="number" placeholder="User ID" class="form-control" name="user_id" required />
                      </div>
                    </div>
                    <!-- End of Form -->
                  </div>
                  <div class="d-grid">
                    <button type="submit" class="btn btn-gray-800" :disabled="submitting || loading">Send Tokens</button>
                  </div>
                </form>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>
  </DashboardLayout>
</template>
