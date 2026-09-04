<script setup>
// Ports agent/dashboard.blade.php (route('agent.dashboard')) — ui_spec.md:
// "byte-for-byte the same component set" as admin/dashboard.blade.php. See
// RoleDashboard.vue for the shared markup/behavior, and
// dashboard_handler.go's agentDashboardHandler for the response shape
// (identical to Admin's minus `token_counts`, which admin's dashboard
// fetches but never renders anyway — see RoleDashboard.vue's header note).
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import RoleDashboard from '@/components/widgets/RoleDashboard.vue'
</script>

<template>
  <DashboardLayout>
    <RoleDashboard role-label="Agent" endpoint="/agent/dashboard" activate-endpoint="/active-package" />
  </DashboardLayout>
</template>
