<script setup>
// Ports auth/register_step3.blade.php — on-screen labelled "Step 3: Get Your
// Upliner Details" (route('register.step3')), the intermediate "wait for
// upliner" screen shown right after package selection (processStep2).
//
// Per the ui_spec.md file/label mismatch note, the original's "Next" button
// links back to this same view; the rebuild instead sends the user on to the
// dedicated Step 4 "Account Status" screen at /register/status/:id.
import { computed, ref } from 'vue'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'

const props = defineProps({
  id: { type: [String, Number], required: true },
})

const parent = ref(null)

try {
  const raw = sessionStorage.getItem(`signet_register_parent_${props.id}`)
  if (raw) parent.value = JSON.parse(raw)
} catch {
  parent.value = null
}

const whatsappDigits = computed(() => (parent.value?.whatsapp_number || '').replace(/[^\d+]/g, ''))
</script>

<template>
  <AuthCardLayout max-width="500px" back-text="Back to log in" back-to="/login">
    <h1 class="h3 mb-1">Step 3: Get Your Upliner Details</h1>
    <h2 class="h6 text-muted mb-4">Upliner Activation</h2>

    <div v-if="!parent" class="alert alert-warning">
      We couldn't find your upliner's details for this session. Please continue below to check your account status.
    </div>

    <template v-else>
      <span v-if="parent.on_vacation" class="badge bg-warning text-dark mb-3">Your Upliner is On Vacation</span>

      <div class="mb-3">
        <div class="text-muted small">Binance ID</div>
        <div class="fw-semibold">{{ parent.binance_pay_id || 'N/A' }}</div>
      </div>

      <div class="mb-4">
        <div class="text-muted small">WhatsApp Number</div>
        <div class="fw-semibold">
          <i class="fab fa-whatsapp text-success me-1"></i>{{ parent.whatsapp_number || 'N/A' }}
        </div>
      </div>

      <a v-if="whatsappDigits" :href="`tel:${whatsappDigits}`" class="btn btn-success w-100 mb-3">
        <i class="fas fa-phone me-2"></i>Call Now
      </a>
    </template>

    <RouterLink :to="{ name: 'register.step3.status', params: { id } }" class="btn btn-primary w-100">Next</RouterLink>
  </AuthCardLayout>
</template>
