<script setup>
// Ports helpers.php's roc($userId), which the dashboards call as
// `@if(roc_status == 'active') {!! roc(id) !!} @endif`. The helper echoes
// its markup unconditionally (a leading <br>, then the summary line with
// empty values when there's no weekly summary yet), so this renders
// whenever roc_status is active, whether or not `roc` has data.
// `roc` is the tree.RocSummary JSON: {week_start, week_end, per_week_total}.
defineProps({
  roc: { type: Object, default: null },
  rocStatus: { type: String, default: '' },
})
</script>

<template>
  <template v-if="rocStatus === 'active'">
    <br />
    <div class="d-flex flex-wrap align-items-center gap-3">
      <span><strong>Last Week Summary:</strong> {{ roc?.week_start ?? '' }} - {{ roc?.week_end ?? '' }}</span>
      <span><strong>Total sales per week:</strong> {{ roc?.per_week_total ?? '' }}</span>
    </div>
  </template>
</template>
