<script>
  import Header from './lib/components/Header.svelte'
  import Sidebar from './lib/components/Sidebar.svelte'
  import RequestPanel from './lib/components/RequestPanel.svelte'
  import ResponsePanel from './lib/components/ResponsePanel.svelte'
  import HistoryPanel from './lib/components/HistoryPanel.svelte'
  import LogPanel from './lib/components/LogPanel.svelte'
  import LoginModal from './lib/components/LoginModal.svelte'
  import ProfileModal from './lib/components/ProfileModal.svelte'

  import { auth, initAuthEvents, rehydrateAuth } from './lib/stores/auth.svelte.js'
  import { loadClusters } from './lib/stores/connection.svelte.js'
  import { loadSettings } from './lib/stores/settings.svelte.js'

  // Wails injects window.go after the runtime is ready; defer until then.
  if (typeof window !== 'undefined') {
    const boot = async () => {
      initAuthEvents()
      await loadSettings()
      await rehydrateAuth()
      await loadClusters()
    }
    if (window['go']) {
      boot()
    } else {
      window.addEventListener('DOMContentLoaded', () => {
        const tick = () => {
          if (window['go']) boot()
          else setTimeout(tick, 50)
        }
        tick()
      })
    }
  }

  function dismissMenus() {
    auth.showUserMenu = false
  }

  // Resizable split pane state
  let splitPercent = $state(50)
  let dragging = $state(false)
  let containerEl = $state(null)

  function onGutterDown(e) {
    e.preventDefault()
    dragging = true
    document.addEventListener('mousemove', onDrag)
    document.addEventListener('mouseup', onDragEnd)
  }

  function onDrag(e) {
    if (!dragging || !containerEl) return
    const rect = containerEl.getBoundingClientRect()
    const pct = ((e.clientX - rect.left) / rect.width) * 100
    splitPercent = Math.min(70, Math.max(25, pct))
  }

  function onDragEnd() {
    dragging = false
    document.removeEventListener('mousemove', onDrag)
    document.removeEventListener('mouseup', onDragEnd)
  }
</script>

<div
  class="h-screen w-screen bg-bg-light dark:bg-bg-dark flex flex-col overflow-hidden {dragging ? 'select-none' : ''}"
  onclick={dismissMenus}
  role="presentation"
>
  <Header />

  <div class="flex-1 flex overflow-hidden relative min-h-0">
    <Sidebar />
    <main class="flex-1 flex flex-col bg-gray-50 dark:bg-gray-900 min-w-0">
      <div class="flex-1 flex min-h-0" bind:this={containerEl}>
        <div style="width: {splitPercent}%; min-width: 0;" class="flex">
          <RequestPanel />
        </div>
        <div
          onmousedown={onGutterDown}
          class="w-1 shrink-0 cursor-col-resize hover:bg-btn-light dark:hover:bg-btn-dark transition-colors {dragging ? 'bg-btn-light dark:bg-btn-dark' : 'bg-gray-200 dark:bg-gray-700'}"
        ></div>
        <div style="width: {100 - splitPercent}%; min-width: 0;" class="flex">
          <ResponsePanel />
        </div>
      </div>
    </main>
    <HistoryPanel />
  </div>

  <LogPanel />
</div>

<LoginModal />
<ProfileModal />
