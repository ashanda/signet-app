<script setup>
// Ports geneology/index.blade.php ("MY Genealogy" — route('my.geneology'),
// non-company roles). See geneology_handler.go's myGeneologyHandler (GET
// /my-geneology).
//
// The Go response (`childerns`) is already exactly ONE level: `UserParent::
// where('parent_id', auth()->id())` — i.e. this is root (you) + your direct
// downline only, not a deep recursive tree; matches ui_spec.md's described
// markup (root box + a row of child link-boxes). Pending-node children are
// already excluded server-side (`node IN ('active','gratitude')`), so no
// client-side `@continue`-equivalent filter is needed.
//
// Color rule (ui_spec.md "Genealogy tree node" table): root = blue
// `#3498db` white text; child with `user.status === 'deactive'` = yellow
// `#f1c40f` black text; everyone else = orange `#FF5733` (the "dynamic
// color" helper always resolves to this one constant per ui_spec.md's
// explicit note).
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import FlashAlert from '@/components/shared/FlashAlert.vue'
import { useAuthStore } from '@/store/auth'

const authStore = useAuthStore()

const loading = ref(true)
const loadError = ref('')
const children = ref([])

function nodeColor(child) {
  if (child.user?.status === 'deactive') return { background: '#f1c40f', color: '#000' }
  return { background: '#FF5733', color: '#fff' }
}

async function fetchData() {
  loading.value = true
  loadError.value = ''
  try {
    const { data } = await api.get('/my-geneology')
    children.value = data.childerns || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load geneology.'
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <DashboardLayout>
    <FlashAlert type="danger" :message="loadError" @close="loadError = ''" />

    <div class="py-4">
      <h1 class="h4">MY Genealogy</h1>
    </div>

    <div class="card border-0 shadow">
      <div class="geneology-tree">
        <ul>
          <li>
            <div class="geneology-box" style="background: #3498db; color: white;">
              {{ authStore.user?.name }}
            </div>
            <ul v-if="children.length">
              <li v-for="child in children" :key="child.id">
                <router-link
                  :to="{ name: 'geneology.show', params: { userId: child.user_id } }"
                  class="geneology-box"
                  :style="nodeColor(child)"
                >
                  {{ child.user?.name }}
                </router-link>
              </li>
            </ul>
          </li>
        </ul>
        <p v-if="!loading && !children.length" class="text-muted text-center py-4">No downline members yet.</p>
      </div>
    </div>
  </DashboardLayout>
</template>

<style scoped>
/* Ported verbatim from geneology/index.blade.php's inline <style> block —
   a classic CSS-only org-chart connector-line tree (per-branch ::before/
   ::after lines meeting at each parent), not a generic divider. */
.geneology-tree {
  display: flex;
  justify-content: center;
  margin-top: 50px;
  overflow-y: auto;
  max-height: 80vh;
}

.geneology-tree ul {
  padding-top: 20px;
  position: relative;
  transition: 0.5s;
  display: flex;
  justify-content: center;
}

.geneology-tree li {
  text-align: center;
  list-style-type: none;
  position: relative;
  padding: 20px 5px 0 5px;
}

.geneology-tree li::before,
.geneology-tree li::after {
  content: '';
  position: absolute;
  top: 0;
  right: 50%;
  border-top: 2px solid #ccc;
  width: 50%;
  height: 20px;
}

.geneology-tree li::after {
  right: auto;
  left: 50%;
}

.geneology-tree li:only-child::after,
.geneology-tree li:only-child::before {
  display: none;
}

.geneology-tree li:only-child {
  padding-top: 0;
}

.geneology-tree li:first-child::before,
.geneology-tree li:last-child::after {
  border: none;
}

.geneology-tree li:last-child::before {
  border-right: 2px solid #ccc;
  border-radius: 0 5px 0 0;
}

.geneology-tree li:first-child::after {
  border-left: 2px solid #ccc;
  border-radius: 5px 0 0 0;
}

.geneology-tree ul ul::before {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  border-left: 2px solid #ccc;
  width: 0;
  height: 20px;
}

.geneology-box {
  display: inline-block;
  border: 2px solid #3498db;
  padding: 10px 15px;
  background: #fff;
  border-radius: 8px;
  font-weight: bold;
  color: #3498db;
  box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.1);
  text-decoration: none;
}

.geneology-box:hover {
  background: #3498db;
  color: #fff;
  transition: 0.3s;
}

@media (max-width: 768px) {
  .geneology-tree {
    max-height: 100vh;
    overflow-y: auto;
  }
}
</style>
