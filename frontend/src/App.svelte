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
</script>

<div
  class="h-screen w-screen bg-bg-light dark:bg-bg-dark flex flex-col overflow-hidden"
  onclick={dismissMenus}
  role="presentation"
>
  <Header />

  <div class="flex-1 flex overflow-hidden relative min-h-0">
    <Sidebar />
    <main class="flex-1 flex flex-col bg-gray-50 dark:bg-gray-900 min-w-0">
      <div class="flex-1 flex min-h-0">
        <RequestPanel />
        <ResponsePanel />
      </div>
    </main>
    <HistoryPanel />
  </div>

  <LogPanel />
</div>

<LoginModal />
<ProfileModal />
