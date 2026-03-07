// ┌──────────────────────────────────────────────────────────────────────┐
// │  Pingo — RPi Pico / Pico 2 Pin Selector                           │
// │  A demo application showcasing the ImmyGo UI framework.           │
// │                                                                    │
// │  This project is structured as a learning guide for building       │
// │  desktop apps with ImmyGo. Each file covers a specific concept:   │
// │                                                                    │
// │  main.go           → App entry point, ui.Run, theme, options       │
// │  state.go          → Reactive state, AI engine, persistent widgets │
// │  layout.go         → Main layout: AppBar, SideNav, page routing    │
// │  page_pinout.go    → Pinout + pin selector + AI assist             │
// │  page_mypins.go    → Selected pins summary, conflict detection     │
// │  page_settings.go  → AI provider & app settings                    │
// │  svg.go            → Embedding & rasterising SVG assets            │
// │  pindata/          → Domain model (no UI dependency)               │
// └──────────────────────────────────────────────────────────────────────┘
//
// # How ImmyGo Apps Work
//
//  1. Call ui.Run() with a title, a build function, and options.
//  2. The build function returns a ui.View tree every frame.
//  3. ImmyGo handles Gio's layout.Context/Dimensions internally.
//  4. Use State[T] for reactive values that trigger re-renders.
//  5. Stateful widgets (SideNav, Toggle, etc.) must persist across
//     frames — store them in package-level variables.
//
// # Running
//
//	go run .
//
// Requires Gio system deps on Linux:
//
//	apt install libwayland-dev libxkbcommon-x11-dev libgles2-mesa-dev \
//	  libegl1-mesa-dev libx11-xcb-dev libvulkan-dev libxcursor-dev libxfixes-dev
package main

import (
	"github.com/amken3d/immygo/ui"
)

func main() {
	// Load all embedded SVGs (board + chip) into raster images.
	// See svg.go for how embed + oksvg/rasterx work together.
	loadAllSVGs()

	// Load persisted settings into UI widgets before AI init.
	loadSettingsIntoUI()

	// Initialise the AI engine. See state.go for engine/assistant setup.
	initAI()

	// ── ui.Run ──────────────────────────────────────────────────────
	// This is the entry point for every ImmyGo app.
	//
	// Arguments:
	//   1. Window title (string)
	//   2. Build function — called every frame, returns the View tree
	//   3. Options — ui.Size, ui.Dark, ui.WithThemeRef, ui.OnInit, etc.
	//
	// The build function is pure: it reads state and returns Views.
	// ImmyGo diffs the tree and only repaints what changed.
	ui.Run("Pingo — RPi Pico Pin Selector", func() ui.View {
		return buildApp()
	},
		// Set the initial window size to 1280×800.
		ui.Size(1280, 800),

		// WithThemeRef allows runtime theme switching.
		// We pass a *ThemeRefValue created in state.go; calling
		// themeRefVal.Set(newTheme) swaps the theme live.
		ui.WithThemeRef(themeRefVal),
	)
}
