<script setup>
// Ports auth/login.blade.php (route('login') / route('login.post')).
// See auth_handler.go's loginHandler: a `status:'success'` response always
// means the session is good to go to the dashboard; a `status:'error'`
// response ("You do not have any package" / "You're package is pending" /
// "Invalid credentials") is NOT necessarily a failed login — the session
// cookie is issued before those checks run, so we always refresh the auth
// store afterwards and only gate the dashboard redirect on `status`.
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = reactive({ email: '', password: '', remember: false })
const errors = ref({})
const flashMessage = ref('')
const submitting = ref(false)
const showPassword = ref(false)

function togglePassword() {
  showPassword.value = !showPassword.value
}

async function onSubmit() {
  errors.value = {}
  flashMessage.value = ''
  submitting.value = true
  try {
    const { data } = await api.post('/login', { email: form.email, password: form.password })
    // The session cookie is issued as soon as credentials check out, even
    // when the response below is `status:'error'` (no package / pending
    // package) — so refresh the shared auth state regardless of status.
    await authStore.bootstrap()
    if (data.status === 'success') {
      router.push(authStore.dashboardRoute)
    } else {
      flashMessage.value = data.message || 'Invalid credentials'
    }
  } catch (err) {
    if (err?.response?.status === 422) {
      errors.value = err.response.data.errors || {}
    } else {
      flashMessage.value = err?.response?.data?.message || 'Something went wrong. Please try again.'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthCardLayout max-width="500px">
    <div class="text-center text-md-center mb-4 mt-md-0">
      <img class="navbar-brand-dark mb-3" src="/resources/assets/img/brand/logo.png" alt="Volt logo" style="height: 80px !important" />
      <h1 class="mb-0 h3">Sign in to Signetint platform</h1>
    </div>

    <FlashAlert type="danger" :message="flashMessage" @close="flashMessage = ''" />

    <form @submit.prevent="onSubmit" novalidate>
      <div class="mb-3">
        <label class="form-label" for="email">Your Email</label>
        <div class="input-group">
          <span class="input-group-text"><i class="fas fa-envelope"></i></span>
          <input
            id="email"
            v-model="form.email"
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
      </div>

      <div class="mb-3">
        <label class="form-label" for="password">Your Password</label>
        <div class="input-group">
          <span class="input-group-text"><i class="fas fa-lock"></i></span>
          <input
            id="password"
            v-model="form.password"
            :type="showPassword ? 'text' : 'password'"
            name="password"
            class="form-control"
            :class="{ 'is-invalid': errors.password }"
            placeholder="Password"
            required
          />
          <span class="input-group-text" role="button" @click="togglePassword">
            <i :class="showPassword ? 'fas fa-eye-slash' : 'fas fa-eye'"></i>
          </span>
          <div v-if="errors.password" class="invalid-feedback">{{ errors.password[0] }}</div>
        </div>
      </div>

      <div class="d-flex justify-content-between align-items-center mb-4">
        <div class="form-check">
          <input id="remember" v-model="form.remember" type="checkbox" name="remember" class="form-check-input" />
          <label class="form-check-label" for="remember">Remember me</label>
        </div>
        <RouterLink to="/password/reset" class="small text-muted">Lost password?</RouterLink>
      </div>

      <div class="d-grid">
        <button type="submit" class="btn btn-gray-800" :disabled="submitting">
          <span v-if="submitting" class="spinner-border spinner-border-sm me-2"></span>
          Sign in
        </button>
      </div>
    </form>
  </AuthCardLayout>
</template>
