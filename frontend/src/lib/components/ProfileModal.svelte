<script>
  import { settings, ui, saveSettings } from '../stores/settings.svelte.js'

  let namespacesText = $state('')
  let grpcPortsText = $state('')
  let excludePatternsText = $state('')

  // Re-sync the editable text fields whenever the modal opens so they reflect the
  // latest persisted values.
  $effect(() => {
    if (ui.showProfile) {
      namespacesText = settings.namespaces.join(', ')
      grpcPortsText = settings.grpcPorts.join(', ')
      excludePatternsText = (settings.serviceExcludePatterns || []).join(', ')
    }
  })

  function close() {
    ui.showProfile = false
  }

  async function save() {
    settings.namespaces = namespacesText.split(',').map(s => s.trim()).filter(Boolean)
    settings.grpcPorts = grpcPortsText.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !Number.isNaN(n))
    settings.serviceExcludePatterns = excludePatternsText.split(',').map(s => s.trim()).filter(Boolean)
    await saveSettings()
    close()
  }
</script>

{#if ui.showProfile}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-bg-light dark:bg-sidebar-dark rounded-lg p-6 w-[28rem] shadow-xl max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
      <h2 class="text-lg font-semibold text-text-light dark:text-text-dark mb-4">Profile / Settings</h2>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Namespaces (comma-separated)</label>
          <input
            type="text"
            bind:value={namespacesText}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Port range start</label>
            <input
              type="number"
              bind:value={settings.portRangeStart}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Port range end</label>
            <input
              type="number"
              bind:value={settings.portRangeEnd}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
            />
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">gRPC ports to scan (comma-separated)</label>
          <input
            type="text"
            bind:value={grpcPortsText}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Exclude service patterns (comma-separated globs)</label>
          <input
            type="text"
            bind:value={excludePatternsText}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
            placeholder="*wassups, *-internal"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Skip k8s services whose name matches any pattern (case-insensitive).</p>
        </div>

        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">NodePort host</label>
          <input
            type="text"
            bind:value={settings.nodePortHost}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Token endpoint</label>
          <input
            type="text"
            bind:value={settings.tokenEndpoint}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
            placeholder="https://auth.example/realms/.../protocol/openid-connect/token"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">OIDC client ID</label>
          <input
            type="text"
            bind:value={settings.clientId}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
          />
        </div>

        <div class="flex space-x-3 pt-2">
          <button
            onclick={close}
            class="flex-1 px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-gray-700 dark:text-gray-300 text-sm hover:bg-gray-100 dark:hover:bg-gray-700"
          >
            Cancel
          </button>
          <button
            onclick={save}
            class="flex-1 px-4 py-2 bg-blue-500 hover:bg-blue-600 text-white text-sm rounded-md"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
