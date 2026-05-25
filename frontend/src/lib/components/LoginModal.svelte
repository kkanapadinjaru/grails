<script>
  import { auth, closeLogin, login } from '../stores/auth.svelte.js'
  import { settings, ui } from '../stores/settings.svelte.js'
  import { connection } from '../stores/connection.svelte.js'
  import { GetActiveAuthEndpoint, ResolveRealm } from '../../../wailsjs/go/main/App.js'

  let subdomain = $state('')
  let username = $state('')
  let password = $state('')
  let endpoint = $state(null)
  let realmPreview = $state('')
  let realmError = $state('')
  let resolvingRealm = $state(false)

  // localStorage key for remembering the last subdomain per (cluster, namespace).
  let memoryKey = $derived(`grails:subdomain:${connection.selectedCluster || ''}:${connection.selectedNamespace || ''}`)

  $effect(() => {
    if (auth.showLoginModal) {
      loadEndpoint()
      const remembered = localStorage.getItem(memoryKey) || ''
      subdomain = remembered
      realmPreview = ''
      realmError = ''
    }
  })

  async function loadEndpoint() {
    try {
      endpoint = await GetActiveAuthEndpoint()
    } catch (err) {
      endpoint = null
    }
  }

  async function previewRealm() {
    if (!endpoint?.needsSubdomain || !subdomain) {
      realmPreview = ''
      realmError = ''
      return
    }
    resolvingRealm = true
    realmError = ''
    try {
      realmPreview = await ResolveRealm(subdomain)
    } catch (err) {
      realmPreview = ''
      realmError = String(err)
    } finally {
      resolvingRealm = false
    }
  }

  async function submit() {
    if (!username || !password) return
    if (endpoint?.needsSubdomain && !subdomain) return
    if (subdomain) localStorage.setItem(memoryKey, subdomain)
    const ok = await login(subdomain, username, password)
    if (ok) {
      username = ''
      password = ''
      closeLogin()
    }
  }

  function openSettings() {
    closeLogin()
    ui.showProfile = true
  }

  let canSubmit = $derived(
    !!endpoint?.found &&
    !!username &&
    !!password &&
    (!endpoint.needsSubdomain || !!subdomain)
  )
</script>

{#if auth.showLoginModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-bg-light dark:bg-sidebar-dark rounded-lg p-6 w-96 shadow-xl" onclick={(e) => e.stopPropagation()}>
      <h2 class="text-lg font-semibold text-text-light dark:text-text-dark mb-4">Login</h2>

      {#if endpoint}
        <div class="mb-4 px-3 py-2 rounded bg-gray-100 dark:bg-gray-800 text-xs text-gray-700 dark:text-gray-300">
          {endpoint.cluster || '—'} <span class="text-gray-400">·</span> {endpoint.namespace || '—'}
        </div>
      {/if}

      {#if endpoint && !endpoint.found}
        <div class="mb-4 p-3 rounded bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-300 dark:border-yellow-700 text-xs text-yellow-800 dark:text-yellow-200">
          ⚠ No auth endpoint configured for this cluster / namespace.
        </div>
        <div class="flex space-x-3">
          <button
            onclick={closeLogin}
            class="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            Close
          </button>
          <button
            onclick={openSettings}
            class="flex-1 px-4 py-2 bg-btn-light dark:bg-btn-dark hover:opacity-90 text-white text-sm rounded-md"
          >
            Open Settings
          </button>
        </div>
      {:else}
        <div class="space-y-4">
          {#if endpoint?.needsSubdomain}
            <div>
              <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Subdomain</label>
              <input
                type="text"
                bind:value={subdomain}
                onblur={previewRealm}
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                placeholder="acme"
              />
              {#if resolvingRealm}
                <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">→ resolving…</p>
              {:else if realmPreview}
                <p class="mt-1 text-xs text-green-600 dark:text-green-400">→ realm: {realmPreview} ✓</p>
              {:else if realmError}
                <p class="mt-1 text-xs text-red-600 dark:text-red-400 break-words">→ {realmError}</p>
              {/if}
            </div>
          {/if}

          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Username</label>
            <input
              type="text"
              bind:value={username}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
              placeholder="Enter username"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Password</label>
            <input
              type="password"
              bind:value={password}
              onkeydown={(e) => { if (e.key === 'Enter') submit() }}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
              placeholder="Enter password"
            />
          </div>
          {#if auth.loginError}
            <div class="text-xs text-red-600 dark:text-red-400 break-words">{auth.loginError}</div>
          {/if}
          <div class="flex space-x-3">
            <button
              onclick={closeLogin}
              class="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              Cancel
            </button>
            <button
              onclick={submit}
              disabled={auth.isGeneratingToken || !canSubmit}
              class="flex-1 px-4 py-2 bg-btn-light dark:bg-btn-dark hover:opacity-90 disabled:bg-gray-400 disabled:cursor-not-allowed text-white text-sm rounded-md"
            >
              {auth.isGeneratingToken ? 'Signing in…' : 'Sign In'}
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}
