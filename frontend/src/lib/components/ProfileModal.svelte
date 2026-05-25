<script>
  import { settings, ui, saveSettings } from '../stores/settings.svelte.js'
  import { connection } from '../stores/connection.svelte.js'
  import { addError } from '../stores/logs.svelte.js'
  import { theme, setPalette, PALETTES } from '../stores/theme.svelte.js'
  import { OpenLogsFolder, GetLogsFolder } from '../../../wailsjs/go/main/App.js'

  let logsFolder = $state('')

  $effect(() => {
    if (ui.showProfile && !logsFolder) {
      GetLogsFolder().then(p => { logsFolder = p }).catch(() => {})
    }
  })

  async function openLogs() {
    try {
      await OpenLogsFolder()
    } catch (err) {
      addError(`Failed to open logs folder: ${err}`)
    }
  }

  let namespacesText = $state('')
  let grpcPortsText = $state('')
  let excludePatternsText = $state('')
  let parentClaimMapText = $state('')
  let endpoints = $state([])
  let themeExpanded = $state(false)
  let diagnosticsExpanded = $state(false)
  let expandedIdx = $state(-1)
  let claimMapExpanded = $state(false)

  // Re-sync the editable text fields whenever the modal opens so they reflect the
  // latest persisted values.
  $effect(() => {
    if (ui.showProfile) {
      namespacesText = settings.namespaces.join(', ')
      grpcPortsText = settings.grpcPorts.join(', ')
      excludePatternsText = (settings.serviceExcludePatterns || []).join(', ')
      parentClaimMapText = Object.entries(settings.parentClaimMap || {})
        .map(([k, v]) => `${k}=${v}`)
        .join(', ')
      endpoints = (settings.authEndpoints || []).map(e => ({ ...e }))
      expandedIdx = -1
      themeExpanded = false
      diagnosticsExpanded = false
      claimMapExpanded = false
    }
  })

  function addEndpoint() {
    const used = new Set(endpoints.map(e => e.cluster))
    const firstFree = (connection.clusters || []).find(c => !used.has(c.context))
    const defaultCluster = firstFree?.context || connection.clusters[0]?.context || ''
    endpoints.push({
      cluster: defaultCluster,
      namespace: '*',
      tokenUrl: '',
      realmResolverUrl: '',
      realmJsonPath: 'realm',
    })
    expandedIdx = endpoints.length - 1
  }

  function removeEndpoint(i) {
    endpoints.splice(i, 1)
    if (expandedIdx === i) expandedIdx = -1
    else if (expandedIdx > i) expandedIdx -= 1
  }

  function endpointLabel(e) {
    const ctx = e.cluster || ''
    const match = (connection.clusters || []).find(c => c.context === ctx)
    const c = match?.name || ctx || '(unset)'
    const n = e.namespace || '*'
    return `${c} · ${n}`
  }

  function parseClaimMap(text) {
    const out = {}
    for (const entry of text.split(',')) {
      const [k, v] = entry.split('=').map(s => (s || '').trim())
      if (k && v) out[k] = v
    }
    return out
  }

  function close() {
    ui.showProfile = false
  }

  async function save() {
    settings.namespaces = namespacesText.split(',').map(s => s.trim()).filter(Boolean)
    settings.grpcPorts = grpcPortsText.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !Number.isNaN(n))
    settings.serviceExcludePatterns = excludePatternsText.split(',').map(s => s.trim()).filter(Boolean)
    settings.parentClaimMap = parseClaimMap(parentClaimMapText)
    settings.authEndpoints = endpoints.map(e => ({
      cluster: (e.cluster || '*').trim() || '*',
      namespace: (e.namespace || '*').trim() || '*',
      tokenUrl: (e.tokenUrl || '').trim(),
      realmResolverUrl: (e.realmResolverUrl || '').trim(),
      realmJsonPath: (e.realmJsonPath || '').trim(),
    }))
    await saveSettings()
    close()
  }
</script>

{#if ui.showProfile}
  <div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
    <div class="bg-bg-light dark:bg-sidebar-dark rounded-lg p-6 w-[48rem] shadow-xl max-h-[90vh] overflow-y-auto" onclick={(e) => e.stopPropagation()}>
      <h2 class="text-lg font-semibold text-text-light dark:text-text-dark mb-4">Settings</h2>
      <div class="space-y-4">
        <div>
          <button
            type="button"
            onclick={() => themeExpanded = !themeExpanded}
            class="flex items-center justify-between w-full text-sm font-medium text-text-light dark:text-text-dark mb-2"
          >
            <span>Theme palette · {PALETTES.find(p => p.id === theme.palette)?.name || theme.palette}</span>
            <span class="text-xs">{themeExpanded ? '▾' : '▸'}</span>
          </button>
          {#if themeExpanded}
          <div class="grid grid-cols-2 gap-2">
            {#each PALETTES as p (p.id)}
              <button
                type="button"
                onclick={() => setPalette(p.id)}
                class="flex items-center gap-2 px-3 py-2 border rounded-md text-left text-sm transition-colors
                       {theme.palette === p.id
                         ? 'border-btn-light dark:border-btn-dark bg-hi-light dark:bg-hi-dark'
                         : 'border-gray-300 dark:border-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700'}"
              >
                <span class="flex shrink-0 gap-0.5">
                  <span class="w-3 h-6 rounded-sm" style="background: {p.swatch[0]}"></span>
                  <span class="w-3 h-6 rounded-sm" style="background: {p.swatch[1]}"></span>
                  <span class="w-3 h-6 rounded-sm" style="background: {p.swatch[2]}"></span>
                </span>
                <div class="min-w-0">
                  <div class="font-medium text-text-light dark:text-text-dark truncate">{p.name}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400 truncate">{p.blurb}</div>
                </div>
              </button>
            {/each}
          </div>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Each palette has a light and dark variant — toggle with the sun/moon button in the header.</p>
          {/if}
        </div>

        <div class="border-t border-gray-300 dark:border-gray-600 pt-4">
          <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Namespaces (comma-separated)</label>
          <input
            type="text"
            bind:value={namespacesText}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
          />
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Port range start</label>
            <input
              type="number"
              bind:value={settings.portRangeStart}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Port range end</label>
            <input
              type="number"
              bind:value={settings.portRangeEnd}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
            />
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">gRPC ports to scan (comma-separated)</label>
            <input
              type="text"
              bind:value={grpcPortsText}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Discovery concurrency</label>
            <input
              type="number"
              min="1"
              max="32"
              bind:value={settings.discoveryConcurrency}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Parallel port-forwards per namespace scan (default 5).</p>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Exclude service patterns (comma-separated globs)</label>
            <input
              type="text"
              bind:value={excludePatternsText}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
              placeholder="*wassups, *-lb"
            />
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Skip k8s services whose name, app label, or selector value matches (case-insensitive).</p>
          </div>
          <div>
            <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">NodePort host</label>
            <input
              type="text"
              bind:value={settings.nodePortHost}
              class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
            />
          </div>
        </div>

        <div class="border-t border-gray-300 dark:border-gray-600 pt-4">
          <h3 class="text-sm font-semibold text-text-light dark:text-text-dark mb-3">Authentication</h3>

          <div class="space-y-3">
            <div class="grid grid-cols-2 gap-3">
              <div>
                <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Provider</label>
                <select
                  bind:value={settings.authProvider}
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                >
                  <option value="keycloak">Keycloak</option>
                  <option value="auth0" disabled>Auth0 (coming soon)</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Client ID</label>
                <input
                  type="text"
                  bind:value={settings.clientId}
                  class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-text-light dark:text-text-dark mb-1">Endpoints</label>
              <p class="mb-2 text-xs text-gray-500 dark:text-gray-400">
                One entry per (cluster, namespace). Use <code>*</code> as a wildcard. First exact-or-wildcard match wins.
              </p>

              <div class="space-y-2">
                {#each endpoints as ep, i (i)}
                  <div class="border border-gray-300 dark:border-gray-600 rounded-md">
                    <div class="flex items-center justify-between px-3 py-2">
                      <button
                        type="button"
                        onclick={() => (expandedIdx = expandedIdx === i ? -1 : i)}
                        class="flex-1 text-left text-sm text-text-light dark:text-text-dark"
                      >
                        <span class="mr-1">{expandedIdx === i ? '▾' : '▸'}</span>
                        {endpointLabel(ep)}
                      </button>
                      <button
                        type="button"
                        onclick={() => removeEndpoint(i)}
                        class="ml-2 text-xs text-red-500 hover:text-red-700"
                        title="Remove"
                      >
                        ×
                      </button>
                    </div>
                    {#if expandedIdx === i}
                      <div class="px-3 pb-3 space-y-2 border-t border-gray-200 dark:border-gray-700">
                        <div class="grid grid-cols-2 gap-2 pt-2">
                          <div>
                            <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Cluster</label>
                            <select
                              bind:value={ep.cluster}
                              class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                            >
                              {#if ep.cluster && !(connection.clusters || []).some(c => c.context === ep.cluster)}
                                <option value={ep.cluster}>{ep.cluster} (not in kubeconfig)</option>
                              {/if}
                              {#each connection.clusters || [] as c (c.context)}
                                <option value={c.context}>{c.name}</option>
                              {/each}
                              {#if (connection.clusters || []).length === 0}
                                <option value="" disabled>No clusters loaded</option>
                              {/if}
                            </select>
                          </div>
                          <div>
                            <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Namespace</label>
                            <input
                              type="text"
                              bind:value={ep.namespace}
                              class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                              placeholder="* or namespace"
                            />
                          </div>
                        </div>
                        <div>
                          <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Token URL</label>
                          <input
                            type="text"
                            bind:value={ep.tokenUrl}
                            class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                            placeholder="https://kc/realms/{`{realm}`}/protocol/openid-connect/token"
                          />
                          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{`{realm}`} is substituted from the resolver below; otherwise used verbatim.</p>
                        </div>
                        <div>
                          <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Realm resolver URL <span class="text-gray-400">(optional)</span></label>
                          <input
                            type="text"
                            bind:value={ep.realmResolverUrl}
                            class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                            placeholder="https://owner.api/owners/{`{subdomain}`}"
                          />
                          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Leave empty if Token URL has no {`{realm}`}.</p>
                        </div>
                        <div>
                          <label class="block text-xs text-gray-500 dark:text-gray-400 mb-1">Realm JSON path</label>
                          <input
                            type="text"
                            bind:value={ep.realmJsonPath}
                            class="w-full px-2 py-1 border border-gray-300 dark:border-gray-600 rounded bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
                            placeholder="realm or data.realm"
                          />
                        </div>
                      </div>
                    {/if}
                  </div>
                {/each}
                {#if endpoints.length === 0}
                  <p class="text-xs text-gray-500 dark:text-gray-400 italic">No endpoints configured. Click + to add one.</p>
                {/if}
              </div>
              <button
                type="button"
                onclick={addEndpoint}
                class="mt-2 px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded text-text-light dark:text-text-dark hover:bg-gray-100 dark:hover:bg-gray-700"
              >
                + Add endpoint
              </button>
            </div>
          </div>
        </div>

        <div>
          <button
            type="button"
            onclick={() => claimMapExpanded = !claimMapExpanded}
            class="flex items-center justify-between w-full text-sm font-medium text-text-light dark:text-text-dark mb-1"
          >
            <span>Parent claim map</span>
            <span class="text-xs">{claimMapExpanded ? '▾' : '▸'}</span>
          </button>
          {#if claimMapExpanded}
          <input
            type="text"
            bind:value={parentClaimMapText}
            class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md bg-bg-light dark:bg-bg-dark text-text-light dark:text-text-dark text-sm"
            placeholder="o=owner_id, org=org_id"
          />
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">For methods with a (google.api.http) parent binding like {`{parent=o/*}`}, fill the wildcard from the named JWT claim.</p>
          {/if}
        </div>

        <div class="border-t border-gray-300 dark:border-gray-600 pt-4">
          <button
            type="button"
            onclick={() => diagnosticsExpanded = !diagnosticsExpanded}
            class="flex items-center justify-between w-full text-sm font-semibold text-text-light dark:text-text-dark mb-2"
          >
            <span>Diagnostics</span>
            <span class="text-xs">{diagnosticsExpanded ? '▾' : '▸'}</span>
          </button>
          {#if diagnosticsExpanded}
          <div class="flex items-center justify-between">
            <div class="text-xs text-gray-500 dark:text-gray-400 truncate pr-3" title={logsFolder}>
              Logs: {logsFolder || '—'}
            </div>
            <button
              type="button"
              onclick={openLogs}
              class="shrink-0 px-3 py-1 text-sm border border-gray-300 dark:border-gray-600 rounded text-text-light dark:text-text-dark hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              Open logs folder
            </button>
          </div>
          <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">Rolling file at 5 MB · keeps last 5 · 30-day retention.</p>
          {/if}
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
            class="flex-1 px-4 py-2 bg-btn-light dark:bg-btn-dark hover:opacity-90 text-white text-sm rounded-md"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  </div>
{/if}
