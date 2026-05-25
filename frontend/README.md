# Grails Frontend

Svelte 5 + Vite + Tailwind CSS frontend for the Grails gRPC client.

## Stack

- **Svelte 5** — using runes (`$state`, `$derived`, `$effect`) for reactivity
- **Tailwind CSS** — utility-first styling with custom theme palettes via CSS variables
- **Vite** — dev server with HMR, production bundling

## Development

From the project root:

```bash
wails dev
```

Or to run the frontend standalone (without Go backend):

```bash
npm run dev
```

The dev server runs on `http://localhost:34115` with access to Go methods via the Wails bridge.

## Structure

```
src/
├── lib/
│   ├── components/       # UI components
│   │   ├── Header.svelte         # Cluster/namespace selectors, theme toggle, user menu
│   │   ├── Sidebar.svelte        # Service dropdown, method list with filter
│   │   ├── RequestPanel.svelte   # Token field, request body editor, send button
│   │   ├── ResponsePanel.svelte  # Response display with status/timing
│   │   ├── HistoryPanel.svelte   # Collapsible request history
│   │   ├── LogPanel.svelte       # Collapsible log viewer with level filter
│   │   ├── LoginModal.svelte     # OIDC login form
│   │   └── ProfileModal.svelte   # Settings modal
│   └── stores/           # Reactive state
│       ├── connection.svelte.js  # Cluster, service, method, request/response state
│       ├── auth.svelte.js        # Token management, login/logout
│       ├── settings.svelte.js    # Persisted user settings
│       ├── theme.svelte.js       # Dark/light mode, palette selection
│       └── logs.svelte.js        # In-app log entries
├── App.svelte            # Root layout with resizable split panes
├── style.css             # Theme palettes, scrollbar, selection, focus styles
└── main.js               # Svelte mount
```

## Theming

Palettes are defined as CSS custom properties in `style.css` and selected via `data-palette` on `<html>`. Each palette provides light and dark variants. The active palette is toggled with the sun/moon button in the header.

Available palettes: Catppuccin, Jungle, Tokyo Night, Dracula, Nebula.

## Fonts

Monospace content (request body, response, logs) uses: JetBrains Mono → Cascadia Mono → system monospace fallbacks.
