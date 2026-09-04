<script setup>
// Ports the centered white-card-over-bg-soft chrome shared by every
// public/auth-style page (login, password reset, register steps,
// buy-package, token/share) per ui_spec.md's repeated "same centered-card
// chrome as auth pages" notes.
//
// The original Blade pages use two different labels/targets for this link
// ("Back to Homepage" on login.blade.php vs "Back to log in" -> route('login')
// on password_reset/register_step1), and a couple of the later register
// steps don't render it at all — so it's exposed as props rather than
// hardcoded, defaulting to the login page's own copy/target.
defineProps({
  maxWidth: { type: String, default: '500px' },
  showBackLink: { type: Boolean, default: true },
  backText: { type: String, default: 'Back to Homepage' },
  backTo: { type: String, default: '/' },
})
</script>

<template>
  <main>
    <section class="vh-lg-100 mt-5 mt-lg-0 bg-soft d-flex align-items-center">
      <div class="container">
        <p v-if="showBackLink" class="text-center">
          <RouterLink :to="backTo" class="d-flex align-items-center justify-content-center">
            <i class="fas fa-arrow-left me-2"></i> {{ backText }}
          </RouterLink>
        </p>
        <div class="row justify-content-center form-bg-image auth-bg-image">
          <div class="col-12 d-flex align-items-center justify-content-center">
            <div class="bg-white shadow border-0 rounded border-light p-4 p-lg-5 w-100" :style="{ maxWidth }">
              <slot />
            </div>
          </div>
        </div>
      </div>
    </section>
  </main>
</template>

<style scoped>
/* Ports volt.js's data-background-lg behavior (login.blade.php's
   form-bg-image row): the illustration only appears as a background at the
   lg breakpoint (>991.98px), same as the original's clientWidth > breakpoints.lg
   check. */
@media (min-width: 992px) {
  .auth-bg-image {
    background-image: url('/resources/assets/img/illustrations/signin.svg');
  }
}
</style>
