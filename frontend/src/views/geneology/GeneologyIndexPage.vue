<script setup>
// Ports geneology/index.blade.php ("MY Genealogy" — route('my.geneology')).
// See geneology_handler.go's myGeneologyHandler: the signed-in user's direct
// downline (UserParent rows with node active/gratitude), one level deep.
//
// As in the original: children whose user is still pending are skipped,
// but still count toward each node's position; getDynamicColor() colours a
// node orange at positions 2, 5, 10, 15, 20, 30, 40, 50, 75, 100 (and
// cyclically past 100), blue otherwise, and yellow with black text when the
// user is deactive. The page's inline <style> — including its global
// `* { margin:0; padding:0; font-family: Arial }` reset — applies to the
// whole document while the page is open.
import { ref, onMounted } from 'vue'
import api from '@/api/client'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import { useAuthStore } from '@/store/auth'
import { usePageStyle } from '@/composables/usePageStyle'

const authStore = useAuthStore()
const childerns = ref([])
const loadError = ref('')

usePageStyle(`
    * { margin: 0; padding: 0; box-sizing: border-box; font-family: Arial, sans-serif; }
    .tree { display: flex; justify-content: center; margin-top: 50px; overflow-y: auto; max-height: 80vh; }
    .tree ul { padding-top: 20px; position: relative; transition: .5s; display: flex; justify-content: center; }
    .tree li { text-align: center; list-style-type: none; position: relative; padding: 20px 5px 0 5px; }
    .tree li::before, .tree li::after { content: ''; position: absolute; top: 0; right: 50%; border-top: 2px solid #ccc; width: 50%; height: 20px; }
    .tree li::after { right: auto; left: 50%; }
    .tree li:only-child::after, .tree li:only-child::before { display: none; }
    .tree li:only-child { padding-top: 0; }
    .tree li:first-child::before, .tree li:last-child::after { border: none; }
    .tree li:last-child::before { border-right: 2px solid #ccc; border-radius: 0 5px 0 0; }
    .tree li:first-child::after { border-left: 2px solid #ccc; border-radius: 5px 0 0 0; }
    .tree ul ul::before { content: ''; position: absolute; top: 0; left: 50%; border-left: 2px solid #ccc; width: 0; height: 20px; }
    .box { display: inline-block; border: 2px solid #3498db; padding: 10px 15px; background: #fff; border-radius: 8px; font-weight: bold; color: #3498db; box-shadow: 2px 2px 5px rgba(0, 0, 0, 0.1); text-decoration: none; }
    .box:hover { background: #3498db; color: #fff; transition: 0.3s; }
    @media (max-width: 768px) { .tree { max-height: 100vh; overflow-y: auto; } }
`)

const COLOR_MAP = { 2: '#FF5733', 5: '#FF5733', 10: '#FF5733', 15: '#FF5733', 20: '#FF5733', 30: '#FF5733', 40: '#FF5733', 50: '#FF5733', 75: '#FF5733', 100: '#FF5733' }
const COLOR_CYCLE = ['#FF5733', '#FF5733', '#FF5733', '#FF5733', '#FF5733']

function getDynamicColor(index) {
  if (COLOR_MAP[index]) return COLOR_MAP[index]
  if (index > 100) return COLOR_CYCLE[(index % 10) % COLOR_CYCLE.length]
  return '#3498db' // Default blue
}

function nodeStyle(childern, index) {
  const isDeactive = childern.user?.status === 'deactive'
  return {
    background: isDeactive ? '#f1c40f' : getDynamicColor(index + 1),
    color: isDeactive ? '#000' : '#fff',
  }
}

onMounted(async () => {
  try {
    const { data } = await api.get('/my-geneology')
    childerns.value = data.childerns || []
  } catch (err) {
    loadError.value = err?.response?.data?.message || 'Could not load geneology.'
  }
})
</script>

<template>
  <DashboardLayout>
    <div v-if="loadError" class="alert alert-danger">{{ loadError }}</div>
    <div class="py-4">
      <div class="d-flex justify-content-between w-100 flex-wrap">
        <div class="mb-3 mb-lg-0">
          <h1 class="h4">MY Genealogy</h1>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow mb-4">
      <div class="tree">
        <ul>
          <li>
            <div class="box" style="background: #3498db; color: white">
              {{ authStore.user?.name }}
            </div>
            <ul>
              <template v-for="(childern, index) in childerns" :key="childern.id">
                <li v-if="childern.user?.status !== 'pending'">
                  <RouterLink :to="{ name: 'geneology.show', params: { userId: childern.user.id } }" class="box" :style="nodeStyle(childern, index)">
                    {{ childern.user.name }}
                  </RouterLink>
                </li>
              </template>
            </ul>
          </li>
        </ul>
      </div>
    </div>
  </DashboardLayout>
</template>
