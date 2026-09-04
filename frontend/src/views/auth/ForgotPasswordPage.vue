<script setup>
// Ports auth/password_reset.blade.php (route('password.request') /
// route('password.email')). See passwordEmailHandler in auth_handler.go.
import { ref } from 'vue'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

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
  <AuthCardLayout max-width="500px" back-text="Back to log in" back-to="/login">
    <h1 class="h3 mb-2">Forgot your password?</h1>
    <p class="text-muted mb-4">
      Enter your email address and we'll send you a link to reset your password.
    </p>

    <FlashAlert type="success" :message="statusMessage" @close="statusMessage = ''" />

    <form @submit.prevent="onSubmit" novalidate>
      <div class="mb-4">
        <label class="form-label" for="email">Your Email</label>
        <input
          id="email"
          v-model="email"
          type="email"
          name="email"
          class="form-control"
          :class="{ 'is-invalid': errors.email }"
          placeholder="example@company.com"
          autofocus
          required
        />
        <div v-if="errors.email" class="invalid-feedback">{{ errors.email[0] }}</div>
      </div>

      <div class="d-grid">
        <button type="submit" class="btn btn-gray-800" :disabled="submitting">
          <span v-if="submitting" class="spinner-border spinner-border-sm me-2"></span>
          Recover password
        </button>
      </div>
    </form>
  </AuthCardLayout>
</template>
