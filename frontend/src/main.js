import { createApp } from 'vue'
import { createPinia } from 'pinia'
// Volt (Themesberg's Bootstrap 5 admin theme) is the actual design system
// the original Blade app uses — it's a full standalone build (reboot, grid,
// components) so it replaces stock bootstrap.min.css rather than layering
// on top of it. custome.css is the original's own project-specific overrides.
import 'bootstrap/dist/js/bootstrap.bundle.min.js'
import './assets/theme/volt.css'
import './assets/theme/custome.css'
import 'sweetalert2/dist/sweetalert2.min.css'
import './assets/app.css'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './store/auth'

const app = createApp(App)
app.use(createPinia())
app.use(router)

// Exposed so api/client.js's 401 interceptor can force-logout without a
// circular import between the store and the api client.
const authStore = useAuthStore()
window.__signetAuthStore = authStore

// Resolve the session once before mounting so the router's auth guard has a
// definitive answer on first navigation instead of racing the /me request.
authStore.bootstrap().finally(() => {
  app.mount('#app')
})
