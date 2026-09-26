import { onMounted, onBeforeUnmount } from 'vue'
import $ from 'jquery'
import select2Factory from 'select2'

// Registers $.fn.select2 on the bundled jQuery (the original pages load
// jQuery + Select2 4.1.0-rc.0 from a CDN).
select2Factory(window, $)

// Initialises Select2 on the <select> bound to `elRef` with the original
// page's options; `onChange(value)` fires when the selection changes.
// Select2 owns the element's DOM, so the <select> should not be given
// v-model or v-for options by the caller.
export function useSelect2(elRef, options = () => ({}), onChange = () => {}) {
  let $el = null

  onMounted(() => {
    $el = $(elRef.value)
    $el.select2(options())
    $el.on('change', () => onChange($el.val()))
  })

  onBeforeUnmount(() => {
    if ($el && $el.data('select2')) $el.select2('destroy')
    $el = null
  })

  return {
    // $('#el').val(value).trigger('change')
    set(value) {
      if ($el) $el.val(value).trigger('change')
    },
  }
}
