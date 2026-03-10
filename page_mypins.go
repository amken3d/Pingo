// ┌──────────────────────────────────────────────────────────────────────┐
// │  page_mypins.go — Selected Pins Summary Page                       │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"gioui.org/layout"
	"gioui.org/widget/material"

	"github.com/amken3d/Pingo/pindata"
	"github.com/amken3d/immygo/theme"
	"github.com/amken3d/immygo/ui"
	"github.com/amken3d/immygo/widget"
)

// myPinsPage shows a summary of all selected GPIO pins,
// detects conflicts, warns about PWM sharing, and provides rename/remove/export actions.
func myPinsPage() ui.View {
	sel := selections.Get()

	if len(sel) == 0 {
		return ui.VStack(
			ui.Text("My Selected Pins").Headline(),
			ui.Divider(),
			ui.Text("No pins selected yet. Go to the Pinout page to assign GPIO pins for your project."),
		).Spacing(12)
	}

	gpios := sortedGPIOs(sel)

	// Conflict and PWM sharing checks.
	conflicts := pindata.CheckConflicts(sel)
	pwmWarnings := checkPWMSharing(sel)

	var sections []ui.View
	sections = append(sections,
		ui.Text("My Selected Pins").Headline(),
		ui.Text(fmt.Sprintf("%d pins selected for your project.", len(sel))).Caption(),
		ui.Divider(),
	)

	// Each pin: badge, label, function, inline name editor, remove — all in one row.
	for _, gp := range gpios {
		g := gp
		fn := sel[g]

		sections = append(sections,
			ui.HStack(
				ui.Badge(fn.Category),
				ui.Text(fmt.Sprintf("GP%d", g)).Bold(),
				ui.Text(fn.Name).Caption(),
				ui.Flex(1, ui.Style(
					pinNameEditor(g),
				).Background(editorBg()).Padding(4).Rounded(4)),
				ui.Button("Remove").OnClick(func() {
					removePin(g)
				}),
			).Spacing(8),
			ui.Divider(),
		)
	}

	if len(conflicts) > 0 {
		sections = append(sections, ui.Text("Conflicts Detected").Title())
		for _, c := range conflicts {
			sections = append(sections,
				ui.HStack(
					ui.Icon(widget.IconWarning),
					ui.Text(fmt.Sprintf("GP%d: %s conflicts with %s",
						c.GPIO, c.Function1.Name, c.Function2.Name)).Bold(),
				).Spacing(8),
			)
		}
		sections = append(sections, ui.Divider())
	}

	if len(pwmWarnings) > 0 {
		sections = append(sections, ui.Text("PWM Sharing Warnings").Title())
		for _, w := range pwmWarnings {
			sections = append(sections,
				ui.HStack(
					ui.Icon(widget.IconWarning),
					ui.Text(w).Caption(),
				).Spacing(8),
			)
		}
		sections = append(sections, ui.Divider())
	}

	// ── Export section ───────────────────────────────────────────────
	status := exportStatus.Get()
	var statusView ui.View
	if status != "" {
		statusView = ui.Text(status).Caption()
	} else {
		statusView = ui.Text("Choose a format to export your pin configuration.").Caption()
	}

	sections = append(sections,
		ui.Text("Export Pin Configuration").Title(),
		statusView,
		ui.HStack(
			ui.Button("C/C++ Header").OnClick(func() {
				names := customNames.Get()
				saveExport("pingo_pins.h", "C/C++ Header", "*.h", exportCHeader(sel, names))
			}),
			ui.Button("TinyGo").OnClick(func() {
				names := customNames.Get()
				saveExport("pingo_pins.go", "Go Source", "*.go", exportTinyGo(sel, names))
			}),
			ui.Button("MicroPython").OnClick(func() {
				names := customNames.Get()
				saveExport("pingo_pins.py", "Python", "*.py", exportMicroPython(sel, names))
			}),
			ui.Button("CSV").OnClick(func() {
				names := customNames.Get()
				saveExport("pingo_pins.csv", "CSV", "*.csv", exportCSV(sel, names))
			}),
		).Spacing(8),
		ui.Divider(),
		ui.Button("Clear All").OnClick(func() {
			clearAllSelections()
		}),
	)

	return ui.ScrollPersistent(myPinsScrollList, ui.VStack(sections...).Spacing(8))
}

// pinNameEditor returns a ViewFunc that renders an inline Gio editor for a pin's custom name.
func editorBg() color.NRGBA {
	if isDark.Get() {
		return color.NRGBA{R: 50, G: 50, B: 50, A: 255}
	}
	return color.NRGBA{R: 240, G: 240, B: 240, A: 255}
}

func pinNameEditor(gpio int) ui.View {
	return ui.ViewFunc(func(gtx layout.Context, th *theme.Theme) layout.Dimensions {
		ed := &pinNameEditors[gpio]

		// Render the editor.
		matTh := material.NewTheme()
		style := material.Editor(matTh, ed, "Custom name...")
		style.Color = th.Palette.OnBackground
		style.HintColor = color.NRGBA{R: 160, G: 160, B: 160, A: 255}
		dims := style.Layout(gtx)

		// Sync editor → state.
		currentNames := customNames.Get()
		currentName := currentNames[gpio]
		if ed.Text() != currentName {
			updated := make(map[int]string, len(currentNames)+1)
			for k, v := range currentNames {
				updated[k] = v
			}
			if t := ed.Text(); t != "" {
				updated[gpio] = t
			} else {
				delete(updated, gpio)
			}
			customNames.Set(updated)
		}

		return dims
	})
}

// sortedGPIOs returns GPIO numbers from the selection map in ascending order.
func sortedGPIOs(sel map[int]pindata.Function) []int {
	gpios := make([]int, 0, len(sel))
	for gp := range sel {
		gpios = append(gpios, gp)
	}
	sort.Ints(gpios)
	return gpios
}

// checkPWMSharing detects GPIOs that share the same PWM slice+channel.
func checkPWMSharing(sel map[int]pindata.Function) []string {
	spec := currentSpec()
	type pwmKey struct {
		slice   int
		channel string
	}
	pwmUsage := make(map[pwmKey][]int)
	for gp := range sel {
		for _, p := range spec.Pins {
			if p.GPIO == gp && p.PWMSlice >= 0 {
				key := pwmKey{p.PWMSlice, p.PWMChannel}
				pwmUsage[key] = append(pwmUsage[key], gp)
			}
		}
	}

	var warnings []string
	for key, gpios := range pwmUsage {
		if len(gpios) > 1 {
			labels := make([]string, len(gpios))
			for i, g := range gpios {
				labels[i] = fmt.Sprintf("GP%d", g)
			}
			warnings = append(warnings,
				fmt.Sprintf("PWM%d%s is shared by %s — they will output the same signal.",
					key.slice, key.channel, strings.Join(labels, " and ")))
		}
	}
	return warnings
}
