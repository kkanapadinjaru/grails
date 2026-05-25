<script>
  import { history, replayHistoryEntry, clearHistory } from '../stores/connection.svelte.js'
  import { ui } from '../stores/settings.svelte.js'
</script>

<div class="transition-all duration-300 {ui.isHistoryPanelOpen ? 'w-80' : 'w-8'} bg-bg-light dark:bg-bg-dark border-l border-gray-200 dark:border-gray-700 flex shrink-0">
  <button
    onclick={() => ui.isHistoryPanelOpen = !ui.isHistoryPanelOpen}
    class="w-8 h-full flex items-center justify-center hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
    title={ui.isHistoryPanelOpen ? 'Collapse' : 'Expand history'}
  >
    <svg class="w-4 h-4 text-gray-400 transition-transform {ui.isHistoryPanelOpen ? 'rotate-180' : ''}" fill="none" stroke="currentColor" viewBox="0 0 24 24">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
    </svg>
  </button>

  {#if ui.isHistoryPanelOpen}
    <div class="flex-1 p-4 overflow-y-auto flex flex-col min-h-0">
      <div class="flex items-center justify-between mb-3">
        <h2 class="text-sm font-semibold text-gray-900 dark:text-white">Request History</h2>
        {#if history.entries.length > 0}
          <button
            onclick={clearHistory}
            class="text-xs text-gray-500 hover:text-red-600 dark:text-gray-400 dark:hover:text-red-400"
            title="Clear all history"
          >
            Clear
          </button>
        {/if}
      </div>
      <div class="space-y-1 flex-1 overflow-y-auto min-h-0">
        {#each history.entries as entry (entry.id)}
          <div
            onclick={() => replayHistoryEntry(entry)}
            role="button"
            tabindex="0"
            class="p-2 rounded cursor-pointer border-l-2 {entry.isError ? 'border-red-500 bg-red-50 dark:bg-red-900/10 hover:bg-red-100 dark:hover:bg-red-900/20' : 'border-green-500 bg-sidebar-light dark:bg-sidebar-dark hover:bg-gray-100 dark:hover:bg-hi-dark'}"
            title={`${entry.serviceName || ''}.${entry.method} → ${entry.localAddress || ''}`}
          >
            <div class="flex items-center justify-between">
              <div class="flex-1 min-w-0">
                <div class="text-xs font-medium text-gray-900 dark:text-white truncate">{entry.method}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400 truncate">{entry.service}</div>
              </div>
              <div class="text-xs text-gray-400 dark:text-gray-500 ml-2 flex-shrink-0 text-right">
                <div>{entry.time}</div>
                <div>{entry.durationMs}ms</div>
              </div>
            </div>
            <div class="mt-1 text-xs {entry.isError ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'} font-medium truncate">{entry.status}</div>
          </div>
        {/each}
        {#if history.entries.length === 0}
          <div class="text-xs text-gray-500 dark:text-gray-400">No history yet</div>
        {/if}
      </div>
    </div>
  {/if}
</div>
