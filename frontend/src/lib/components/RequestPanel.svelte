<script>
  import { connection, generateSampleRequest, cacheCurrentBody, sendRequest as doSend, authOverrides, toggleAuthRequired } from '../stores/connection.svelte.js'
  import { auth, openLogin } from '../stores/auth.svelte.js'
  import { settings } from '../stores/settings.svelte.js'

  let activeService = $derived(
    connection.services.find(s => s.displayName === connection.selectedService)
  )

  let authRequired = $derived((settings.authEndpoints || []).length > 0 && !!settings.clientId)

  let authNeeded = $derived(authRequired && (!(connection.selectedService in authOverrides) || authOverrides[connection.selectedService]))

  let canSend = $derived(
    !!connection.selectedService && !!connection.selectedMethod && !connection.isSending
  )

  function onBodyInput() {
    cacheCurrentBody()
  }

  async function sendRequest() {
    if (authNeeded && !auth.isLoggedIn) {
      openLogin()
      return
    }
    await doSend()
  }
</script>

<div class="flex-1 bg-sidebar-light dark:bg-sidebar-dark overflow-y-auto">
  <div class="p-6 border-b border-gray-200 dark:border-gray-700">
    <h2 class="text-lg font-semibold text-text-light dark:text-text-dark">
      {activeService?.serviceName || 'No service selected'}
      {#if connection.selectedMethod}
        <span class="text-text-light dark:text-text-dark"> · {connection.selectedMethod}</span>
      {/if}
    </h2>
    {#if connection.requestType}
      <p class="text-sm text-gray-600 dark:text-gray-400 mt-1 truncate" title={`${connection.requestType} → ${connection.responseType}`}>
        {connection.requestType} → {connection.responseType}
      </p>
    {/if}
  </div>

  <div class="p-6">
    <h3 class="text-sm font-semibold text-text-light dark:text-text-dark mb-3">Request</h3>

    <div class="space-y-3">
      <div>
        <div class="flex items-center justify-between mb-1">
          <label class="text-xs font-medium text-text-light dark:text-text-dark">Bearer Token</label>
          {#if authRequired}
            <button
              type="button"
              onclick={() => toggleAuthRequired(connection.selectedService)}
              class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400"
              title={authNeeded ? 'Auth required — click to skip for this service' : 'Auth skipped — click to require for this service'}
            >
              <span>{authNeeded ? 'Required' : 'Skipped'}</span>
              <span class="relative inline-flex h-4 w-7 items-center rounded-full transition-colors {authNeeded ? 'bg-btn-light dark:bg-btn-dark' : 'bg-gray-300 dark:bg-gray-600'}">
                <span class="inline-block h-3 w-3 rounded-full bg-white transition-transform {authNeeded ? 'translate-x-3.5' : 'translate-x-0.5'}"></span>
              </span>
            </button>
          {/if}
        </div>
        <div class="relative">
          <input
            type="password"
            value={authNeeded ? (auth.bearerToken || '') : ''}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm pr-8"
            placeholder={!authNeeded ? 'Auth skipped for this service' : auth.isLoggedIn ? 'Token generated' : 'Login to generate token...'}
            readonly
          />
          {#if authNeeded && auth.isLoggedIn}
            <div class="absolute inset-y-0 right-0 flex items-center px-2">
              <div class="w-2 h-2 bg-green-500 rounded-full"></div>
            </div>
          {/if}
        </div>
      </div>

      <div>
        <div class="flex items-center justify-between mb-1">
          <label class="text-xs font-medium text-text-light dark:text-text-dark">Request Body (JSON)</label>
          <button
            onclick={generateSampleRequest}
            class="p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            title="Fill with random sample values"
          >
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
            </svg>
          </button>
        </div>
        <textarea
          bind:value={connection.requestBody}
          oninput={onBodyInput}
          class="w-full h-64 px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm font-mono"
          placeholder={'{"user_id": "123"}'}
        ></textarea>
      </div>

      <button
        onclick={sendRequest}
        disabled={!canSend}
        class="px-4 py-2 bg-btn-light dark:bg-btn-dark hover:bg-blue-600 text-white rounded-md text-sm font-medium transition-colors disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center"
      >
        {#if connection.isSending}
          <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8v4a4 4 0 00-4 4H4z"></path>
          </svg>
          Sending...
        {:else}
          Send Request
        {/if}
      </button>
    </div>
  </div>
</div>
