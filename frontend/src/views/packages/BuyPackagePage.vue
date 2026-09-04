<script setup>
// Ports packages/buy-package.blade.php ("Buy Package" — route('buy.package')
// / posts route('buy.packages'), any authenticated). See
// package_handler.go's buyPackageHandler (GET /buy-package: active
// packages only) and buyPackagesHandler (POST /buy-packages).
//
// buyPackagesHandler's success response is {status:'success',
// parent_data:<models.User>} — no separate "done" page endpoint exists to
// re-fetch this from, so per FRONTEND_CONVENTIONS the response is stashed
// (sessionStorage) and the SPA navigates to buy.package.done, which reads
// it back rather than re-fetching.
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useApiAction } from '@/composables/useApiAction'

const router = useRouter()
const { run } = useApiAction()

const loading = ref(true)
const loadError = ref('')
const packages = ref([])
const selectedPackage = ref('')
const submitting = ref(false)

async function fetchPackages() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/buy-package')
    packages.value = data.packages || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load packages.'
  } finally {
    loading.value = false
  }
}

async function submit() {
  if (!selectedPackage.value) return
  submitting.value = true
  // No separate "saved" alert — the response drives the very next screen
  // (Upliner Activation), matching the original's direct-redirect flow.
  const { ok, data } = await run(() => api.post('/buy-packages', { package: String(selectedPackage.value) }), {
    showSuccessAlert: false,
  })
  submitting.value = false
  if (ok) {
    sessionStorage.setItem('signet:buyPackageDone', JSON.stringify(data.parent_data))
    router.push({ name: 'buy.package.done' })
  }
}

onMounted(fetchPackages)
</script>

<template>
  <AuthCardLayout :show-back-link="false">
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <form @submit.prevent="submit">
      <div class="form-group">
        <label for="package">Choose your package</label>
        <select id="package" v-model="selectedPackage" class="form-control" required>
          <option value="">Select Package</option>
          <option v-for="pkg in packages" :key="pkg.id" :value="pkg.id">{{ pkg.name }} USD</option>
        </select>
      </div>
      <button type="submit" class="btn btn-primary mt-4" :disabled="submitting || loading">Next</button>
    </form>
  </AuthCardLayout>
</template>
