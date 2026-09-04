<script setup>
// Ports layouts/topbar.blade.php: slim dark top navbar, just the user's
// name and a dropdown with a single Logout item. The sidebar's own
// d-lg-none mobile bar (see Sidebar.vue) carries the only navbar-toggler in
// the original; this component intentionally has none.
//
// UX deviation from the original (user-requested): the original's toggle
// was bare text with no visible affordance that it opens a menu — a real
// user found it hard to notice/click. Kept the same dropdown + single
// Logout action, but added a user-icon avatar, a chevron, and a
// hover/pressed background so the click target is obvious and larger.
import { useAuthStore } from '@/store/auth'

const auth = useAuthStore()
</script>

<template>
  <nav class="navbar navbar-top navbar-expand navbar-dashboard navbar-dark ps-0 pe-2 pb-0">
    <div class="container-fluid px-0">
      <div class="d-flex justify-content-between w-100">
        <div class="d-flex align-items-center"></div>
        <ul class="navbar-nav align-items-center">
          <li class="nav-item dropdown ms-lg-3">
            <a
              class="nav-link dropdown-toggle account-toggle d-flex align-items-center px-3 py-2 rounded-3"
              href="#"
              role="button"
              data-bs-toggle="dropdown"
              aria-expanded="false"
            >
              <span class="account-avatar rounded-circle bg-primary bg-opacity-10 text-primary d-flex align-items-center justify-content-center me-2">
                <i class="fas fa-user"></i>
              </span>
              <span class="fw-bold text-gray-900 d-none d-lg-inline">{{ auth.user?.name }}</span>
            </a>
            <div class="dropdown-menu dashboard-dropdown dropdown-menu-end mt-2 py-1">
              <a class="dropdown-item d-flex align-items-center py-2" href="javascript:void(0)" @click="auth.logout()">
                <i class="fas fa-sign-out-alt text-danger me-2"></i>
                Logout
              </a>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </nav>
</template>

<style scoped>
.account-toggle {
  cursor: pointer;
  transition: background-color 0.15s ease-in-out;
}
.account-toggle:hover,
.account-toggle:focus,
.account-toggle[aria-expanded='true'] {
  background-color: rgba(0, 0, 0, 0.05);
}
.account-avatar {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
  font-size: 0.9rem;
}
</style>
