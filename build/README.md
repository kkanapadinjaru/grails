# Build Directory

Build assets and platform-specific files for Grails.

## Structure

```
build/
├── bin/              # Compiled output (grails.exe)
├── darwin/           # macOS plist files
├── windows/          # Windows manifest, icon, installer scripts
│   ├── icon.ico          # Auto-generated from appicon.png
│   ├── info.json         # App metadata (version, copyright, etc.)
│   ├── wails.exe.manifest
│   └── installer/        # NSIS installer scripts
└── appicon.png       # Source icon (256x256) — used to generate platform icons
```

## Icon

`appicon.png` is the source icon. On build, Wails generates `windows/icon.ico` from it automatically. To update the app icon, replace `appicon.png` and rebuild.

Requirements:
- 256x256 PNG
- Transparent background recommended (for visibility on both dark and light OS themes)
- Graphic should fill ~85-90% of the canvas

## Building

From the project root:

```bash
# Development
wails dev

# Production (Windows, compressed with UPX)
wails build -clean -trimpath -ldflags "-s -w -H windowsgui" -tags "production" -upx
```

## Windows Installer

The `windows/installer/` directory contains NSIS scripts for generating a Windows installer. Run `wails build -nsis` to produce an installer executable alongside the main binary.

## Metadata

Edit `windows/info.json` to update:
- Application name and description
- Version number
- Company name
- Copyright
