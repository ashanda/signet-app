<script setup>
// Ports admin/setup_google_authenticator.blade.php
// (route('setup.google.auth', $userId)). See google_auth_handler.go's
// setupGoogleAuthHandler: GET /admin/{userId}/setup-google-auth regenerates
// (overwrites) the target user's secret on every call and returns
// {status, secret, otpauth_url} — QR rendering is left to the frontend: the
// otpauth:// URL is rendered here as the same 300px SVG data URI the
// original builds with BaconQrCode (RendererStyle(300), default margin 4).
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

async function load() {
  loading.value = true
  loadError.value = ''
  qrDataUrl.value = ''
  try {
    const { data } = await api.get(`/admin/${props.userId}/setup-google-auth`)
    if (data.status === 'success') {
      secret.value = data.secret
      otpauthUrl.value = data.otpauth_url
      const svg = await QRCode.toString(data.otpauth_url, { type: 'svg', width: 300, margin: 4 })
      qrDataUrl.value = 'data:image/svg+xml;base64,' + btoa(svg)
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
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />
    <div v-if="!loading && !loadError" class="container text-center">
      <h2>Setup Google Authenticator</h2>
      <p>Scan the QR code below with your Google Authenticator app.</p>
      <img :src="qrDataUrl" alt="Google Authenticator QR Code" class="img-fluid" />
      <p><strong>Secret Key:</strong> {{ secret }}</p>
      <p>Use this key if you cannot scan the QR code.</p>
    </div>
  </DashboardLayout>
</template>
