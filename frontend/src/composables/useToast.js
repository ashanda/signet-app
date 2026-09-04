import Swal from 'sweetalert2'

// Ports the `Swal.mixin({ toast: true, position: 'top-end', showConfirmButton:
// false, timer: 1500, timerProgressBar: true, ... })` pattern that ui_spec.md
// documents as copy-pasted verbatim across ROC Income, Direct Share, Direct
// Share Log, Leadership Bonus Log, Users (leader-status toggle), and
// Leaders/Executives Gain (copy-to-clipboard) pages. One shared composable
// per ui_spec.md's explicit suggestion ("Worth extracting into one shared
// Vue composable/util (useToast())").
const toastMixin = Swal.mixin({
  toast: true,
  position: 'top-end',
  showConfirmButton: false,
  timer: 1500,
  timerProgressBar: true,
  didOpen: (toast) => {
    toast.addEventListener('mouseenter', Swal.stopTimer)
    toast.addEventListener('mouseleave', Swal.resumeTimer)
  },
})

export function useToast() {
  function toastSuccess(title) {
    return toastMixin.fire({ icon: 'success', title })
  }
  function toastError(title) {
    return toastMixin.fire({ icon: 'error', title })
  }
  return { toastSuccess, toastError, toastMixin }
}

// Standard (non-toast) success/error dialogs, used by the many pages that
// call plain `Swal.fire({icon:'success', title:'Updated!'})` /
// `Swal.fire({icon:'error', ...})` after a form submit.
export function useAlert() {
  function alertSuccess(title, text) {
    return Swal.fire({ icon: 'success', title, text })
  }
  function alertError(title, text) {
    return Swal.fire({ icon: 'error', title: title || 'Error', text })
  }
  function confirmDanger(title, text, confirmButtonText = 'Yes, delete it!') {
    return Swal.fire({
      title,
      text,
      icon: 'warning',
      showCancelButton: true,
      confirmButtonColor: '#d33',
      cancelButtonColor: '#6c757d',
      confirmButtonText,
    })
  }
  return { alertSuccess, alertError, confirmDanger }
}
