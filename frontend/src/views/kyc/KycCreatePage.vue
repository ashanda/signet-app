<script setup>
// Ports kyc/create.blade.php — see components/shared/KycForm.vue.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import KycForm from '@/components/shared/KycForm.vue'
</script>

<template>
  <DashboardLayout>
    <KycForm />
  </DashboardLayout>
</template>
