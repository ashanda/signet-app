<script setup>
// Ports helpers.php's roc($userId) raw-HTML line (ui_spec.md "Helper-driven
// inline widgets") — only rendered when the user's roc_status == 'active'.
// `roc` prop is the tree.RocSummary JSON: {week_start, week_end,
// per_week_total} (all null when there's no data yet, per tree.go's Roc()).
defineProps({
  roc: { type: Object, default: null },
  rocStatus: { type: String, default: '' },
})
</script>

<template>
  <div v-if="rocStatus === 'active' && roc && roc.week_start" class="d-flex flex-wrap align-items-center gap-3">
    <span><strong>Last Week Summary:</strong> {{ roc.week_start }} - {{ roc.week_end }}</span>
    <span><strong>Total sales per week:</strong> {{ roc.per_week_total }}</span>
  </div>
</template>
