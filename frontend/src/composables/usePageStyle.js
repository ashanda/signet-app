import { onMounted, onBeforeUnmount } from 'vue'

// A Blade view's inline <style> block: global CSS that applies to the whole
// document only while that page is showing.
export function usePageStyle(css) {
  let el = null
  onMounted(() => {
    el = document.createElement('style')
    el.textContent = css
    document.head.appendChild(el)
  })
  onBeforeUnmount(() => {
    if (el) el.remove()
    el = null
  })
}
