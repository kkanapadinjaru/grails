<script>
  import { logs, clearLogs } from '../stores/logs.svelte.js'
  import { connection } from '../stores/connection.svelte.js'

  let isExpanded = $state(true)
  let filter = $state('all')

  let filtered = $derived(
    filter === 'all' ? logs.entries : logs.entries.filter(e => e.level === filter)
  )

  let hasError = $derived(logs.entries.some(e => e.level === 'error'))

  function levelColor(level) {
    switch (level) {
      case 'error': return 'text-red-400'
      case 'warning': return 'text-yellow-400'
      case 'info': return 'text-blue-400'
      case 'debug': return 'text-gray-400'
      default: return 'text-gray-400'
    }
  }

  function fmt(d) {
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  }
</script>

<div class="border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shrink-0">
  <div
    onclick={() => isExpanded = !isExpanded}
    role="button"
    tabindex="0"
    class="flex items-center justify-between px-6 py-2 cursor-pointer transition-colors select-none {hasError ? 'bg-red-50 dark:bg-red-900/20' : 'bg-white dark:bg-gray-800'} hover:bg-gray-50 dark:hover:bg-gray-700/50"
    style="min-height: 36px;"
  >
    <div class="flex items-center justify-between w-full text-xs text-gray-900 dark:text-white">
      <div class="flex items-center space-x-4">
        <div class="flex items-center space-x-2">
          <div class="w-2 h-2 rounded-full {connection.connectedContext ? 'bg-green-500' : 'bg-gray-400'}"></div>
          <span>{connection.connectedContext ? `Connected to ${connection.connectedContext}` : 'Disconnected'}</span>
        </div>
        {#if hasError}
          <span class="text-red-500 font-medium">{logs.entries.filter(l => l.level === 'error').length} Errors</span>
        {/if}
      </div>
      <div>
        <span>{isExpanded ? '▼' : '▶'} Logs ({logs.entries.length})</span>
      </div>
      <div>
        <span>Grails v1.0.0</span>
      </div>
    </div>
  </div>

  {#if isExpanded}
    <div class="border-t border-gray-200 dark:border-gray-700">
      <div class="flex items-center justify-between p-2 bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-700">
        <div class="flex items-center space-x-2">
          <label class="text-xs text-gray-900 dark:text-white">Filter:</label>
          <select
            bind:value={filter}
            class="text-xs px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="all">All</option>
            <option value="error">Errors</option>
            <option value="warning">Warnings</option>
            <option value="info">Info</option>
            <option value="debug">Debug</option>
          </select>
          <span class="text-xs text-gray-500 dark:text-gray-400">
            Showing {filtered.length} of {logs.entries.length}
          </span>
        </div>
        <button
          onclick={clearLogs}
          class="text-xs text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          Clear
        </button>
      </div>

      <div class="p-3 bg-white dark:bg-gray-900 max-h-48 overflow-y-auto font-mono text-xs">
        {#if filtered.length === 0}
          <div class="text-center text-gray-500 dark:text-gray-400 py-4">No logs</div>
        {:else}
          {#each filtered as entry (entry.id)}
            <div class="{levelColor(entry.level)} mb-1">
              <span class="text-gray-500 dark:text-gray-400">[{fmt(entry.timestamp)}]</span>
              <span class="ml-2 text-gray-900 dark:text-white">{entry.message}</span>
            </div>
          {/each}
        {/if}
      </div>
    </div>
  {/if}
</div>
