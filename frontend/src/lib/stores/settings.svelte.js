import { GetSettings, SaveSettings } from '../../../wailsjs/go/main/App.js'
import { addInfo, addError } from './logs.svelte.js'

export const settings = $state({
  namespaces: ['default', 'am-dev', 'am-qa', 'am-demo'],
  portRangeStart: 35000,
  portRangeEnd: 60000,
  grpcPorts: [5001, 5002],
  nodePortHost: '127.0.0.1',
  tokenEndpoint: '',
  clientId: '',
  serviceExcludePatterns: ['*wassups'],
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
  settings.nodePortHost = cfg.nodePortHost || '127.0.0.1'
  settings.tokenEndpoint = cfg.tokenEndpoint || ''
  settings.clientId = cfg.clientId || ''
  settings.serviceExcludePatterns = cfg.serviceExcludePatterns || []
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
      nodePortHost: settings.nodePortHost,
      tokenEndpoint: settings.tokenEndpoint,
      clientId: settings.clientId,
      serviceExcludePatterns: settings.serviceExcludePatterns,
    })
    addInfo('Settings saved')
  } catch (err) {
    addError(`Failed to save settings: ${err}`)
  }
}
