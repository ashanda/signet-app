<script setup>
// Ports executive_code_logs/index.blade.php. The page body is shared with the other
// code-log report — see components/shared/CodeLogsReport.vue.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import CodeLogsReport from '@/components/shared/CodeLogsReport.vue'
</script>

<template>
  <DashboardLayout>
    <CodeLogsReport kind="executive" />
  </DashboardLayout>
</template>
