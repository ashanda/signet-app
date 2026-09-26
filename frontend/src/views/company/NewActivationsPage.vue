<script setup>
// Ports company/new-activations.blade.php (route('new.activations')).
// Lists user_packages with company_status = 0 (see dashboard_handler.go's
// companyNewActivationsHandler). The "Active" button posts to
// /company/new-active-package (NOT /active-package — that's Pending
// Activations' endpoint), per useActivePackage's documented parameter.
//
// dashActivationRow's user_name/user_whatsapp/package_name may be null
// (LEFT JOINs); nullStr() (utils/format.js) renders those as "".
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import Paginator from '@/components/shared/Paginator.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useActivePackage } from '@/composables/useActivePackage'
import { nullStr } from '@/utils/format'

const { activate } = useActivePackage('/company/new-active-package')

const activations = ref(null)
const loadError = ref('')

async function fetchList(page = 1) {
  loadError.value = ''
  try {
    const { data } = await api.get('/company/new-activations', { params: { page } })
    activations.value = data.activations
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load activations.'
  }
}

onMounted(() => fetchList())
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">New Packages/TopUp Activations</h1>
        </div>
      </div>
    </div>
    <div class="card border-0 shadow mb-4">
      <div class="card-body">
        <div class="table-responsive">
          <table class="table align-items-center table-flush">
            <thead class="thead-light">
              <tr>
                <th class="border-bottom" scope="col">User name</th>
                <th class="border-bottom" scope="col">Whats app</th>
                <th class="border-bottom" scope="col">Package</th>
                <th class="border-bottom" scope="col">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="row in activations?.data" :key="row.id">
                <td>{{ nullStr(row.user_name) || 'Unknown User' }}</td>
                <td>{{ nullStr(row.user_whatsapp) }}</td>
                <td>{{ nullStr(row.package_name) || 'Unknown Package' }}</td>
                <td>
                  <button class="btn btn-primary active-package" @click="activate(row.id)">Active</button>
                </td>
              </tr>
            </tbody>
          </table>
          <div class="d-flex justify-content-center">
            <Paginator :pagination="activations" @change="(p) => fetchList(p)" />
          </div>
        </div>
      </div>
    </div>
  </DashboardLayout>
</template>
