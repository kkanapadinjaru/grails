<script>
  import { connection } from '../stores/connection.svelte.js'

  let statusClass = $derived(
    !connection.responseStatus
      ? 'text-gray-500 dark:text-gray-400'
      : connection.isResponseError
        ? 'text-red-600 dark:text-red-400'
        : 'text-green-600 dark:text-green-400'
  )

  function copyResponse() {
    if (connection.response) navigator.clipboard.writeText(connection.response)
  }
</script>

<div class="flex-1 bg-bg-light dark:bg-bg-dark flex flex-col min-h-0">
  <div class="p-6 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between shrink-0">
    <h3 class="text-sm font-semibold text-text-light dark:text-text-dark">Response</h3>
    <button
      onclick={copyResponse}
      class="p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
      title="Copy response"
    >
      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
      </svg>
    </button>
  </div>

  <div class="p-6 flex-1 flex flex-col min-h-0">
    <div class="flex flex-col h-full space-y-3">
      <div class="flex items-center space-x-4 mb-2 shrink-0">
        <div class="flex items-center space-x-2">
          <span class="text-xs text-gray-500">Status:</span>
          <span class="text-xs font-medium {statusClass}">{connection.responseStatus || '—'}</span>
        </div>
        <div class="flex items-center space-x-2">
          <span class="text-xs text-gray-500">Time:</span>
          <span class="text-xs font-medium text-text-light dark:text-text-dark">{connection.responseTime || '—'}</span>
        </div>
      </div>
      <div class="flex-1 p-3 bg-sidebar-light dark:bg-sidebar-dark rounded-md overflow-auto">
        <pre class="text-sm text-text-light dark:text-text-dark font-mono whitespace-pre-wrap">{connection.response || 'No response yet'}</pre>
      </div>
    </div>
  </div>
</div>
