<script setup>
// Ports auth/password_reset_form.blade.php (route('password.reset', $token) /
// route('password.update')). See passwordResetTokenHandler / passwordResetHandler
// in auth_handler.go.
import { onMounted, reactive, ref } from 'vue'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const props = defineProps({
  token: { type: String, required: true },
})

const form = reactive({ email: '', password: '', password_confirmation: '' })
const errors = ref({})
const flashMessage = ref('')
const flashType = ref('danger')
const loading = ref(true)
const submitting = ref(false)
const done = ref(false)

onMounted(async () => {
  try {
    const { data } = await api.get(`/password/reset/${props.token}`)
    if (data.status === 'success') {
      form.email = data.email || ''
    } else {
      flashMessage.value = data.message || 'This password reset token is invalid.'
      flashType.value = 'danger'
    }
  } catch (err) {
    flashMessage.value = err?.response?.data?.message || 'This password reset token is invalid.'
    flashType.value = 'danger'
  } finally {
    loading.value = false
  }
})

async function onSubmit() {
  errors.value = {}
  flashMessage.value = ''
  submitting.value = true
  try {
    const { data } = await api.post('/password/reset', {
      token: props.token,
      email: form.email,
      password: form.password,
      password_confirmation: form.password_confirmation,
    })
    if (data.status === 'success') {
      done.value = true
      flashType.value = 'success'
      flashMessage.value = data.message || 'Your password has been reset!'
    } else {
      flashType.value = 'danger'
      flashMessage.value = data.message || 'Something went wrong. Please try again.'
    }
  } catch (err) {
    if (err?.response?.status === 422) {
      errors.value = err.response.data.errors || {}
    } else {
      flashType.value = 'danger'
      flashMessage.value = err?.response?.data?.message || 'Something went wrong. Please try again.'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthCardLayout max-width="500px" back-text="Back to log in" back-to="/login">
    <h1 class="h3 mb-4">Reset password</h1>

    <FlashAlert :type="flashType" :message="flashMessage" @close="flashMessage = ''" />

    <div v-if="loading" class="text-center text-muted py-4">Loading&hellip;</div>

    <div v-else-if="done" class="text-center py-3">
      <RouterLink to="/login" class="btn btn-gray-800">Go to Login</RouterLink>
    </div>

    <form v-else @submit.prevent="onSubmit" novalidate>
      <input type="hidden" name="token" :value="token" />

      <div class="mb-3">
        <label class="form-label" for="email">Your Email</label>
        <input
          id="email"
          v-model="form.email"
          type="email"
          name="email"
          class="form-control"
          :class="{ 'is-invalid': errors.email }"
          placeholder="example@company.com"
          required
        />
        <div v-if="errors.email" class="invalid-feedback">{{ errors.email[0] }}</div>
      </div>

      <div class="mb-3">
        <label class="form-label" for="password">Password</label>
        <div class="input-group">
          <span class="input-group-text"><i class="fas fa-lock"></i></span>
          <input
            id="password"
            v-model="form.password"
            type="password"
            name="password"
            class="form-control"
            :class="{ 'is-invalid': errors.password }"
            placeholder="Password"
            required
          />
          <div v-if="errors.password" class="invalid-feedback">{{ errors.password[0] }}</div>
        </div>
      </div>

      <div class="mb-4">
        <label class="form-label" for="password_confirmation">Confirm Password</label>
        <div class="input-group">
          <span class="input-group-text"><i class="fas fa-lock"></i></span>
          <input
            id="password_confirmation"
            v-model="form.password_confirmation"
            type="password"
            name="password_confirmation"
            class="form-control"
            placeholder="Confirm Password"
            required
          />
        </div>
      </div>

      <div class="d-grid">
        <button type="submit" class="btn btn-gray-800" :disabled="submitting">
          <span v-if="submitting" class="spinner-border spinner-border-sm me-2"></span>
          Reset password
        </button>
      </div>
    </form>
  </AuthCardLayout>
</template>
