<script setup>
// Renders the Bootstrap-styled prev/next + numbered-page control every list
// page uses (ui_spec.md "Tables" section: Laravel's default paginator,
// centered `d-flex justify-content-center` below the table). Consumes the
// httpx.Paginated envelope verbatim (current_page/last_page/total/...).
const props = defineProps({
  pagination: { type: Object, default: null },
})
const emit = defineEmits(['change'])

function pageNumbers() {
  if (!props.pagination) return []
  const { current_page, last_page } = props.pagination
  const span = 2
  const start = Math.max(1, current_page - span)
  const end = Math.min(last_page, current_page + span)
  const out = []
  for (let p = start; p <= end; p++) out.push(p)
  return out
}

function go(p) {
  if (!props.pagination) return
  if (p < 1 || p > props.pagination.last_page || p === props.pagination.current_page) return
  emit('change', p)
}
</script>

<template>
  <div v-if="pagination && pagination.last_page > 1" class="d-flex justify-content-center mt-3">
    <nav>
      <ul class="pagination mb-0">
        <li class="page-item" :class="{ disabled: pagination.current_page <= 1 }">
          <a class="page-link" href="#" @click.prevent="go(pagination.current_page - 1)">&laquo;</a>
        </li>
        <li v-for="p in pageNumbers()" :key="p" class="page-item" :class="{ active: p === pagination.current_page }">
          <a class="page-link" href="#" @click.prevent="go(p)">{{ p }}</a>
        </li>
        <li class="page-item" :class="{ disabled: pagination.current_page >= pagination.last_page }">
          <a class="page-link" href="#" @click.prevent="go(pagination.current_page + 1)">&raquo;</a>
        </li>
      </ul>
    </nav>
  </div>
</template>
