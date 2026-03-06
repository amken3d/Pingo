// ┌──────────────────────────────────────────────────────────────────────┐
// │  layout.go — Main Layout, Navigation & Theme Switching             │
// │                                                                    │
// │  Demonstrates:                                                     │
// │  • ui.VStack / ui.HStack — vertical and horizontal stacking        │
// │  • ui.AppBar — top application bar with action buttons             │
// │  • ui.Expanded — fills remaining space in a stack                  │
// │  • ui.Scroll — vertical scroll wrapper                             │
// │  • ui.Style(view).Padding() — modifier chaining                    │
// │  • ui.ViewFunc — bridging a lower-level Gio widget into a View     │
// │  • Theme switching via ThemeRefValue.Set()                         │
// │  • Page routing with a plain Go switch statement                   │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"gioui.org/layout"

	"github.com/amken3d/immygo/theme"
	"github.com/amken3d/immygo/ui"
)

// buildApp constructs the top-level View tree. It is called every frame.
//
// Layout structure:
//
//	VStack
//	├── AppBar          (fixed at top)
//	└── HStack
//	    ├── SideNav     (fixed width, left)
//	    └── Expanded
//	        └── Scroll
//	            └── pageContent()   (switches on currentPage)
func buildApp() ui.View {
	// ── ViewFunc bridge ─────────────────────────────────────────────
	// The sideNav is a *widget.SideNav (lower-level API) stored in
	// state.go. We wrap it in ui.ViewFunc to use it in the declarative
	// View tree. ViewFunc adapts any func(gtx, th) → Dimensions into
	// a View.
	sideNavView := ui.ViewFunc(func(gtx layout.Context, th *theme.Theme) layout.Dimensions {
		return sideNav.Layout(gtx, th)
	})

	return ui.VStack(
		headerBar(),
		// Flex(1) gives the body all remaining vertical space after the AppBar,
		// so the Scroll inside knows its available height and can scroll.
		ui.Flex(1, ui.HStack(
			sideNavView,
			ui.Flex(1,
				ui.Style(pageContent()).Padding(24),
			),
		)),
	)
}

// headerBar creates the top AppBar.
//
// ui.AppBar("title") creates a bar. .Actions(...) adds views to the right.
// Any View can be an action — here we use Buttons.
func headerBar() ui.View {
	spec := currentSpec()
	return ui.AppBar("Pingo — "+spec.Name).
		Actions(
			// Board toggle buttons
			ui.Button("Pico").OnClick(func() { boardChoice.Set(0) }),
			ui.Button("Pico 2").OnClick(func() { boardChoice.Set(1) }),

			ui.Text("Dark"),
			themeToggle,
		)
}

// pageContent routes to the correct page based on currentPage.
//
// This uses a plain Go switch. ImmyGo also offers ui.Switch(index, views...)
// for index-based routing, but a plain switch gives more flexibility
// (e.g., passing different arguments to each page builder).
func pageContent() ui.View {
	switch currentPage {
	case 0:
		return pinoutPage()
	case 1:
		return myPinsPage()
	case 2:
		return settingsPage()
	default:
		return pinoutPage()
	}
}
