<script setup>
// Shared body for Admin/Agent/User dashboards — ui_spec.md's Admin/Agent
// role sections state these are "byte-for-byte the same component set"
// (admin/dashboard.blade.php + agent/dashboard.blade.php + User's
// equivalent), so this one component is reused by the three page files
// instead of copy-pasting ~200 lines of markup three times. Company's
// dashboard is structurally different (no mining widget, token/pool
// summary cards instead) and stays its own page (CompanyDashboardPage.vue).
//
// Response shape comes from dashboard_handler.go's admin/agent/user
// handlers (all three return the same envelope): {ref_link, my_tokens,
// my_wallet, total_value, my_package, fee_percentage, pool_amount,
// total_poolshare_value, my_share_value, my_globle_director_share,
// activations:{..paginated..}, rank, roc, all_users, new_activations}.
// Admin's response additionally carries `token_counts` (a paginated
// per-user active/deactive token breakdown) — ui_spec.md's documented page
// structure for admin/dashboard.blade.php does not mention a table for it
// (the original computes it but the view never renders it), so it's
// fetched-but-unused here too, matching the source.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import { useAuthStore } from '@/store/auth'
import { useDashboardMetaStore } from '@/store/dashboardMeta'
import { useApiAction } from '@/composables/useApiAction'
import { useToast } from '@/composables/useToast'
import { useActivePackage } from '@/composables/useActivePackage'
import RankWidget from '@/components/widgets/RankWidget.vue'
import RocWidget from '@/components/widgets/RocWidget.vue'
import MiningWidget from '@/components/widgets/MiningWidget.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'

const props = defineProps({
  roleLabel: { type: String, required: true }, // "Admin" | "Agent" | "User"
  endpoint: { type: String, required: true }, // '/admin/dashboard' | '/agent/dashboard' | '/user/dashboard'
  activateEndpoint: { type: String, default: '/active-package' },
})

// admin/dashboard.blade.php and agent/dashboard.blade.php show "Welcome,
// {Role} Head" (capitalized) with a vacation toggle and an extra duplicated
// "My Wallet" stat card (a copy-paste artifact in the original, admin only);
// user/dashboard.blade.php shows plain "Welcome, {role}" with no vacation
// toggle and no duplicate card. Derived from roleLabel rather than a
// separate prop since the two always agree.
const role = props.roleLabel.toLowerCase()
const showHeadSuffix = role !== 'user'
const showVacationToggle = role !== 'user'
const showDuplicateWalletCard = role === 'admin'

const authStore = useAuthStore()
const dashboardMeta = useDashboardMetaStore()
const { run } = useApiAction()
const { toastSuccess, toastError } = useToast()
const { activate } = useActivePackage(props.activateEndpoint)

const resp = ref(null)
const loading = ref(true)
const loadError = ref('')
const onVacation = ref(false)
const copied = ref(false)

async function fetchDashboard(page = 1) {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get(props.endpoint, { params: { page } })
    resp.value = data
    dashboardMeta.setNewActivations(data.new_activations)
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load dashboard data.'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  onVacation.value = !!authStore.user?.on_vacation
  fetchDashboard()
})

async function toggleVacation() {
  const { ok, data } = await run(() => api.post('/toggle-vacation'), {
    showSuccessAlert: false,
    showErrorAlert: false,
  })
  if (ok) {
    onVacation.value = !!data.on_vacation
    if (authStore.user) authStore.user.on_vacation = onVacation.value
    toastSuccess(`Vacation turned ${onVacation.value ? 'ON' : 'OFF'}`)
  } else {
    toastError('Could not update vacation status')
  }
}

async function copyLink() {
  if (!resp.value?.ref_link) return
  try {
    await navigator.clipboard.writeText(resp.value.ref_link)
    copied.value = true
    toastSuccess('Referral link copied!')
    setTimeout(() => (copied.value = false), 1500)
  } catch {
    toastError('Could not copy link')
  }
}

function needTokens(row) {
  const price = Number(row.package_price || 0)
  const fee = Number(resp.value?.fee_percentage || 0)
  return (price - (price * fee) / 100).toFixed(2)
}

async function onActivate(id) {
  const ok = await activate(id)
  if (ok) fetchDashboard(resp.value?.activations?.current_page || 1)
}
</script>

<template>
  <div>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div v-if="resp" class="card bg-yellow-100 border-0 shadow">
      <div class="card-header d-sm-flex flex-row align-items-center flex-0" :class="{ 'justify-content-between': showVacationToggle }">
        <div class="d-block mb-3 mb-sm-0">
          <h2 class="fs-3 fw-extrabold">
            <template v-if="showHeadSuffix"><span class="text-capitalize">{{ roleLabel }}</span> Head</template>
            <template v-else>{{ roleLabel }}</template>
          </h2>
          <h3 v-if="resp.my_package" class="fs-3 fw-extrabold">
            <span class="text-capitalize">{{ resp.my_package.rank }} SIG-00{{ authStore.user?.id }}</span>
          </h3>
          <RankWidget :rank="resp.rank" />
          <RocWidget :roc="resp.roc" :roc-status="authStore.user?.roc_status" />
        </div>

        <div v-if="showVacationToggle" class="form-check form-switch">
          <input
            id="vacationSwitch"
            class="form-check-input"
            type="checkbox"
            role="switch"
            :checked="onVacation"
            @change="toggleVacation"
          />
          <label class="form-check-label" for="vacationSwitch">
            Vacation {{ onVacation ? 'ON' : 'OFF' }}
          </label>
        </div>
      </div>
      <div class="card-body p-2">
        <div class="container ref">
          <label>Your Referral Link</label>
          <input type="text" id="urlInput" class="form-control" placeholder="Enter a URL here" :value="resp.ref_link" readonly />
          <button type="button" id="copyButton" class="btn btn-primary d-inline-flex align-items-center" @click="copyLink">
            Copy URL
            <svg class="icon icon-xxs ms-2" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M2 9.5A3.5 3.5 0 005.5 13H9v2.586l-1.293-1.293a1 1 0 00-1.414 1.414l3 3a1 1 0 001.414 0l3-3a1 1 0 00-1.414-1.414L11 15.586V13h2.5a4.5 4.5 0 10-.616-8.958 4.002 4.002 0 10-7.753 1.977A3.5 3.5 0 002 9.5zm9 3.5H9V8a1 1 0 012 0v5z" clip-rule="evenodd"></path></svg>
          </button>
          <div id="urlDisplay" class="url-display">{{ copied ? 'Copied!' : '' }}</div>
        </div>
      </div>
    </div>

    <div v-if="resp" class="row mt-4">
      <div class="col-12 col-sm-6 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-primary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path d="M13 6a3 3 0 11-6 0 3 3 0 016 0zM18 8a2 2 0 11-4 0 2 2 0 014 0zM14 15a4 4 0 00-8 0v3h8v-3zM6 8a2 2 0 11-4 0 2 2 0 014 0zM16 18v-3a5.972 5.972 0 00-.75-2.906A3.005 3.005 0 0119 15v3h-3zM4.75 12.094A5.973 5.973 0 004 15v3H1v-3a3 3 0 013.75-2.906z"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="h5">My Token</h2>
                  <h3 class="fw-extrabold mb-1">{{ resp.my_tokens }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">My Token</h2>
                  <h3 class="fw-extrabold mb-2">{{ resp.my_tokens }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-sm-6 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">My Wallet</h2>
                  <h3 class="mb-1">{{ (resp.total_value - resp.wallet_balance).toFixed(2) }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">My Wallet</h2>
                  <h3 class="fw-extrabold mb-2">{{ (resp.total_value - resp.wallet_balance).toFixed(2) }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- admin/dashboard.blade.php duplicates this exact "My Wallet" card
           a second time (col-sm-12, no USDT suffix) — a copy-paste artifact
           agent/user's dashboards don't have. Reproduced verbatim, admin only. -->
      <div v-if="showDuplicateWalletCard" class="col-12 col-sm-12 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">My Wallet</h2>
                  <h3 class="mb-1">{{ (resp.total_value - resp.wallet_balance).toFixed(2) }}</h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">My Wallet</h2>
                  <h3 class="fw-extrabold mb-2">{{ (resp.total_value - resp.wallet_balance).toFixed(2) }}</h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="resp.my_globle_director_share" class="col-12 col-sm-12 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">My Global Director Share Wallet</h2>
                  <h3 class="mb-1">{{ resp.my_globle_director_share.balance ?? 0 }}</h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">My Global Director Share Wallet</h2>
                  <h3 class="fw-extrabold mb-2">{{ resp.my_globle_director_share.balance ?? 0 }}</h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-sm-12 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">My Earn</h2>
                  <h3 class="mb-1">{{ resp.wallet_balance ?? 0 }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">My Earn</h2>
                  <h3 class="fw-extrabold mb-2">{{ resp.wallet_balance ?? 0 }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-sm-12 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">This Time Direct Share Pool value</h2>
                  <h3 class="mb-1">{{ resp.pool_amount ?? 0 }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">This Time Direct Share Pool value</h2>
                  <h3 class="fw-extrabold mb-2">{{ resp.pool_amount ?? 0 }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="col-12 col-sm-12 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">This Time All Share Value</h2>
                  <h3 class="mb-1">{{ resp.total_poolshare_value ?? 0 }}</h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">This Time All Share Value</h2>
                  <h3 class="fw-extrabold mb-2">{{ resp.total_poolshare_value ?? 0 }}</h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="authStore.user?.global_director_share_status == 1" class="col-12 col-sm-12 col-xl-4 mb-4">
        <div class="card border-0 shadow">
          <div class="card-body">
            <div class="row d-block d-xl-flex align-items-center">
              <div class="col-12 col-xl-5 text-xl-center mb-3 mb-xl-0 d-flex align-items-center justify-content-xl-center">
                <div class="icon-shape icon-shape-secondary rounded me-4 me-sm-0">
                  <svg class="icon" fill="currentColor" viewBox="0 0 20 20" xmlns="http://www.w3.org/2000/svg"><path fill-rule="evenodd" d="M10 2a4 4 0 00-4 4v1H5a1 1 0 00-.994.89l-1 9A1 1 0 004 18h12a1 1 0 00.994-1.11l-1-9A1 1 0 0015 7h-1V6a4 4 0 00-4-4zm2 5V6a2 2 0 10-4 0v1h4zm-6 3a1 1 0 112 0 1 1 0 01-2 0zm7-1a1 1 0 100 2 1 1 0 000-2z" clip-rule="evenodd"></path></svg>
                </div>
                <div class="d-sm-none">
                  <h2 class="fw-extrabold h5">This Time My Share Portion</h2>
                  <h3 class="mb-1">{{ Math.trunc(resp.my_share_value ?? 0) }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
              <div class="col-12 col-xl-7 px-xl-0">
                <div class="d-none d-sm-block">
                  <h2 class="h6 text-gray-400 mb-0">This Time My Share Portion</h2>
                  <h3 class="fw-extrabold mb-2">{{ Math.trunc(resp.my_share_value ?? 0) }}<span class="text-success fw-bold fs-5"> USDT</span></h3>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <MiningWidget v-if="authStore.user" :user-id="authStore.user.id" :all-users-count="resp?.all_users || 0" />

    <div class="row">
      <div class="col-12 col-xl-12">
        <div class="row">
          <div class="col-12 mb-4">
            <div class="card border-0 shadow">
              <div class="card-header">
                <div class="row align-items-center">
                  <div class="col">
                    <h2 class="fs-5 fw-bold mb-0">Activations</h2>
                  </div>
                  <div class="col text-end">
                    <a href="#" class="btn btn-sm btn-primary">See all</a>
                  </div>
                </div>
              </div>
              <div class="table-responsive">
                <table class="table align-items-center table-flush">
                  <thead class="thead-light">
                    <tr>
                      <th class="border-bottom" scope="col">User name</th>
                      <th class="border-bottom" scope="col">Whats app</th>
                      <th class="border-bottom" scope="col">Need Tokens</th>
                      <th class="border-bottom" scope="col">Action</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-if="!resp?.activations?.data?.length">
                      <td colspan="4" class="text-center text-muted py-4">No pending activations found</td>
                    </tr>
                    <tr v-for="row in resp?.activations?.data" :key="row.id">
                      <td>{{ row.user_name || 'Unknown User' }}</td>
                      <td>{{ row.user_whatsapp }}</td>
                      <td><span class="badge bg-success">{{ needTokens(row) }} USDT</span></td>
                      <td>
                        <span v-if="row.company_status == 0" class="badge bg-success p-2">Wait for company activation</span>
                        <button v-else type="button" class="btn btn-primary active-package" @click="onActivate(row.id)">
                          Active
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <div class="d-flex justify-content-center">
                  <Paginator :pagination="resp?.activations" @change="(p) => fetchDashboard(p)" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
