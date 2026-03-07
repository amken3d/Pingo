// ┌──────────────────────────────────────────────────────────────────────┐
// │  state.go — Reactive State Management                              │
// │                                                                    │
// │  ImmyGo provides State[T], a generic thread-safe reactive value.  │
// │  When you call state.Set(v), the UI automatically re-renders on   │
// │  the next frame. Use state.Get() in your build function to read.  │
// │                                                                    │
// │  Key rules:                                                        │
// │  • State is goroutine-safe (uses sync.RWMutex internally).        │
// │  • Calling Set() bumps an internal version counter that ImmyGo    │
// │    uses to detect changes.                                         │
// │  • Update(fn) applies a transformation atomically.                │
// │                                                                    │
// │  For stateful Gio widgets (SideNav, Toggle, Clickable, Editor),   │
// │  you MUST store the widget in a package-level var so it persists  │
// │  across frames. Creating a new widget every frame destroys the    │
// │  click/press state and breaks interactivity.                       │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"gioui.org/layout"
	giowidget "gioui.org/widget"

	"github.com/amken3d/Pingo/pindata"
	"github.com/amken3d/immygo/ai"
	"github.com/amken3d/immygo/theme"
	"github.com/amken3d/immygo/ui"
	"github.com/amken3d/immygo/widget"
)

// ─── Reactive State Variables ─────────────────────────────────────────
//
// ui.NewState[T](initial) creates a reactive value of any type.
// Read with .Get(), write with .Set(v) or .Update(fn).

var (
	// boardChoice toggles between Pico (0) and Pico 2 (1).
	// Changing this causes the entire UI to re-render with the
	// correct board specifications.
	boardChoice = ui.NewState(0)

	// activeFilter holds the current peripheral category filter
	// on the Selector page ("All", "SPI", "I2C", etc.).
	activeFilter = ui.NewState("All")

	// selections tracks which GPIO pins the user has selected
	// and what function they chose for each one.
	selections = ui.NewState(map[int]pindata.Function{})

	// selectedPeriphFunc holds the peripheral function being assigned
	// (e.g., "SPI0 RX"). Empty string means no function is active.
	selectedPeriphFunc = ui.NewState("")

	// isDark tracks the current theme mode for the toggle.
	isDark = ui.NewState(false)

	// AI settings state.
	aiProviderChoice = ui.NewState("auto")
	aiTemperature    = ui.NewState(float32(0.7))

	// aiStatus holds a human-readable status string for the AI engine.
	// Values: "loading", "ready: <provider>", "error: <message>"
	aiStatus = ui.NewState("loading")
)

// Persistent theme toggle widget — must live at package level.
var themeToggle = ui.Toggle(false).OnChange(func(on bool) {
	isDark.Set(on)
	if on {
		themeRefVal.Set(theme.FluentDark())
	} else {
		themeRefVal.Set(theme.FluentLight())
	}
})

// ─── Theme Reference ──────────────────────────────────────────────────
//
// ui.NewThemeRef(initial) creates a *ThemeRefValue that can be passed
// to ui.WithThemeRef(). Calling themeRefVal.Set(newTheme) swaps the
// theme at runtime — all widgets re-render with new colors instantly.

var themeRefVal = ui.NewThemeRef(theme.FluentLight())

// ─── Persistent SideNav Widget ────────────────────────────────────────
//
// IMPORTANT: This is a lower-level widget (widget.NewSideNav), not the
// declarative ui.SideNav(). We store it at package level because Gio's
// Clickable widgets track mouse press/release events across frames.
// If we recreated the SideNav every frame (as ui.SideNav() does),
// the Clickable that received the mouse-down would be discarded before
// the mouse-up, and clicks would never register.
//
// We wrap this in a ui.ViewFunc in layout.go to bridge it into the
// declarative View tree.

var (
	currentPage int

	sideNav = widget.NewSideNav(
		widget.NavItem{Label: "Pinout", Icon: "\u25A0"},   // ■
		widget.NavItem{Label: "My Pins", Icon: "\u2611"},  // ☑
		widget.NavItem{Label: "Settings", Icon: "\u2699"}, // ⚙
	).WithOnSelect(func(i int) { currentPage = i }).WithWidth(180)
)

// ─── AI Engine ────────────────────────────────────────────────────────
//
// ImmyGo's ai package supports multiple providers:
//   - Yzma  (local in-process LLM via GGUF models — most private)
//   - Ollama (local server, e.g. qwen2.5-coder)
//   - Anthropic Claude (cloud API via ANTHROPIC_API_KEY)
//   - MCP server (external tool integration)
//
// The engine auto-detects available providers at startup.
// ai.NewAssistant wraps an Engine with conversation history management.
// ai.NewChatPanel provides a ready-made chat UI widget.

var (
	engine    *ai.Engine
	assistant *ai.Assistant
	chatPanel *ai.ChatPanel
)

func pingoSystemPrompt() string {
	return `You are Pingo, an expert assistant for Raspberry Pi Pico and Pico 2 hardware design.
You help users choose the right GPIO pins for their projects.
When asked about pin selection:
- Consider peripheral function conflicts (SPI, I2C, UART share GPIOs)
- Note PWM slice sharing (e.g. GP0 and GP16 share PWM0A)
- Remember ADC is only on GP26-GP28
- All GPIOs are 3.3V, not 5V tolerant
- Suggest optimal pin groupings for common peripherals
Keep answers concise and practical. Use pin names like GP0, GP1, etc.`
}

func initAI() {
	cfg := ai.DefaultConfig()
	cfg.SystemPrompt = pingoSystemPrompt()

	// Apply saved settings to the initial AI config.
	cfg.Temperature = aiTemperature.Get()
	provider := aiProviderChoice.Get()
	if provider != "auto" {
		cfg.ProviderConfig.Provider = provider
	}

	engine = ai.NewEngine(cfg)
	assistant = ai.NewAssistant("Pingo", engine)
	chatPanel = ai.NewChatPanel(assistant)

	aiStatus.Set("loading")
	assistant.LoadAsync(func(err error) {
		if err != nil {
			aiStatus.Set("error: " + err.Error())
		} else {
			aiStatus.Set("ready: " + engine.ProviderName())
		}
	})
}

// ─── Persistent DataGrid for My Pins Page ────────────────────────────
// Created once so the widget state persists across frames (no flicker).

var myPinsGrid = ui.DataGrid(
	ui.Col("GPIO"),
	ui.Col("Function"),
	ui.Col("Category"),
).Striped(true)

// ─── Persistent Scroll List ──────────────────────────────────────────
// Must persist across frames so scroll position is retained.

var settingsScrollList = func() *giowidget.List {
	l := &giowidget.List{}
	l.Axis = layout.Vertical
	return l
}()

// ─── Board Selector Dropdown ─────────────────────────────────────────
// Persistent dropdown widget for selecting the board/chip variant.

var boardDropdown = ui.Dropdown(
	"Pico", "Pico 2", "RP2040", "RP2350A", "RP2350B",
).Placeholder("Board / Chip").OnSelect(func(i int, s string) {
	switchBoard(i)
})

// ─── Helpers ──────────────────────────────────────────────────────────

var boardList = []pindata.Board{
	pindata.Pico,
	pindata.Pico2,
	pindata.RP2040Chip,
	pindata.RP2350AChip,
	pindata.RP2350BChip,
}

func currentSpec() pindata.BoardSpec {
	idx := boardChoice.Get()
	if idx >= 0 && idx < len(boardList) {
		return pindata.GetSpec(boardList[idx])
	}
	return pindata.GetSpec(pindata.Pico)
}

func switchBoard(idx int) {
	if boardChoice.Get() != idx {
		boardChoice.Set(idx)
		selections.Set(map[int]pindata.Function{})
		activeFilter.Set("All")
		selectedPeriphFunc.Set("")
	}
}
