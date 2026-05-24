<script>
  import { auth, closeLogin, login } from '../stores/auth.svelte.js'
  import { settings } from '../stores/settings.svelte.js'

  let username = $state('')
  let password = $state('')

  let endpointConfigured = $derived(!!settings.tokenEndpoint && !!settings.clientId)

  async function submit() {
    if (!username || !password) return
    const ok = await login(username, password)
    if (ok) {
      username = ''
      password = ''
      closeLogin()
    }
  }
</script>

{#if auth.showLoginModal}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-bg-light dark:bg-sidebar-dark rounded-lg p-6 w-96 shadow-xl" onclick={(e) => e.stopPropagation()}>
      <h2 class="text-lg font-semibold text-text-light dark:text-text-dark mb-4">Login</h2>
      {#if !endpointConfigured}
        <div class="mb-4 p-3 rounded bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-300 dark:border-yellow-700 text-xs text-yellow-800 dark:text-yellow-200">
          Token endpoint and OIDC client ID are not configured. Open the Settings cog to set them.
        </div>
      {/if}
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Username</label>
          <input
            type="text"
            bind:value={username}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
            placeholder="Enter username"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Password</label>
          <input
            type="password"
            bind:value={password}
            onkeydown={(e) => { if (e.key === 'Enter') submit() }}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
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
            disabled={auth.isGeneratingToken || !username || !password || !endpointConfigured}
            class="flex-1 px-4 py-2 bg-blue-500 hover:bg-blue-600 disabled:bg-gray-400 disabled:cursor-not-allowed text-white text-sm rounded-md"
          >
            {auth.isGeneratingToken ? 'Logging in...' : 'Login'}
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
