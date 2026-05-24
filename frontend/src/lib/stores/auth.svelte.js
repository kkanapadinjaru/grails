import {
  Login as LoginCall,
  Logout as LogoutCall,
  RefreshToken as RefreshCall,
  GetAuthState,
} from '../../../wailsjs/go/main/App.js'
import { EventsOn } from '../../../wailsjs/runtime/runtime.js'
import { addInfo, addError, addWarn } from './logs.svelte.js'

export const auth = $state({
  isLoggedIn: false,
  username: '',
  bearerToken: '',
  expiresAt: 0,            // unix seconds
  refreshExpiresAt: 0,
  showLoginModal: false,
  showUserMenu: false,
  isGeneratingToken: false,
  isRefreshingToken: false,
  loginError: '',
})

function applyState(state) {
  if (!state) {
    auth.isLoggedIn = false
    auth.username = ''
    auth.bearerToken = ''
    auth.expiresAt = 0
    auth.refreshExpiresAt = 0
    return
  }
  auth.isLoggedIn = !!state.loggedIn
  auth.username = state.username || ''
  auth.bearerToken = state.accessToken || ''
  auth.expiresAt = state.expiresAt || 0
  auth.refreshExpiresAt = state.refreshExpiresAt || 0
}

export function openLogin() {
  auth.loginError = ''
  auth.showLoginModal = true
}

export function closeLogin() {
  auth.showLoginModal = false
  auth.loginError = ''
}

export async function login(username, password) {
  auth.isGeneratingToken = true
  auth.loginError = ''
  try {
    const state = await LoginCall(username, password)
    applyState(state)
    addInfo(`Logged in as ${auth.username}`)
    return true
  } catch (err) {
    const msg = String(err)
    auth.loginError = msg
    addError(`Login failed: ${msg}`)
    return false
  } finally {
    auth.isGeneratingToken = false
  }
}

export async function refreshToken() {
  if (!auth.isLoggedIn) return
  auth.isRefreshingToken = true
  try {
    const state = await RefreshCall()
    applyState(state)
    addInfo('Token refreshed')
  } catch (err) {
    addError(`Refresh failed: ${err}`)
  } finally {
    auth.isRefreshingToken = false
  }
}

export async function logout() {
  try {
    await LogoutCall()
  } catch (err) {
    addWarn(`Logout warning: ${err}`)
  }
  applyState(null)
  addInfo('Logged out')
}

// Subscribe to backend-emitted events. Wails' EventsOn returns an unsubscribe
// function but we want these listeners for the lifetime of the app.
export function initAuthEvents() {
  EventsOn('token:refreshed', state => {
    applyState(state)
    addInfo(`Token refreshed (expires in ~${remainingSec(state)}s)`)
  })
  EventsOn('token:expired', state => {
    applyState(null)
    addWarn('Session expired — please log in again')
    auth.showLoginModal = true
  })
  EventsOn('token:cleared', () => {
    applyState(null)
  })
}

export async function rehydrateAuth() {
  try {
    const state = await GetAuthState()
    applyState(state)
  } catch (err) {
    addWarn(`Failed to rehydrate auth: ${err}`)
  }
}

function remainingSec(state) {
  if (!state || !state.expiresAt) return '?'
  const now = Math.floor(Date.now() / 1000)
  return Math.max(0, state.expiresAt - now)
}
