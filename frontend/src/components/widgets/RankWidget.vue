<script setup>
// Ports helpers.php's rank($userId) raw-HTML badge row (ui_spec.md "Helper-
// driven inline widgets"). `rank` prop is the tree.RankResult JSON exactly
// as every dashboard endpoint embeds it: {team_sales_direct, gratitude,
// current_rank, next_rank, remaining_team, remaining_super}.
defineProps({
  rank: { type: Object, default: null },
})

// Matches PHP's number_format() thousands separator (e.g. "4,450").
function nf(n) {
  return Number(n || 0).toLocaleString('en-US')
}
</script>

<template>
  <div v-if="rank" class="d-flex flex-wrap align-items-center gap-3">
    <span><strong>Team Sales:</strong> <span class="badge bg-info text-dark">{{ nf(rank.team_sales_direct) }}</span></span>
    <span><strong>Gratitude:</strong> <span class="badge bg-info text-dark">{{ nf(rank.gratitude) }}</span></span>
    <span><strong>Current Rank:</strong> <span class="badge bg-success">{{ rank.current_rank }}</span></span>
    <span><strong>Next Rank:</strong> <span class="badge bg-warning text-dark">{{ rank.next_rank || 'N/A' }}</span></span>
    <span><strong>Remaining Team:</strong> <span class="badge bg-danger">{{ nf(rank.remaining_team) }}</span></span>
    <span><strong>Remaining Gratitude:</strong> <span class="badge bg-danger">{{ nf(rank.remaining_super) }}</span></span>
  </div>
</template>
