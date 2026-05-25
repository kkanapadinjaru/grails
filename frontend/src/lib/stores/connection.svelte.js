import {
  GetClusters,
  ConnectToCluster,
  DisconnectFromCluster,
  IsConnected,
  SelectNamespace,
  GetGrpcMethods,
  DescribeGrpcMethod,
  GenerateRequestSkeleton,
  GenerateSampleRequest,
  SendGrpcRequest,
} from '../../../wailsjs/go/main/App.js'
import { auth } from './auth.svelte.js'
import { addInfo, addError, addWarn } from './logs.svelte.js'

const HISTORY_LIMIT = 100

export const connection = $state({
  clusters: [],
  selectedCluster: '',
  connectedContext: '',
  isConnecting: false,

  namespaces: [],
  selectedNamespace: '',

  services: [],
  selectedService: '',
  isLoadingServices: false,

  methods: [],
  selectedMethod: '',
  isLoadingMethods: false,

  requestType: '',
  responseType: '',
  requestBody: '{}',

  response: '',
  responseStatus: '',
  responseTime: '',
  isSending: false,
  isResponseError: false,
})

export const history = $state({ entries: [] })

let nextHistoryId = 1

// Per-service auth requirement override. When a service is toggled to
// "not required", we skip the login modal and send without a token.
// Keyed by service displayName. `true` = auth required (default), `false` = skip.
export const authOverrides = $state({})

export function isAuthRequired(serviceName) {
  if (!(serviceName in authOverrides)) return true
  return authOverrides[serviceName]
}

export function toggleAuthRequired(serviceName) {
  if (!serviceName) return
  const current = isAuthRequired(serviceName)
  authOverrides[serviceName] = !current
}

// Per-(service,method) request body cache so user edits survive method
// switching within a session. Keyed by `${service.displayName}::${method}`.
const bodyCache = new Map()

export async function loadClusters() {
  try {
    const list = await GetClusters()
    connection.clusters = (list || []).map(c => ({ name: c.name, context: c.context, server: c.server }))
    if (connection.clusters.length > 0 && !connection.selectedCluster) {
      connection.selectedCluster = connection.clusters[0].context
    }
    addInfo(`Loaded ${connection.clusters.length} clusters from kubeconfig`)
  } catch (err) {
    addError(`Failed to load clusters: ${err}`)
  }
}

export async function connect() {
  const target = connection.selectedCluster
  if (!target) {
    addError('No cluster selected')
    return
  }
  connection.isConnecting = true
  try {
    if (connection.connectedContext && connection.connectedContext !== target) {
      addInfo(`Disconnecting from ${connection.connectedContext} before connecting to ${target}`)
      try { await DisconnectFromCluster() } catch (err) { addWarn(`Disconnect warning: ${err}`) }
      resetClusterState()
    }

    addInfo(`Connecting to cluster: ${target}`)
    const nsResults = await ConnectToCluster(target)

    const allowed = []
    for (const ns of nsResults || []) {
      if (ns.allowed) {
        allowed.push(ns.name)
      } else {
        addWarn(`Namespace ${ns.name} not accessible: ${ns.reason || 'forbidden'}`)
      }
    }

    connection.namespaces = allowed
    connection.connectedContext = (await IsConnected()) ? target : ''

    addInfo(`Connected to cluster: ${target}`)
    addInfo(`Accessible namespaces: ${allowed.length > 0 ? allowed.join(', ') : '(none)'}`)

    if (allowed.length > 0) {
      await selectNamespace(allowed[0])
    } else {
      connection.selectedNamespace = ''
    }
  } catch (err) {
    addError(`Failed to connect: ${err}`)
    connection.connectedContext = ''
  } finally {
    connection.isConnecting = false
  }
}

export async function selectNamespace(namespace) {
  if (!namespace) {
    connection.selectedNamespace = ''
    connection.services = []
    connection.selectedService = ''
    return
  }
  connection.selectedNamespace = namespace
  connection.isLoadingServices = true
  connection.services = []
  connection.selectedService = ''
  resetMethodState()
  bodyCache.clear()
  try {
    addInfo(`Discovering gRPC services in namespace: ${namespace}`)
    const list = await SelectNamespace(namespace)
    connection.services = list || []
    if (connection.services.length > 0) {
      addInfo(`Found ${connection.services.length} gRPC service(s) in ${namespace}`)
      await selectService(connection.services[0].displayName)
    } else {
      addWarn(`No gRPC services found in namespace ${namespace}`)
    }
  } catch (err) {
    addError(`Failed to discover services in ${namespace}: ${err}`)
  } finally {
    connection.isLoadingServices = false
  }
}

export async function selectService(displayName) {
  connection.selectedService = displayName
  resetMethodState()

  const svc = connection.services.find(s => s.displayName === displayName)
  if (!svc) return

  connection.isLoadingMethods = true
  try {
    addInfo(`Listing methods on ${svc.serviceName} (${svc.localAddress})`)
    const names = await GetGrpcMethods(svc.localAddress, svc.serviceName)
    connection.methods = (names || []).map(name => ({ name, requestType: '', responseType: '' }))
    if (connection.methods.length > 0) {
      await selectMethod(connection.methods[0].name)
    } else {
      addWarn(`No methods found on ${svc.serviceName}`)
    }
  } catch (err) {
    addError(`Failed to list methods on ${svc.serviceName}: ${err}`)
  } finally {
    connection.isLoadingMethods = false
  }
}

export async function selectMethod(methodName) {
  connection.selectedMethod = methodName
  const svc = connection.services.find(s => s.displayName === connection.selectedService)
  if (!svc || !methodName) {
    connection.requestType = ''
    connection.responseType = ''
    connection.requestBody = '{}'
    return
  }

  try {
    const desc = await DescribeGrpcMethod(svc.localAddress, svc.serviceName, methodName)
    connection.requestType = desc?.requestType || ''
    connection.responseType = desc?.responseType || ''

    const idx = connection.methods.findIndex(m => m.name === methodName)
    if (idx !== -1) {
      connection.methods[idx] = {
        name: methodName,
        requestType: connection.requestType,
        responseType: connection.responseType,
      }
    }

    const cacheKey = `${svc.displayName}::${methodName}`
    if (bodyCache.has(cacheKey)) {
      connection.requestBody = bodyCache.get(cacheKey)
    } else if (connection.requestType) {
      const skel = await GenerateRequestSkeleton(svc.localAddress, connection.requestType)
      connection.requestBody = skel || '{}'
      bodyCache.set(cacheKey, connection.requestBody)
    } else {
      connection.requestBody = '{}'
    }
  } catch (err) {
    addError(`Failed to describe ${methodName}: ${err}`)
  }
}

export async function generateSampleRequest() {
  const svc = connection.services.find(s => s.displayName === connection.selectedService)
  if (!svc || !connection.selectedMethod) {
    addWarn('No method selected to generate sample for')
    return
  }
  try {
    const sample = await GenerateSampleRequest(svc.localAddress, svc.serviceName, connection.selectedMethod)
    connection.requestBody = sample || '{}'
    bodyCache.set(`${svc.displayName}::${connection.selectedMethod}`, connection.requestBody)
    addInfo(`Generated sample request for ${connection.requestType}`)
  } catch (err) {
    addError(`Failed to generate sample request: ${err}`)
  }
}

// cacheCurrentBody is called on every requestBody change so per-method edits
// persist across method-switching.
export function cacheCurrentBody() {
  if (!connection.selectedService || !connection.selectedMethod) return
  bodyCache.set(`${connection.selectedService}::${connection.selectedMethod}`, connection.requestBody)
}

export async function sendRequest() {
  const svc = connection.services.find(s => s.displayName === connection.selectedService)
  if (!svc || !connection.selectedMethod) {
    addWarn('Select a service and method before sending')
    return
  }

  // Validate JSON locally so we get a clearer error than grpcurl's parser would give.
  try {
    JSON.parse(connection.requestBody || '{}')
  } catch (err) {
    const msg = `Invalid request JSON: ${err.message || err}`
    connection.response = msg
    connection.responseStatus = 'INVALID JSON'
    connection.responseTime = ''
    connection.isResponseError = true
    addError(msg)
    return
  }

  connection.isSending = true
  connection.isResponseError = false
  connection.responseStatus = ''
  connection.responseTime = ''
  const fullMethod = `${svc.serviceName}.${connection.selectedMethod}`
  addInfo(`Sending ${fullMethod} → ${svc.localAddress}`)

  const startedAt = performance.now()
  let responseText = ''
  let status = 'OK'
  let isError = false
  try {
    responseText = await SendGrpcRequest(
      svc.localAddress,
      svc.serviceName,
      connection.selectedMethod,
      connection.requestBody || '{}',
      isAuthRequired(connection.selectedService) ? (auth.bearerToken || '') : '',
    )
  } catch (err) {
    isError = true
    responseText = String(err)
    status = parseGrpcStatus(responseText)
    addError(`Request failed: ${responseText}`)
  }
  const durationMs = Math.round(performance.now() - startedAt)

  connection.response = formatResponse(responseText)
  connection.responseStatus = status
  connection.responseTime = `${durationMs}ms`
  connection.isResponseError = isError
  connection.isSending = false

  if (!isError) {
    addInfo(`Request completed in ${durationMs}ms`)
  }

  appendHistory({
    service: connection.selectedService,
    serviceName: svc.serviceName,
    method: connection.selectedMethod,
    localAddress: svc.localAddress,
    request: connection.requestBody || '{}',
    response: connection.response,
    status,
    isError,
    durationMs,
  })
}

function appendHistory(entry) {
  const now = new Date()
  const time = now.toTimeString().slice(0, 5)
  history.entries.unshift({
    id: nextHistoryId++,
    timestamp: now.getTime(),
    time,
    ...entry,
  })
  if (history.entries.length > HISTORY_LIMIT) {
    history.entries.length = HISTORY_LIMIT
  }
}

function parseGrpcStatus(errText) {
  const m = errText.match(/Code:\s*([A-Za-z_]+)/)
  if (m) return m[1].toUpperCase()
  if (/^rpc error:/i.test(errText)) return 'ERROR'
  return 'ERROR'
}

function formatResponse(text) {
  if (!text) return ''
  try {
    return JSON.stringify(JSON.parse(text), null, 2)
  } catch {
    return text
  }
}

export function clearHistory() {
  history.entries = []
}

export function replayHistoryEntry(entry) {
  if (!entry) return
  // If the recorded service is still in the current list, preserve method
  // selection so the request body editor lights up properly. Otherwise we
  // just paste the request/response without touching selection state.
  const svc = connection.services.find(s => s.displayName === entry.service)
  if (svc) {
    connection.selectedService = entry.service
    connection.selectedMethod = entry.method
  }
  connection.requestBody = entry.request
  connection.response = entry.response
  connection.responseStatus = entry.status
  connection.responseTime = `${entry.durationMs}ms`
  connection.isResponseError = !!entry.isError
}

export async function disconnect() {
  try {
    await DisconnectFromCluster()
  } catch (err) {
    addError(`Disconnect failed: ${err}`)
  }
  resetClusterState()
  bodyCache.clear()
  addInfo('Disconnected')
}

function resetClusterState() {
  connection.connectedContext = ''
  connection.namespaces = []
  connection.selectedNamespace = ''
  connection.services = []
  connection.selectedService = ''
  resetMethodState()
}

function resetMethodState() {
  connection.methods = []
  connection.selectedMethod = ''
  connection.requestType = ''
  connection.responseType = ''
  connection.requestBody = '{}'
}
