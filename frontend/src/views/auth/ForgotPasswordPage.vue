<script setup>
// Ports auth/password_reset.blade.php (route('password.request') /
// route('password.email')). See passwordEmailHandler in auth_handler.go.
import { ref } from 'vue'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'

const email = ref('')
const errors = ref({})
const statusMessage = ref('')
const submitting = ref(false)

async function onSubmit() {
  errors.value = {}
  statusMessage.value = ''
  submitting.value = true
  try {
    const { data } = await api.post('/password/email', { email: email.value })
    statusMessage.value = data.message || 'We have emailed your password reset link!'
  } catch (err) {
    if (err?.response?.status === 422) {
      errors.value = err.response.data.errors || {}
    } else {
      errors.value = { email: [err?.response?.data?.message || 'Something went wrong. Please try again.'] }
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthCardLayout
    back-text="Back to log in"
    back-to="/login"
    back-placement="row"
    card-class="signin-inner my-3 my-lg-0 bg-white shadow border-0 rounded p-4 p-lg-5 w-100 fmxw-500"
  >
    <h1 class="h3">Forgot your password?</h1>
    <p class="mb-4">Don't fret! Just type in your email and we will send you a code to reset your password!</p>
    <div v-if="statusMessage" class="alert alert-success" role="alert">
      {{ statusMessage }}
    </div>
    <form method="POST" @submit.prevent="onSubmit">
      <!-- Form -->
      <div class="mb-4">
        <label for="email">Your Email</label>
        <div class="input-group">
          <input id="email" v-model="email" type="email" class="form-control" :class="{ 'is-invalid': errors.email }" name="email" required autofocus />
        </div>
        <span v-if="errors.email" class="invalid-feedback" role="alert">
          <strong>{{ errors.email[0] }}</strong>
        </span>
      </div>
      <!-- End of Form -->
      <div class="d-grid">
        <button type="submit" class="btn btn-gray-800" :disabled="submitting">
          Recover password
        </button>
      </div>
    </form>
  </AuthCardLayout>
</template>
