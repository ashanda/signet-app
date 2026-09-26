<script setup>
// Ports auth/register_step2.blade.php — on-screen labelled "Step 2: Select a
// Package" (route('register.step2', $id) / posts to register.processStep2).
// See registerStep2FormHandler / registerStep2SubmitHandler in auth_handler.go.
//
// The POST response's `parent` object (Binance ID / WhatsApp / vacation flag)
// is only ever returned here, in-band — there's no separate GET to refetch it
// — so it's stashed in sessionStorage for the Step 3 "wait for upliner" page
// to read after navigation.
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api/client'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const props = defineProps({
  id: { type: [String, Number], required: true },
})

const router = useRouter()

const loading = ref(true)
const loadError = ref('')
const packages = ref([])
const selectedPackage = ref('')
const errors = ref({})
const formError = ref('')
const submitting = ref(false)

onMounted(async () => {
  try {
    const { data } = await api.get(`/register/step2/${props.id}`)
    packages.value = data.packages || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load packages.'
  } finally {
    loading.value = false
  }
})

async function onSubmit() {
  errors.value = {}
  formError.value = ''
  submitting.value = true
  try {
    const { data } = await api.post('/register/step2', {
      package: selectedPackage.value,
      newUserID: String(props.id),
    })
    if (data.status === 'success') {
      try {
        sessionStorage.setItem(`signet_register_parent_${props.id}`, JSON.stringify(data.parent))
      } catch {
        // sessionStorage unavailable — Step 3 will just render without parent details.
      }
      router.push({ name: 'register.step3.wait', params: { id: data.user_id } })
    } else {
      formError.value = data.message || 'Something went wrong. Please try again.'
    }
  } catch (err) {
    if (err?.response?.status === 422) {
      errors.value = err.response.data.errors || {}
    } else {
      formError.value = err?.response?.data?.message || 'Something went wrong. Please try again.'
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <AuthCardLayout back-text="Back to log in" back-to="/login" back-placement="row">
    <h1 class="h3 mb-4">Step 2: Select a Package</h1>
    <FlashAlert type="danger" :message="loadError || formError" @close="formError = ''" />
    <form method="POST" @submit.prevent="onSubmit">
      <div class="form-group">
        <label for="package">Choose your package</label>
        <input type="hidden" name="newUserID" :value="id" />
        <select v-model="selectedPackage" name="package" class="form-control" required>
          <option value="">Select Package</option>
          <option v-for="pack in packages" :key="pack.id" :value="String(pack.id)">{{ pack.name }} USD</option>
        </select>
      </div>
      <button type="submit" class="btn btn-primary mt-4" :disabled="submitting">Next</button>
    </form>
    <div class="form-group"></div>
  </AuthCardLayout>
</template>
