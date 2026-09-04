<script setup>
// Ports layouts/sidebar.blade.php (ui_spec.md "Layout shells" section).
// Role-gated nav item visibility reproduced 1:1 from the documented
// `@if(Auth::user()->role == '...')` blocks, including the documented dead
// branch (My Activations' inner admin check never fires because it's nested
// inside an outer role=='company' check — we just render the company link).
import { computed } from 'vue'
import { useAuthStore } from '@/store/auth'
import { useNewActivationsCount } from '@/composables/useNewActivationsCount'

const auth = useAuthStore()
const role = computed(() => auth.role)
const { count: waitingCount } = useNewActivationsCount()
</script>

<template>
  <nav class="navbar navbar-dark navbar-theme-primary px-4 col-12 d-lg-none">
    <RouterLink class="navbar-brand me-lg-5" :to="auth.dashboardRoute">
      <img class="navbar-brand-dark" src="/resources/assets/img/brand/logo.png" alt="Volt logo" />
      <img class="navbar-brand-light" src="/resources/assets/img/brand/logo.png" alt="Volt logo" />
    </RouterLink>
    <div class="d-flex align-items-center">
      <button
        class="navbar-toggler d-lg-none collapsed"
        type="button"
        data-bs-toggle="collapse"
        data-bs-target="#sidebarMenu"
        aria-controls="sidebarMenu"
        aria-expanded="false"
        aria-label="Toggle navigation"
      >
        <span class="navbar-toggler-icon"></span>
      </button>
    </div>
  </nav>

  <nav id="sidebarMenu" class="sidebar d-lg-block bg-gray-800 text-white collapse">
    <div class="sidebar-inner px-4 pt-3">
      <ul class="nav flex-column pt-3 pt-md-0">
        <li class="nav-item">
          <RouterLink class="navbar-brand me-lg-5" :to="auth.dashboardRoute">
            <img class="navbar-brand-dark mb-3" src="/resources/assets/img/brand/logo.png" alt="Volt logo" style="height: 65px !important" />
          </RouterLink>
        </li>
        <li class="nav-item">
          <RouterLink :to="auth.dashboardRoute" class="nav-link" active-class="active">
            <span class="sidebar-icon"><i class="fas fa-home me-2"></i></span>
            <span class="sidebar-text">Dashboard</span>
          </RouterLink>
        </li>

        <template v-if="role === 'company'">
          <li class="nav-item"><RouterLink :to="`/admin/${auth.user?.id}/setup-google-auth`" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-shield-alt me-2"></i></span><span class="sidebar-text">Google Auth Setup</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/countries" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-globe me-2"></i></span><span class="sidebar-text">Countries</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/packages" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-box me-2"></i></span><span class="sidebar-text">Packages</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/mining/users" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-coins me-2"></i></span><span class="sidebar-text">Mining Token</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/users" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-users me-2"></i></span><span class="sidebar-text">Users</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/roc" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-chart-line me-2"></i></span><span class="sidebar-text">ROC Income</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/direct-share" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-hand-holding-usd me-2"></i></span><span class="sidebar-text">Global Director Share</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/direct-share-log" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-list me-2"></i></span><span class="sidebar-text">GDS Log</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/salaries" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-money-check-alt me-2"></i></span><span class="sidebar-text">Salaries</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/leaders/gain" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-crown me-2"></i></span><span class="sidebar-text">Leaders Gain</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/executives/gain" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-user-tie me-2"></i></span><span class="sidebar-text">Executives Gain</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/leader-code-logs" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-history me-2"></i></span><span class="sidebar-text">Leader Code Logs</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/leadership-bonus-log" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-file-invoice-dollar me-2"></i></span><span class="sidebar-text">LB Log</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/user-parent-logs" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-user-slash me-2"></i></span><span class="sidebar-text">Fake Accounts</span></RouterLink></li>
        </template>

        <template v-if="role === 'admin'">
          <li class="nav-item"><RouterLink to="/users" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-users me-2"></i></span><span class="sidebar-text">Users</span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/leader-code-logs" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-history me-2"></i></span><span class="sidebar-text">Leader Code Logs</span></RouterLink></li>
        </template>

        <template v-if="['admin', 'user', 'agent'].includes(role)">
          <li class="nav-item"><RouterLink to="/token-shares" class="nav-link d-flex justify-content-between" active-class="active"><span><span class="sidebar-icon"><i class="fas fa-exchange-alt me-2"></i></span><span class="sidebar-text">Token Share</span></span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/token/share/logs" class="nav-link d-flex justify-content-between" active-class="active"><span><span class="sidebar-icon"><i class="fas fa-receipt me-2"></i></span><span class="sidebar-text">Token Share Log</span></span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/buy-package-history" class="nav-link d-flex justify-content-between" active-class="active"><span><span class="sidebar-icon"><i class="fas fa-arrow-up me-2"></i></span><span class="sidebar-text">Top Up</span></span></RouterLink></li>
          <li class="nav-item"><RouterLink to="/earn/history" class="nav-link d-flex justify-content-between" active-class="active"><span><span class="sidebar-icon"><i class="fas fa-wallet me-2"></i></span><span class="sidebar-text">Earn Log</span></span></RouterLink></li>
        </template>

        <template v-if="role === 'company'">
          <li class="nav-item"><RouterLink to="/company/pending-activation" class="nav-link d-flex justify-content-between" active-class="active"><span><span class="sidebar-icon"><i class="fas fa-clock me-2"></i></span><span class="sidebar-text">My Activations</span></span></RouterLink></li>
          <li class="nav-item">
            <RouterLink to="/new-activations" class="nav-link d-flex justify-content-between align-items-center" active-class="active">
              <span><span class="sidebar-icon"><i class="fas fa-hourglass-half me-2"></i></span><span class="sidebar-text">Waiting Activations</span></span>
              <span class="badge bg-danger rounded-pill">{{ waitingCount || 0 }}</span>
            </RouterLink>
          </li>
        </template>

        <li v-if="role !== 'company'" class="nav-item">
          <RouterLink to="/my-geneology" class="nav-link" active-class="active"><span class="sidebar-icon"><i class="fas fa-sitemap me-2"></i></span><span class="sidebar-text">Geneology</span></RouterLink>
        </li>

        <li class="nav-item">
          <RouterLink :to="role === 'company' ? '/kyc' : '/kyc/show'" class="nav-link d-flex justify-content-between" active-class="active">
            <span><span class="sidebar-icon"><i class="fas fa-id-card me-2"></i></span><span class="sidebar-text">KYC</span></span>
          </RouterLink>
        </li>

        <li role="separator" class="dropdown-divider mt-4 mb-3 border-gray-700"></li>
        <li class="nav-item">
          <a href="javascript:void(0)" target="_blank" class="nav-link d-flex align-items-center">
            <span class="sidebar-icon"><i class="fas fa-info-circle me-2"></i></span>
            <span class="sidebar-text">Support</span>
          </a>
        </li>
        <li class="nav-item">
          <a href="javascript:void(0)" class="nav-link d-flex align-items-center" @click="auth.logout()">
            <span class="sidebar-icon"><i class="fas fa-sign-out-alt me-2"></i></span>
            <span class="sidebar-text">Logout</span>
          </a>
        </li>
      </ul>
    </div>
  </nav>
</template>
