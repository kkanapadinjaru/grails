<script>
  import { connection, selectService, selectMethod } from '../stores/connection.svelte.js'

  let activeService = $derived(
    connection.services.find(s => s.displayName === connection.selectedService)
  )

  let methodFilter = $state('')

  // Reset the filter whenever the user switches services so the new list
  // shows in full.
  $effect(() => {
    connection.selectedService
    methodFilter = ''
  })

  let filteredMethods = $derived.by(() => {
    const q = methodFilter.trim().toLowerCase()
    if (!q) return connection.methods
    return connection.methods.filter(m => m.name.toLowerCase().includes(q))
  })
</script>

<aside class="w-72 bg-bg-light dark:bg-sidebar-dark border-r border-gray-200 dark:border-gray-700 flex flex-col">
  <div class="p-4 border-b border-gray-200 dark:border-gray-700">
    <h2 class="text-xs font-semibold text-text-light dark:text-text-dark mb-2">SERVICE</h2>
    <select
      value={connection.selectedService}
      onchange={(e) => selectService(e.currentTarget.value)}
      class="w-full px-3 py-1.5 text-sm bg-bg-light dark:bg-bg-dark border border-gray-300 dark:border-gray-600 rounded-md text-text-light dark:text-text-dark"
      disabled={!connection.connectedContext || connection.services.length === 0}
    >
      {#if connection.services.length === 0}
        <option value="">No services discovered</option>
      {/if}
      {#each connection.services as svc (svc.displayName)}
        <option value={svc.displayName} title={`${svc.namespace}/${svc.k8sService} • ${svc.viaNodePort ? 'NodePort' : 'port-forward'}`}>{svc.displayName}</option>
      {/each}
    </select>
    {#if activeService}
      <div class="mt-2 text-xs text-gray-500 dark:text-gray-400 truncate" title={activeService.serviceName}>
        {activeService.serviceName}
      </div>
      <div class="mt-1 text-xs text-gray-500 dark:text-gray-400 truncate">
        {activeService.namespace}/{activeService.k8sService} · {activeService.viaNodePort ? 'NodePort' : 'port-forward'}
      </div>
    {/if}
  </div>

  <div class="flex-1 p-4 overflow-y-auto flex flex-col min-h-0">
    <div class="flex items-center justify-between mb-2">
      <h2 class="text-xs font-semibold text-text-light dark:text-text-dark">METHODS</h2>
      {#if methodFilter && filteredMethods.length !== connection.methods.length}
        <span class="text-xs text-gray-500 dark:text-gray-400">{filteredMethods.length}/{connection.methods.length}</span>
      {/if}
    </div>
    {#if connection.methods.length > 0}
      <div class="relative mb-2">
        <input
          type="text"
          bind:value={methodFilter}
          placeholder="Filter methods..."
          class="w-full pl-7 pr-7 py-1.5 text-xs bg-bg-light dark:bg-bg-dark border border-gray-300 dark:border-gray-600 rounded-md text-text-light dark:text-text-dark"
        />
        <svg class="absolute left-2 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-4.35-4.35M11 19a8 8 0 100-16 8 8 0 000 16z" />
        </svg>
        {#if methodFilter}
          <button
            type="button"
            onclick={() => (methodFilter = '')}
            class="absolute right-1.5 top-1/2 -translate-y-1/2 p-0.5 text-gray-400 hover:text-gray-700 dark:hover:text-gray-200"
            title="Clear filter"
          >
            <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        {/if}
      </div>
    {/if}
    <div class="space-y-1 overflow-y-auto flex-1 min-h-0">
      {#if filteredMethods.length === 0 && methodFilter}
        <div class="text-xs text-gray-500 dark:text-gray-400 py-2 text-center">No methods match "{methodFilter}"</div>
      {/if}
      {#each filteredMethods as method (method.name)}
        <div
          onclick={() => selectMethod(method.name)}
          role="button"
          tabindex="0"
          class="w-full text-left px-4 py-2 text-sm cursor-pointer transition-colors {connection.selectedMethod === method.name ? 'bg-hi-light dark:bg-hi-dark text-btn-light dark:text-btn-dark border-l-4 border-btn-light dark:border-btn-dark' : 'text-text-light dark:text-text-dark hover:bg-hi-light dark:hover:bg-hi-dark'}"
        >
          <div class="text-sm font-medium">{method.name}</div>
          {#if method.requestType || method.responseType}
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {method.requestType ?? ''} → {method.responseType ?? ''}
            </div>
          {/if}
        </div>
      {/each}
    </div>
  </div>
</aside>
