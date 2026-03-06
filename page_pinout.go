// ┌──────────────────────────────────────────────────────────────────────┐
// │  page_pinout.go — Interactive Pinout & Pin Selector (CubeMX-style) │
// │                                                                    │
// │  The pinout page is the primary pin selection interface:            │
// │  1. Select a peripheral category (SPI, I2C, UART, etc.)           │
// │  2. Pick a specific function (e.g., SPI0 RX)                      │
// │  3. Click an eligible pin on the diagram to assign it              │
// │  4. Click an assigned pin to deassign it                           │
// │                                                                    │
// │  Eligible pins highlight in cyan, assigned pins show their         │
// │  function in the peripheral's category color.                      │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strings"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/amken3d/Pingo/pindata"
	"github.com/amken3d/immygo/theme"
	"github.com/amken3d/immygo/ui"
)

// Persistent click state for each of the 40 physical pins.
var pinClickables [40]widget.Clickable

// Board display constants.
const (
	boardDisplayW = 160 // dp width of the board image
	boardDisplayH = 440 // dp height of the board image
	labelColW     = 280 // dp width for each label column

	// Pin geometry from the SVG (in mm).
	svgTotalMM   = 52.3
	firstPinMM   = 1.37
	pinSpacingMM = 2.54
)

func pinoutPage() ui.View {
	spec := currentSpec()

	// DIP-40 layout: pins 1-20 down the left, pins 21-40 up the right.
	var leftPins, rightPins []pindata.Pin
	for _, p := range spec.Pins {
		if p.PhysicalPin <= 20 {
			leftPins = append(leftPins, p)
		} else {
			rightPins = append(rightPins, p)
		}
	}
	// Right side pins are numbered bottom-to-top (21 at bottom, 40 at top).
	// Reverse so they render top-to-bottom matching the left side rows.
	for i, j := 0, len(rightPins)-1; i < j; i, j = i+1, j-1 {
		rightPins[i], rightPins[j] = rightPins[j], rightPins[i]
	}

	pinDiagram := ui.ViewFunc(func(gtx layout.Context, th *theme.Theme) layout.Dimensions {
		return layoutPinDiagram(gtx, leftPins, rightPins)
	})

	specLine := fmt.Sprintf("%s — %s, %d cores @ %d MHz, %d KB RAM, %d KB Flash",
		spec.Chip, spec.CPUArch, spec.CPUCores, spec.MaxClockMHz, spec.RAMKB, spec.FlashKB)
	periphLine := fmt.Sprintf("SPI: %d  |  I2C: %d  |  UART: %d  |  ADC: %d  |  PWM: %d pins",
		len(pindata.PinsForCategory(spec, "SPI")),
		len(pindata.PinsForCategory(spec, "I2C")),
		len(pindata.PinsForCategory(spec, "UART")),
		len(pindata.PinsForCategory(spec, "ADC")),
		len(pindata.PinsForCategory(spec, "PWM")))
	pioLine := fmt.Sprintf("PIO: %d blocks x %d SMs  |  PWM channels: %d",
		spec.PIOBlocks, spec.PIOSMs, spec.PWMChannels)

	// AI chat panel (persistent widget from state.go).
	sel := selections.Get()
	var selCtx string
	if len(sel) > 0 {
		parts := make([]string, 0, len(sel))
		for gp, fn := range sel {
			parts = append(parts, fmt.Sprintf("GP%d->%s", gp, fn.Name))
		}
		selCtx = "Currently selected: " + strings.Join(parts, ", ")
	} else {
		selCtx = "No pins selected yet."
	}
	selectionContext := selCtx

	quickActions := ui.HStack(
		ui.Button("Best I2C pins?").OnClick(func() {
			go askAI("What are the best GPIO pins to use for I2C on this board? " + selectionContext)
		}),
		ui.Button("SPI setup?").OnClick(func() {
			go askAI("Recommend a good SPI pin configuration. " + selectionContext)
		}),
		ui.Button("Check my selection").OnClick(func() {
			go askAI("Review my pin selections for conflicts or issues: " + selectionContext)
		}),
	).Spacing(8)

	chatView := ui.ViewFunc(func(gtx layout.Context, th *theme.Theme) layout.Dimensions {
		gtx.Constraints.Min.Y = gtx.Dp(unit.Dp(300))
		return chatPanel.Layout(gtx, th)
	})

	// AI status indicator.
	status := aiStatus.Get()
	var statusView ui.View
	switch {
	case strings.HasPrefix(status, "ready:"):
		provider := strings.TrimPrefix(status, "ready: ")
		statusView = ui.Text("Connected: " + provider).Caption().Color(color.NRGBA{R: 50, G: 180, B: 50, A: 255})
	case strings.HasPrefix(status, "error:"):
		errMsg := strings.TrimPrefix(status, "error: ")
		statusView = ui.Text("Error: " + errMsg).Caption().Color(color.NRGBA{R: 220, G: 50, B: 50, A: 255})
	default:
		statusView = ui.Text("Loading AI engine...").Caption().Color(color.NRGBA{R: 180, G: 150, B: 50, A: 255})
	}

	// Right panel: assigned pins table (top-aligned, grows down) + AI chat (bottom).
	rightPanel := ui.VStack(
		selectionTable(),
		ui.Spacer(),
		ui.Divider(),
		ui.HStack(
			ui.Text("AI Pin Assistant").Title(),
			ui.Spacer(),
			statusView,
		),
		quickActions,
		chatView,
	).Spacing(4)

	return ui.VStack(
		ui.Text("Pin Header — "+spec.Name).Headline().Center(),
		ui.Text(specLine).Caption().Center(),
		ui.Text(periphLine).Caption().Center(),
		ui.Text(pioLine).Caption().Center(),
		ui.Divider(),
		peripheralSelector(spec),
		ui.Text("Select a peripheral, pick a function, then click an eligible pin. Click an assigned pin to remove it.").Caption(),
		ui.Divider(),
		ui.HStack(
			ui.Flex(3, pinDiagram),
			ui.Flex(2, rightPanel),
		),
		ui.Divider(),
		legendView(),
	).Spacing(8)
}

// peripheralSelector builds the category + function picker.
// Row 1: peripheral category buttons (toggle on/off).
// Row 2: specific function buttons for the active category.
func peripheralSelector(spec pindata.BoardSpec) ui.View {
	filter := activeFilter.Get()
	activeFn := selectedPeriphFunc.Get()

	categories := []string{"SPI", "I2C", "UART", "PWM", "ADC"}
	var catBtns []ui.View
	for _, cat := range categories {
		c := cat
		btn := ui.Button(c).OnClick(func() {
			if activeFilter.Get() == c {
				activeFilter.Set("")
				selectedPeriphFunc.Set("")
			} else {
				activeFilter.Set(c)
				selectedPeriphFunc.Set("")
			}
		})
		if c == filter {
			catBtns = append(catBtns, ui.Style(btn).Background(categoryToColor(c)))
		} else {
			catBtns = append(catBtns, btn)
		}
	}

	views := []ui.View{
		ui.HStack(catBtns...).Spacing(4),
	}

	if filter != "" {
		funcs := functionsForCategory(spec, filter)
		var funcBtns []ui.View
		for _, fn := range funcs {
			f := fn
			btn := ui.Button(f).OnClick(func() {
				if selectedPeriphFunc.Get() == f {
					selectedPeriphFunc.Set("")
				} else {
					selectedPeriphFunc.Set(f)
				}
			})
			if f == activeFn {
				funcBtns = append(funcBtns, ui.Style(btn).Background(categoryToColor(filter)))
			} else {
				funcBtns = append(funcBtns, btn.Outline())
			}
		}
		views = append(views, ui.HStack(funcBtns...).Spacing(4))
	}

	return ui.VStack(views...).Spacing(4)
}

// selectionTable shows assigned pins in a table on the right side of the diagram.
func selectionTable() ui.View {
	sel := selections.Get()
	if len(sel) == 0 {
		return ui.Centered(
			ui.VStack(
				ui.Text("Assigned Pins").Title(),
				ui.Text("No pins assigned yet.").Caption(),
			).Spacing(4),
		)
	}

	// Sort by GPIO number for stable row order across frames.
	gpios := make([]int, 0, len(sel))
	for gp := range sel {
		gpios = append(gpios, gp)
	}
	sort.Ints(gpios)

	rows := make([]ui.View, 0, len(gpios))
	for _, gp := range gpios {
		fn := sel[gp]
		rows = append(rows,
			ui.HStack(
				ui.Badge(fn.Category),
				ui.Text(fmt.Sprintf("GP%d", gp)).Bold(),
				ui.Text(fn.Name).Caption(),
			).Spacing(8),
		)
	}

	items := []ui.View{
		ui.Text(fmt.Sprintf("Assigned Pins (%d)", len(sel))).Title(),
		ui.Divider(),
	}
	items = append(items, rows...)
	items = append(items,
		ui.Divider(),
		ui.Button("Clear All").OnClick(func() {
			selections.Set(map[int]pindata.Function{})
		}),
	)

	return ui.Style(ui.VStack(items...).Spacing(4)).Padding(8)
}

// functionsForCategory returns sorted unique function names for a category.
func functionsForCategory(spec pindata.BoardSpec, category string) []string {
	seen := map[string]bool{}
	var names []string
	for _, p := range spec.Pins {
		if !p.IsGPIO {
			continue
		}
		for _, f := range p.Functions {
			if f.Category == category && !seen[f.Name] {
				seen[f.Name] = true
				names = append(names, f.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}

// handlePinClick processes a click on a GPIO pin.
// If the pin is already assigned, it deassigns it.
// If a peripheral function is active and the pin supports it, it assigns.
func handlePinClick(p pindata.Pin) {
	sel := selections.Get()
	activeFn := selectedPeriphFunc.Get()

	// Already assigned → deassign (toggle off).
	if _, ok := sel[p.GPIO]; ok {
		updated := make(map[int]pindata.Function, len(sel))
		for k, v := range sel {
			if k != p.GPIO {
				updated[k] = v
			}
		}
		selections.Set(updated)
		return
	}

	if activeFn == "" {
		return
	}

	// Assign if pin supports the active function.
	// Only one pin can hold a given function — remove any existing assignment.
	for _, f := range p.Functions {
		if f.Name == activeFn {
			updated := make(map[int]pindata.Function, len(sel)+1)
			for k, v := range sel {
				if v.Name != activeFn {
					updated[k] = v
				}
			}
			updated[p.GPIO] = f
			selections.Set(updated)
			return
		}
	}
}

// layoutPinDiagram renders the board image with clickable pin labels.
// Label columns adapt to the constraint width provided by the parent Flex.
func layoutPinDiagram(gtx layout.Context, leftPins, rightPins []pindata.Pin) layout.Dimensions {
	boardW := gtx.Dp(boardDisplayW)
	boardH := gtx.Dp(boardDisplayH)
	gap := gtx.Dp(4)
	labelH := gtx.Dp(16)

	// Use the constraint width from the Flex parent for label sizing.
	availW := gtx.Constraints.Max.X
	labelW := (availW - boardW) / 2
	maxLabelW := gtx.Dp(labelColW)
	if labelW > maxLabelW {
		labelW = maxLabelW
	}
	if labelW < gtx.Dp(80) {
		labelW = gtx.Dp(80)
	}
	totalW := labelW + boardW + labelW

	firstPinY := float64(boardH) * (firstPinMM / svgTotalMM)
	pinSpacing := float64(boardH) * (pinSpacingMM / svgTotalMM)

	// ── Draw the board image ────────────────────────────────────────
	if picoImage != nil {
		imgBounds := picoImage.Bounds()
		scaleX := float32(boardW) / float32(imgBounds.Dx())
		scaleY := float32(boardH) / float32(imgBounds.Dy())

		offStack := op.Offset(image.Pt(labelW, 0)).Push(gtx.Ops)
		affStack := op.Affine(f32.NewAffine2D(scaleX, 0, 0, 0, scaleY, 0)).Push(gtx.Ops)

		imgOp := paint.NewImageOp(picoImage)
		imgOp.Add(gtx.Ops)
		clipStack := clip.Rect(imgBounds).Push(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		clipStack.Pop()

		affStack.Pop()
		offStack.Pop()
	}

	matTh := material.NewTheme()

	// ── Left labels (odd physical pins: 1, 3, 5, ..., 39) ──────────
	for i, p := range leftPins {
		pinY := int(firstPinY+float64(i)*pinSpacing) - gtx.Dp(7)
		physIdx := p.PhysicalPin - 1

		offStack := op.Offset(image.Pt(0, pinY)).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Exact(image.Pt(labelW-gap, labelH))

		if p.IsGPIO {
			for pinClickables[physIdx].Clicked(cgtx) {
				handlePinClick(p)
			}
			pinClickables[physIdx].Layout(cgtx, func(gtx layout.Context) layout.Dimensions {
				return makePinLabel(matTh, interactivePinText(p), text.End, interactivePinColor(p)).Layout(gtx)
			})
		} else {
			makePinLabel(matTh, formatPinLabel(p), text.End, colorForPin(p)).Layout(cgtx)
		}

		offStack.Pop()
	}

	// ── Right labels (even physical pins: 2, 4, 6, ..., 40) ────────
	for i, p := range rightPins {
		pinY := int(firstPinY+float64(i)*pinSpacing) - gtx.Dp(7)
		physIdx := p.PhysicalPin - 1

		offStack := op.Offset(image.Pt(labelW+boardW+gap, pinY)).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Exact(image.Pt(labelW-gap, labelH))

		if p.IsGPIO {
			for pinClickables[physIdx].Clicked(cgtx) {
				handlePinClick(p)
			}
			pinClickables[physIdx].Layout(cgtx, func(gtx layout.Context) layout.Dimensions {
				return makePinLabel(matTh, interactivePinText(p), text.Start, interactivePinColor(p)).Layout(gtx)
			})
		} else {
			makePinLabel(matTh, formatPinLabel(p), text.Start, colorForPin(p)).Layout(cgtx)
		}

		offStack.Pop()
	}

	// ── Connector lines ─────────────────────────────────────────────
	lineColor := color.NRGBA{R: 160, G: 160, B: 160, A: 180}
	for i := range leftPins {
		py := int(firstPinY + float64(i)*pinSpacing)
		drawHLine(gtx.Ops, labelW-gap, py, labelW, lineColor)
	}
	for i := range rightPins {
		py := int(firstPinY + float64(i)*pinSpacing)
		drawHLine(gtx.Ops, labelW+boardW, py, labelW+boardW+gap, lineColor)
	}

	// ── Pin numbers on the board (white text on dark green) ────────
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	pinNumW := gtx.Dp(28)

	// Left side pin numbers — right-aligned near the left edge of the board
	for i, p := range leftPins {
		pinY := int(firstPinY+float64(i)*pinSpacing) - gtx.Dp(7)
		offStack := op.Offset(image.Pt(labelW+gtx.Dp(2), pinY)).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Exact(image.Pt(pinNumW, labelH))
		makePinLabel(matTh, fmt.Sprintf("%d", p.PhysicalPin), text.Start, white).Layout(cgtx)
		offStack.Pop()
	}

	// Right side pin numbers — left-aligned near the right edge of the board
	for i, p := range rightPins {
		pinY := int(firstPinY+float64(i)*pinSpacing) - gtx.Dp(7)
		offStack := op.Offset(image.Pt(labelW+boardW-pinNumW-gtx.Dp(2), pinY)).Push(gtx.Ops)
		cgtx := gtx
		cgtx.Constraints = layout.Exact(image.Pt(pinNumW, labelH))
		makePinLabel(matTh, fmt.Sprintf("%d", p.PhysicalPin), text.End, white).Layout(cgtx)
		offStack.Pop()
	}

	return layout.Dimensions{Size: image.Pt(totalW, boardH)}
}

// makePinLabel creates a styled material label for a pin.
func makePinLabel(matTh *material.Theme, labelText string, align text.Alignment, c color.NRGBA) material.LabelStyle {
	lbl := material.Body2(matTh, labelText)
	lbl.Font.Weight = font.Bold
	lbl.Alignment = align
	lbl.MaxLines = 1
	lbl.Color = c
	return lbl
}

// interactivePinText returns the label text for a GPIO pin, reflecting assignments.
func interactivePinText(p pindata.Pin) string {
	sel := selections.Get()
	if fn, ok := sel[p.GPIO]; ok {
		return fmt.Sprintf("%s > %s", p.Label, fn.Name)
	}
	return formatPinLabel(p)
}

// interactivePinColor returns the label color for a GPIO pin based on
// selection state and the active peripheral function.
func interactivePinColor(p pindata.Pin) color.NRGBA {
	sel := selections.Get()
	activeFn := selectedPeriphFunc.Get()

	// Assigned: use category color.
	if fn, ok := sel[p.GPIO]; ok {
		hex := pindata.CategoryColor(fn.Category)
		return color.NRGBA{
			R: uint8((hex >> 16) & 0xFF),
			G: uint8((hex >> 8) & 0xFF),
			B: uint8(hex & 0xFF),
			A: 255,
		}
	}

	// Selection mode: highlight eligible pins, dim the rest.
	if activeFn != "" {
		for _, f := range p.Functions {
			if f.Name == activeFn {
				return color.NRGBA{R: 0, G: 220, B: 255, A: 255} // cyan
			}
		}
		return color.NRGBA{R: 120, G: 120, B: 120, A: 100} // dimmed
	}

	return colorForPin(p)
}

// drawHLine draws a 1px horizontal line.
func drawHLine(ops *op.Ops, x1, y, x2 int, c color.NRGBA) {
	stack := clip.Rect(image.Rect(x1, y, x2, y+1)).Push(ops)
	paint.ColorOp{Color: c}.Add(ops)
	paint.PaintOp{}.Add(ops)
	stack.Pop()
}

// colorForPin returns a label color based on pin type (non-interactive).
func colorForPin(p pindata.Pin) color.NRGBA {
	switch {
	case p.IsGround:
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255}
	case p.IsPower:
		return color.NRGBA{R: 220, G: 50, B: 50, A: 255}
	case p.IsSpecial:
		return color.NRGBA{R: 200, G: 150, B: 0, A: 255}
	default:
		return color.NRGBA{R: 50, G: 150, B: 50, A: 255}
	}
}

// formatPinLabel creates a concise label for a pin (non-GPIO or default).
func formatPinLabel(p pindata.Pin) string {
	if p.IsGround {
		return "GND"
	}
	if p.IsPower || p.IsSpecial || !p.IsGPIO {
		return p.Label
	}
	var funcs []string
	for _, f := range p.Functions {
		if f.Category != "GPIO" && f.Category != "PIO" {
			funcs = append(funcs, f.Name)
		}
	}
	if len(funcs) > 3 {
		funcs = funcs[:3]
	}
	if len(funcs) > 0 {
		return fmt.Sprintf("%s (%s)", p.Label, strings.Join(funcs, ", "))
	}
	return p.Label
}

// askAI sends a question through the chat panel so the user sees both the
// question and the response (or error) in the conversation history.
func askAI(question string) {
	spec := currentSpec()
	fullQ := fmt.Sprintf("[Board: %s] %s", spec.Name, question)
	chatPanel.SendMessage(fullQ)
}

// categoryToColor converts a peripheral category name to a color.
func categoryToColor(cat string) color.NRGBA {
	hex := pindata.CategoryColor(cat)
	return color.NRGBA{
		R: uint8((hex >> 16) & 0xFF),
		G: uint8((hex >> 8) & 0xFF),
		B: uint8(hex & 0xFF),
		A: 255,
	}
}

// legendView shows the color key for pin types.
func legendView() ui.View {
	return ui.HStack(
		ui.Badge("GPIO").Success(), ui.Text("User GPIO"),
		ui.FixedSpacer(12, 0),
		ui.Badge("GND").Secondary(), ui.Text("Ground"),
		ui.FixedSpacer(12, 0),
		ui.Badge("PWR").Danger(), ui.Text("Power"),
		ui.FixedSpacer(12, 0),
		ui.Badge("SPC").Warning(), ui.Text("Special"),
	).Spacing(4)
}
