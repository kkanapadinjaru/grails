import { GetSettings, SaveSettings } from '../../../wailsjs/go/main/App.js'
import { addInfo, addError } from './logs.svelte.js'

export const settings = $state({
  namespaces: ['default', 'am-dev', 'am-qa', 'am-demo'],
  portRangeStart: 35000,
  portRangeEnd: 60000,
  grpcPorts: [5001, 5002],
  discoveryConcurrency: 5,
  nodePortHost: '127.0.0.1',
  authProvider: 'keycloak',
  clientId: '',
  authEndpoints: [],
  serviceExcludePatterns: ['*wassups'],
  parentClaimMap: { o: 'owner_id' },
  loaded: false,
})

export const ui = $state({
  showProfile: false,
  isHistoryPanelOpen: false,
})

function applyFromBackend(cfg) {
  settings.namespaces = cfg.namespaces || []
  settings.portRangeStart = cfg.portRangeStart || 35000
  settings.portRangeEnd = cfg.portRangeEnd || 60000
  settings.grpcPorts = cfg.grpcPorts || [5001, 5002]
  settings.discoveryConcurrency = cfg.discoveryConcurrency || 5
  settings.nodePortHost = cfg.nodePortHost || '127.0.0.1'
  settings.authProvider = cfg.authProvider || 'keycloak'
  settings.clientId = cfg.clientId || ''
  settings.authEndpoints = cfg.authEndpoints || []
  settings.serviceExcludePatterns = cfg.serviceExcludePatterns || []
  settings.parentClaimMap = cfg.parentClaimMap || {}
  settings.loaded = true
}

export async function loadSettings() {
  try {
    const cfg = await GetSettings()
    applyFromBackend(cfg)
    addInfo('Loaded settings from disk')
  } catch (err) {
    addError(`Failed to load settings: ${err}`)
  }
}

export async function saveSettings() {
  try {
    await SaveSettings({
      namespaces: settings.namespaces,
      portRangeStart: settings.portRangeStart,
      portRangeEnd: settings.portRangeEnd,
      grpcPorts: settings.grpcPorts,
      discoveryConcurrency: settings.discoveryConcurrency,
      nodePortHost: settings.nodePortHost,
      authProvider: settings.authProvider,
      clientId: settings.clientId,
      authEndpoints: settings.authEndpoints,
      serviceExcludePatterns: settings.serviceExcludePatterns,
      parentClaimMap: settings.parentClaimMap,
    })
    addInfo('Settings saved')
  } catch (err) {
    addError(`Failed to save settings: ${err}`)
  }
}
