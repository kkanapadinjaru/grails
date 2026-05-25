# Grails

A desktop gRPC client for Kubernetes environments. Grails automatically discovers clusters from your kubeconfig, port-forwards to gRPC services, discovers available methods via server reflection, and lets you compose and send requests — all from a single UI.

## Features

- **Cluster discovery** — reads kubeconfig and lists available contexts
- **Namespace scanning** — discovers gRPC-enabled pods across configured namespaces
- **Server reflection** — lists services and methods without proto files
- **Request editor** — JSON body with auto-generated skeletons and random sample data
- **Auth integration** — Keycloak OIDC token generation with automatic refresh
- **Per-service auth toggle** — skip authentication for services that don't require it
- **Request history** — replay previous calls with one click
- **Theming** — 5 dark/light palette pairs (Catppuccin, Jungle, Tokyo Night, Dracula, Nebula)
- **Resizable panels** — drag the gutter between request and response for more space
- **Rolling logs** — file-based logging with in-app log viewer

## Tech Stack

- **Backend** — Go, Wails v2, k8s.io/client-go, kubectl (port-forwarding), grpcurl (reflection & invocation)
- **Frontend** — Svelte 5, Tailwind CSS, Vite

## Prerequisites

- Go 1.26+
- Node.js 18+
- [Wails CLI](https://wails.io/docs/gettingstarted/installation) v2.12+
- `kubectl` in PATH
- `grpcurl` in PATH
- A valid kubeconfig with cluster access

## Development

```bash
wails dev
```

Runs the app with Vite HMR for the frontend. Backend changes trigger a rebuild automatically.

## Building

```bash
wails build -clean -trimpath -ldflags "-s -w -H windowsgui" -tags "production" -upx
```

Produces a compressed, production-ready executable in `build/bin/`.

## Configuration

Settings are accessible via the gear icon in the header:

- **Namespaces** — which namespaces to scan for gRPC services
- **Port range** — local port range for kubectl port-forwards
- **gRPC ports** — container ports to probe for reflection
- **Discovery concurrency** — parallel port-forwards during scanning
- **Exclude patterns** — glob patterns to skip certain k8s services
- **Auth endpoints** — per-cluster/namespace Keycloak token URLs
- **Parent claim map** — JWT claim mapping for google.api.http parent bindings

Settings persist to the Go backend config file. Theme and auth overrides persist in the WebView's localStorage.

## Project Structure

```
grails/
├── build/          # Build assets (icons, manifests, installer scripts)
├── frontend/       # Svelte 5 + Tailwind frontend
│   └── src/
│       ├── lib/
│       │   ├── components/   # UI components
│       │   └── stores/       # Svelte 5 reactive stores
│       ├── App.svelte
│       └── style.css         # Theme palettes & global styles
├── grpc/           # gRPC reflection & request execution (grpcurl wrapper)
├── kubernetes/     # Cluster, namespace, service discovery & port-forwarding
├── logging/        # Rolling file logger
├── cmdutil/        # Platform-specific process helpers (hide console on Windows)
├── auth/           # OIDC token generation
├── config/         # Settings persistence
├── app.go          # Wails app struct & bound methods
└── main.go         # Entry point
```

## License

Private — not for redistribution.
