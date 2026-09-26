<script setup>
// Ports kyc/edit.blade.php — see components/shared/KycForm.vue.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import KycForm from '@/components/shared/KycForm.vue'

defineProps({
  id: { type: [String, Number], required: true },
})
</script>

<template>
  <DashboardLayout>
    <KycForm :id="id" />
  </DashboardLayout>
</template>
