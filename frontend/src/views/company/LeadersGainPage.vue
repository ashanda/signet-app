<script setup>
// Ports leaders/index.blade.php (route('leaders.gain')). The page body is shared with the
// other code-gain report — see components/shared/CodeGainReport.vue.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import CodeGainReport from '@/components/shared/CodeGainReport.vue'
</script>

<template>
  <DashboardLayout>
    <CodeGainReport kind="leader" />
  </DashboardLayout>
</template>
