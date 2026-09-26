import { ref, onBeforeUnmount } from 'vue'
import bootstrap from 'bootstrap/dist/js/bootstrap.bundle.min.js'

// A real Bootstrap modal (fade, backdrop, Esc/backdrop-click to close) for
// markup ported from the Blade views' `.modal.fade` dialogs. Bind the
// returned ref to the modal's root element.
export function useBsModal() {
  const el = ref(null)
  const instance = () => (el.value ? bootstrap.Modal.getOrCreateInstance(el.value) : null)

  onBeforeUnmount(() => {
    const m = el.value && bootstrap.Modal.getInstance(el.value)
    if (m) m.dispose()
    // A modal disposed while open leaves its backdrop and body lock behind.
    document.querySelectorAll('.modal-backdrop').forEach((b) => b.remove())
    document.body.classList.remove('modal-open')
    document.body.style.removeProperty('overflow')
    document.body.style.removeProperty('padding-right')
  })

  return {
    el,
    show: () => instance()?.show(),
    hide: () => instance()?.hide(),
  }
}
