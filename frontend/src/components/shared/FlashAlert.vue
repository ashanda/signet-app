<script setup>
// Ports the `@if(session('success'))/@if(session('error'))` inline alert
// pattern repeated at the top of nearly every page (ui_spec.md "Flash
// messages / alerts"). Since this is now a JSON API rather than
// server-rendered redirects with flashed session data, pages set these from
// the JSON response's `status`/`message` fields after an action, or pass a
// route-query-string message through (e.g. after a redirect-like navigation).
defineProps({
  type: { type: String, default: 'success' }, // 'success' | 'danger' | 'error' | 'info'
  message: { type: String, default: '' },
  dismissible: { type: Boolean, default: true },
})
const emit = defineEmits(['close'])

function cls(type) {
  return type === 'error' ? 'alert-danger' : `alert-${type}`
}
</script>

<template>
  <div v-if="message" class="alert" :class="[cls(type), { 'alert-dismissible fade show': dismissible }]" role="alert">
    {{ message }}
    <button v-if="dismissible" type="button" class="btn-close" @click="emit('close')"></button>
  </div>
</template>
