<script setup>
// Ports packages/buy-package-done.blade.php (post-purchase "Upliner
// Activation" wait screen — the original renders this as the direct
// server response to the buy.packages POST, no separate GET route).
//
// There is no re-fetch endpoint for this screen's data (see
// package_handler.go's buyPackagesHandler doc comment) — BuyPackagePage.vue
// stashes the POST /buy-packages response's `parent_data` in
// sessionStorage before navigating here; this page just reads it back. A
// direct/refreshed visit with nothing stashed shows a fallback instead of
// fabricating data.
//
// parent_data is a raw models.User struct: binance_pay_id/whatsapp_number
// are sql.NullString (no custom MarshalJSON), so they may arrive as
// {String,Valid} objects — handled defensively, same as elsewhere.
import { ref, onMounted } from 'vue'
import AuthCardLayout from '@/components/layout/AuthCardLayout.vue'

function nsGet(v) {
  if (v == null) return ''
  if (typeof v === 'object') return v.Valid ? v.String : ''
  return v
}

const parentData = ref(null)

onMounted(() => {
  try {
    const raw = sessionStorage.getItem('signet:buyPackageDone')
    if (raw) parentData.value = JSON.parse(raw)
  } catch {
    parentData.value = null
  }
})
</script>

<template>
  <AuthCardLayout :show-back-link="false">
    <template v-if="parentData">
      <div class="mt-4">
        <h3>Upliner Activation</h3>
        Binance ID: {{ nsGet(parentData.binance_pay_id) || '—' }}
        <br />
        <i class="fab fa-whatsapp mt-2"></i> WhatsApp no: {{ nsGet(parentData.whatsapp_number) || '—' }}
        <br />
        <a
          v-if="nsGet(parentData.whatsapp_number)"
          :href="`tel:${nsGet(parentData.whatsapp_number)}`"
          class="btn btn-success mt-2"
        >
          <i class="fas fa-phone-alt"></i> Call Now
        </a>
      </div>
      <router-link :to="{ name: 'buy.package.history' }" class="btn btn-primary mt-4">Next</router-link>
    </template>
    <template v-else>
      <h3>Upliner Activation</h3>
      <p class="text-muted mt-3">
        We don't have activation details to show right now. If you just bought a package, go back and try again.
      </p>
      <router-link :to="{ name: 'buy.package' }" class="btn btn-primary">Buy Package</router-link>
    </template>
  </AuthCardLayout>
</template>
