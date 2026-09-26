<script setup>
// Ports kyc/index.blade.php. The page body is shared with the other KYC list
// view — see components/shared/KycTable.vue.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import KycTable from '@/components/shared/KycTable.vue'
</script>

<template>
  <DashboardLayout>
    <KycTable />
  </DashboardLayout>
</template>
