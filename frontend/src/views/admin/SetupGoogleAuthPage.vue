<script setup>
// Ports admin/setup_google_authenticator.blade.php
// (route('setup.google.auth', $userId)). See google_auth_handler.go's
// setupGoogleAuthHandler: GET /admin/{userId}/setup-google-auth regenerates
// (overwrites) the target user's secret on every call and returns
// {status, secret, otpauth_url} — QR rendering is left to the frontend, so
// the otpauth:// URL is rendered client-side with the `qrcode` package. If
// that ever fails (e.g. offline/blocked), fall back to showing just the
// secret key with a manual-entry note rather than a broken image.
import { ref, watch, onMounted } from 'vue'
import QRCode from 'qrcode'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const props = defineProps({
  userId: { type: [String, Number], required: true },
})

const loading = ref(true)
const loadError = ref('')
const secret = ref('')
const otpauthUrl = ref('')
const qrDataUrl = ref('')
const qrFailed = ref(false)

async function load() {
  loading.value = true
  loadError.value = ''
  qrDataUrl.value = ''
  qrFailed.value = false
  try {
    const { data } = await api.get(`/admin/${props.userId}/setup-google-auth`)
    if (data.status === 'success') {
      secret.value = data.secret
      otpauthUrl.value = data.otpauth_url
      try {
        qrDataUrl.value = await QRCode.toDataURL(data.otpauth_url, { width: 260, margin: 1 })
      } catch {
        qrFailed.value = true
      }
    } else {
      loadError.value = data.message || 'Could not set up Google Authenticator.'
    }
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not set up Google Authenticator.'
  } finally {
    loading.value = false
  }
}

onMounted(load)
watch(() => props.userId, load)
</script>

<template>
  <DashboardLayout>
    <div class="text-center mx-auto" style="max-width: 420px;">
      <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

      <template v-if="!loading && !loadError">
        <h2 class="h3 mb-3">Setup Google Authenticator</h2>
        <p class="text-muted">Scan the QR code below with your Google Authenticator app.</p>

        <img
          v-if="qrDataUrl"
          :src="qrDataUrl"
          alt="Google Authenticator QR code"
          class="img-fluid mb-3"
        />
        <div v-else-if="qrFailed" class="alert alert-warning">
          Scan is unavailable, enter this key manually.
        </div>

        <p class="mb-1"><strong>Secret Key:</strong> {{ secret }}</p>
        <p class="text-muted small">Use this key if you cannot scan the QR code.</p>
      </template>
    </div>
  </DashboardLayout>
</template>
