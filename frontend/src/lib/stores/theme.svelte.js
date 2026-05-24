const STORAGE_KEY = 'grails:theme'

function readInitial() {
  if (typeof localStorage === 'undefined') return 'dark'
  const v = localStorage.getItem(STORAGE_KEY)
  return v === 'light' ? 'light' : 'dark'
}

function applyToDocument(theme) {
  if (typeof document === 'undefined') return
  if (theme === 'dark') document.documentElement.classList.add('dark')
  else document.documentElement.classList.remove('dark')
}

export const theme = $state({ value: readInitial() })

applyToDocument(theme.value)

export function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  if (typeof localStorage !== 'undefined') localStorage.setItem(STORAGE_KEY, theme.value)
  applyToDocument(theme.value)
}
