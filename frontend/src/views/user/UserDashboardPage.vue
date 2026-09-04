<script setup>
// Ports UserController@index's dashboard view (route('user.dashboard'),
// implied by "/dashboard" for role=user) — ui_spec.md: "byte-for-byte the
// same component set" as admin/agent's dashboards. See RoleDashboard.vue for
// the shared markup/behavior, and dashboard_handler.go's userDashboardHandler
// for the response shape.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import RoleDashboard from '@/components/widgets/RoleDashboard.vue'
</script>

<template>
  <DashboardLayout>
    <RoleDashboard role-label="User" endpoint="/user/dashboard" activate-endpoint="/active-package" />
  </DashboardLayout>
</template>
