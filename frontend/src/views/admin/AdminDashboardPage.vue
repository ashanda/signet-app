<script setup>
// Ports admin/dashboard.blade.php (route('admin.dashboard')). ui_spec.md
// states Admin/Agent/User dashboards are "byte-for-byte the same component
// set" — the shared markup/behavior lives in RoleDashboard.vue, this file
// only supplies the role label + endpoint. See dashboard_handler.go's
// adminDashboardHandler for the response shape.
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import RoleDashboard from '@/components/widgets/RoleDashboard.vue'
</script>

<template>
  <DashboardLayout>
    <RoleDashboard role-label="Admin" endpoint="/admin/dashboard" activate-endpoint="/active-package" />
  </DashboardLayout>
</template>
