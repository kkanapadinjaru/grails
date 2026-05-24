<script>
  import { theme, toggleTheme } from '../stores/theme.svelte.js'
  import { connection, connect, selectNamespace } from '../stores/connection.svelte.js'
  import { auth, openLogin, logout } from '../stores/auth.svelte.js'
  import { ui } from '../stores/settings.svelte.js'

  function onConnect() {
    connect()
  }

  let isConnectedToSelected = $derived(
    connection.connectedContext !== '' && connection.connectedContext === connection.selectedCluster
  )
</script>

<header class="h-12 bg-bg-light dark:bg-sidebar-dark border-b border-gray-200 dark:border-gray-700 flex items-center justify-between px-6 shrink-0">
  <div class="flex items-center space-x-6">
    <div class="flex items-center space-x-2">
      <div class="relative w-10 h-6 flex items-center justify-center">
        <svg class="absolute inset-0 w-10 h-6" viewBox="0 0 40 24">
          <path d="M12 4 L12 8 Q12 16 20 16 Q28 16 28 8 L28 4" fill="none" stroke="#3B82F6" stroke-width="2" opacity="0.9"/>
          <line x1="20" y1="16" x2="20" y2="20" stroke="#3B82F6" stroke-width="2" opacity="0.7"/>
          <line x1="16" y1="20" x2="24" y2="20" stroke="#3B82F6" stroke-width="2" opacity="0.7"/>
          <path d="M14 6 Q14 6 20 6 Q26 6 26 10 Q26 12 22 12" fill="none" stroke="#3B82F6" stroke-width="1.5" opacity="0.6"/>
          <circle cx="18" cy="8" r="0.8" fill="#F59E0B" class="animate-pulse"/>
          <circle cx="20" cy="10" r="0.6" fill="#F59E0B" class="animate-pulse" style="animation-delay: 0.1s"/>
          <circle cx="22" cy="12" r="0.8" fill="#F59E0B" class="animate-pulse" style="animation-delay: 0.2s"/>
          <circle cx="20" cy="14" r="0.6" fill="#F59E0B" class="animate-pulse" style="animation-delay: 0.3s"/>
        </svg>
      </div>
      <span class="text-lg font-bold text-text-light dark:text-text-dark">Grails</span>
    </div>

    <div class="flex items-center space-x-3">
      <div class="flex items-center space-x-2">
        <label class="text-xs font-medium text-text-light dark:text-text-dark">ENVIRONMENT</label>
        <div class="w-56">
          <select
            bind:value={connection.selectedCluster}
            class="w-full px-3 py-1.5 text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 dark:text-gray-100"
          >
            {#each connection.clusters as c (c.context)}
              <option value={c.context}>{c.name}</option>
            {/each}
          </select>
        </div>
      </div>

      {#if connection.namespaces.length > 0 && connection.connectedContext}
        <div class="flex items-center space-x-2">
          <label class="text-xs font-medium text-text-light dark:text-text-dark">NAMESPACE</label>
          <div class="w-44">
            <select
              value={connection.selectedNamespace}
              onchange={(e) => selectNamespace(e.currentTarget.value)}
              class="w-full px-3 py-1.5 text-sm bg-white dark:bg-gray-800 border border-gray-300 dark:border-gray-600 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 text-gray-900 dark:text-gray-100"
            >
              {#each connection.namespaces as n (n)}
                <option value={n}>{n}</option>
              {/each}
            </select>
          </div>
        </div>
      {/if}

      <button
        onclick={onConnect}
        disabled={connection.isConnecting || isConnectedToSelected}
        class="flex items-center space-x-2 px-4 py-1.5 rounded-md text-sm font-medium transition-colors {isConnectedToSelected ? 'bg-green-500 text-white cursor-default' : 'bg-btn-light dark:bg-btn-dark hover:bg-blue-600 text-white'} {connection.isConnecting ? 'opacity-50 cursor-not-allowed' : ''}"
      >
        {#if connection.isConnecting}
          <svg class="w-4 h-4 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          <span>Connecting...</span>
        {:else if isConnectedToSelected}
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span>Connected</span>
        {:else}
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7m9 11l-1-1-4-4-4 4-1 1V4z" />
          </svg>
          <span>Connect</span>
        {/if}
      </button>
    </div>
  </div>

  <div class="flex items-center space-x-2">
    <button
      onclick={() => ui.showProfile = true}
      class="p-2 text-text-light hover:text-text-light dark:text-text-dark dark:hover:text-white"
      title="Settings"
    >
      <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317a1 1 0 011.35 0l.852.852a1 1 0 001.35 0l.852-.852a1 1 0 011.35 0l.852.852M19 13a4 4 0 11-8 0 4 4 0 018 0zM5 17a4 4 0 100-8 4 4 0 000 8z" />
      </svg>
    </button>
    <button
      onclick={toggleTheme}
      class="p-2 text-text-light hover:text-text-light dark:text-text-dark dark:hover:text-white"
      title="Toggle theme"
    >
      {#if theme.value === 'dark'}
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z" />
        </svg>
      {:else}
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z" />
        </svg>
      {/if}
    </button>

    <div class="relative">
      {#if auth.isLoggedIn}
        <button
          onclick={(e) => { e.stopPropagation(); auth.showUserMenu = !auth.showUserMenu }}
          class="flex items-center space-x-2 p-2 text-text-light hover:text-text-light dark:text-text-dark dark:hover:text-white"
          title="User menu"
        >
          <div class="w-6 h-6 bg-green-500 rounded-full flex items-center justify-center">
            <svg class="w-3 h-3 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
            </svg>
          </div>
          <span class="text-sm text-text-light dark:text-text-dark">{auth.username}</span>
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </button>

        {#if auth.showUserMenu}
          <div class="absolute right-0 mt-2 w-48 bg-bg-light dark:bg-sidebar-dark rounded-lg shadow-lg border border-gray-200 dark:border-gray-700 py-1 z-50" onclick={(e) => e.stopPropagation()}>
            <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700">
              <div class="text-sm font-medium text-text-light dark:text-text-dark">{auth.username}</div>
              <div class="text-xs text-gray-500 dark:text-gray-400">Logged in</div>
            </div>
            <button
              onclick={() => { ui.showProfile = true; auth.showUserMenu = false }}
              class="w-full text-left px-4 py-2 text-sm text-text-light dark:text-text-dark hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317a1 1 0 011.35 0l.852.852a1 1 0 001.35 0l.852-.852a1 1 0 011.35 0l.852.852M19 13a4 4 0 11-8 0 4 4 0 018 0zM5 17a4 4 0 100-8 4 4 0 000 8z" />
              </svg>
              <span>Profile</span>
            </button>
            <button
              onclick={() => { logout(); auth.showUserMenu = false }}
              class="w-full text-left px-4 py-2 text-sm text-text-light dark:text-text-dark hover:bg-gray-100 dark:hover:bg-gray-700 flex items-center space-x-2"
            >
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
              </svg>
              <span>Logout</span>
            </button>
          </div>
        {/if}
      {:else}
        <button
          onclick={openLogin}
          class="p-2 text-text-light hover:text-text-light dark:text-text-dark dark:hover:text-white"
          title="Login"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
          </svg>
        </button>
      {/if}
    </div>
  </div>
</header>
