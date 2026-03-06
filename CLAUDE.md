# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Pingo is a Raspberry Pi Pico / Pico 2 pin selector desktop app built with the **ImmyGo** UI framework (`github.com/amken3d/immygo`). It serves as both a useful tool and a learning guide for ImmyGo. The UI is built on top of Gio (`gioui.org`).

## Build & Run

```bash
# Run the app
go run .

# Build
go build .
```

**Linux system dependencies** (Gio requires):
```bash
apt install libwayland-dev libxkbcommon-x11-dev libgles2-mesa-dev \
  libegl1-mesa-dev libx11-xcb-dev libvulkan-dev libxcursor-dev libxfixes-dev
```

No test suite exists in this project.

## Architecture

All Go source files are in package `main` at the root, plus one sub-package:

- **`main.go`** — Entry point. Calls `ui.Run()` with a build function, window size, and theme ref.
- **`state.go`** — All reactive state (`ui.NewState[T]`), the persistent `SideNav` widget, AI engine init, and helpers. This is the central wiring file.
- **`layout.go`** — Top-level layout: `AppBar` + `SideNav` + page router (plain `switch` on `currentPage`).
- **`page_*.go`** — Each page is a function returning `ui.View`. Pages: pinout viewer, pin selector, selected pins summary, board info, AI assistant.
- **`svg.go`** — Embeds `assets/pico.svg` via `//go:embed` and rasterizes it at startup using `oksvg`/`rasterx`.
- **`pindata/`** — Pure domain model (no UI imports). Board specs, pin definitions, conflict detection, category helpers.

## Key ImmyGo Patterns

- **Declarative views**: Build functions return `ui.View` trees (`VStack`, `HStack`, `Card`, `Text`, etc.). ImmyGo diffs and repaints.
- **Reactive state**: `ui.NewState[T](initial)` creates reactive values. `.Get()` reads, `.Set(v)` triggers re-render. For maps, always replace the entire map (immutable update pattern).
- **Persistent widgets**: Stateful Gio widgets (`SideNav`, `Toggle`, `Clickable`) must be stored in package-level vars — recreating them per frame loses click state.
- **ViewFunc bridge**: `ui.ViewFunc(func(gtx, th) Dimensions)` adapts lower-level Gio widgets (like `widget.SideNav` or `ai.ChatPanel`) into the declarative View tree.
- **Theme switching**: `ui.NewThemeRef()` + `ui.WithThemeRef()` enables runtime theme swaps via `themeRefVal.Set(newTheme)`.
- **Modifier chaining**: `ui.Style(view).Background(c).Padding(n).Rounded(r)` — each modifier returns `*Styled` which implements `View`.

## AI Integration

ImmyGo's `ai` package supports Yzma (local GGUF), Ollama, Anthropic Claude (`ANTHROPIC_API_KEY`), and MCP servers. The engine auto-detects available providers. `ai.NewChatPanel(assistant)` provides a ready-made chat widget. AI calls (`assistant.Chat()`) block and should run in goroutines.