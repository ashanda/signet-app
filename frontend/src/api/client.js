import axios from 'axios'

// Single axios instance for the whole app. withCredentials so the Go
// backend's httpOnly `signet_session` cookie is sent/received (SameSite=Lax
// session auth — see backend/internal/auth/auth.go). Base path matches
// every RegisterXRoutes route prefix (`/api/v1/...`).
const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 401 anywhere (except the auth/me probe itself and the login call) means the
// session expired or was never established — bounce to /login. The auth
// store's bootstrap() call intentionally swallows its own 401 so this global
// handler doesn't cause a redirect loop on first load for a logged-out user.
api.interceptors.response.use(
  (res) => res,
  (err) => {
    const url = err?.config?.url || ''
    const isAuthProbe = url.includes('/me') || url.includes('/login')
    if (err?.response?.status === 401 && !isAuthProbe) {
      const authStore = window.__signetAuthStore
      if (authStore) authStore.forceLogout()
    }
    return Promise.reject(err)
  }
)

export default api
