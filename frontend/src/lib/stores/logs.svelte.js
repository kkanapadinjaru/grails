function makeEntry(level, message) {
  return {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
    timestamp: new Date(),
    level,
    message,
  }
}

export const logs = $state({ entries: [] })

export function addLog(level, message) {
  logs.entries = [...logs.entries, makeEntry(level, message)]
}

export const addInfo = (m) => addLog('info', m)
export const addWarn = (m) => addLog('warning', m)
export const addError = (m) => addLog('error', m)
export const addDebug = (m) => addLog('debug', m)

export function clearLogs() {
  logs.entries = []
}

export function seedSampleLogs() {
  if (logs.entries.length > 0) return
  const seed = [
    ['info', 'Loaded 2 clusters from kubeconfig'],
    ['info', 'Loaded 2 clusters from kubeconfig'],
    ['info', 'Connected to cluster: docker-desktop'],
    ['info', 'Discovered 4 gRPC services'],
    ['info', 'Auto-selected service: grpc.health.v1.Health at 127.0.0.1:32000'],
    ['info', 'Discovering methods for grpc.health.v1.Health...'],
    ['info', 'Discovered 2 gRPC methods'],
  ]
  logs.entries = seed.map(([l, m]) => makeEntry(l, m))
}
