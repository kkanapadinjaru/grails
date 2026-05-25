const STORAGE_KEY = 'grails:theme'
const PALETTE_KEY = 'grails:palette'

// Palette catalog — each entry maps to a [data-palette] block in style.css.
// Adding a new palette: append here and define matching CSS variables.
// `swatch` is the [bg, sidebar, btn] dark-mode triplet shown in the picker.
// Hard-coded so previews don't all collapse to the active palette's vars.
export const PALETTES = [
  {
    id: 'catppuccin',
    name: 'Catppuccin',
    blurb: 'Pastel mauve · Latte / Macchiato.',
    swatch: ['#24273a', '#1e2030', '#8aadf4'],
  },
  {
    id: 'one-dark-pro',
    name: 'One Dark Pro',
    blurb: 'Atom / VS Code — slate with cool blue.',
    swatch: ['#282c34', '#21252b', '#61afef'],
  },
  {
    id: 'tokyo-night',
    name: 'Tokyo Night',
    blurb: 'Deep navy · soft blue accent.',
    swatch: ['#1a1b26', '#16161e', '#7aa2f7'],
  },
  {
    id: 'dracula',
    name: 'Dracula',
    blurb: 'Bold purple on midnight.',
    swatch: ['#282a36', '#21222c', '#bd93f9'],
  },
  {
    id: 'nebula',
    name: 'Nebula',
    blurb: 'Indigo drift — navy fading to lavender.',
    swatch: ['#1a1a40', '#13132f', '#7070a3'],
  },
]

const VALID_PALETTES = new Set(PALETTES.map(p => p.id))

function readInitialMode() {
  if (typeof localStorage === 'undefined') return 'dark'
  const v = localStorage.getItem(STORAGE_KEY)
  return v === 'light' ? 'light' : 'dark'
}

function readInitialPalette() {
  if (typeof localStorage === 'undefined') return 'catppuccin'
  const v = localStorage.getItem(PALETTE_KEY)
  return v && VALID_PALETTES.has(v) ? v : 'catppuccin'
}

function applyToDocument(mode, palette) {
  if (typeof document === 'undefined') return
  if (mode === 'dark') document.documentElement.classList.add('dark')
  else document.documentElement.classList.remove('dark')
  if (palette === 'catppuccin') {
    delete document.documentElement.dataset.palette
  } else {
    document.documentElement.dataset.palette = palette
  }
}

export const theme = $state({
  value: readInitialMode(),
  palette: readInitialPalette(),
})

applyToDocument(theme.value, theme.palette)

export function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  if (typeof localStorage !== 'undefined') localStorage.setItem(STORAGE_KEY, theme.value)
  applyToDocument(theme.value, theme.palette)
}

export function setPalette(id) {
  if (!VALID_PALETTES.has(id)) return
  theme.palette = id
  if (typeof localStorage !== 'undefined') localStorage.setItem(PALETTE_KEY, id)
  applyToDocument(theme.value, theme.palette)
}
