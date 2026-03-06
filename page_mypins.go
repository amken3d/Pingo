// ┌──────────────────────────────────────────────────────────────────────┐
// │  page_mypins.go — Selected Pins Summary Page                       │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/amken3d/Pingo/pindata"
	"github.com/amken3d/immygo/ui"
	"github.com/amken3d/immygo/widget"
)

// myPinsPage shows a summary of all selected GPIO pins,
// detects conflicts, warns about PWM sharing, and provides remove/clear actions.
func myPinsPage() ui.View {
	sel := selections.Get()

	if len(sel) == 0 {
		return ui.VStack(
			ui.Text("My Selected Pins").Headline(),
			ui.Divider(),
			ui.Text("No pins selected yet. Go to the Pinout page to assign GPIO pins for your project."),
		).Spacing(12)
	}

	// Sync the persistent grid with current selections (sorted by GPIO).
	gpios := sortedGPIOs(sel)
	rows := make([][]string, len(gpios))
	for i, gp := range gpios {
		fn := sel[gp]
		rows[i] = []string{fmt.Sprintf("GP%d", gp), fn.Name, fn.Category}
	}
	myPinsGrid.Rows(rows)

	// Conflict and PWM sharing checks.
	conflicts := pindata.CheckConflicts(sel)
	pwmWarnings := checkPWMSharing(sel)

	var sections []ui.View
	sections = append(sections,
		ui.Text("My Selected Pins").Headline(),
		ui.Text(fmt.Sprintf("%d pins selected for your project.", len(sel))).Caption(),
		ui.Divider(),
		myPinsGrid,
	)

	if len(conflicts) > 0 {
		sections = append(sections, ui.Divider(), ui.Text("Conflicts Detected").Title())
		for _, c := range conflicts {
			sections = append(sections,
				ui.HStack(
					ui.Icon(widget.IconWarning),
					ui.Text(fmt.Sprintf("GP%d: %s conflicts with %s",
						c.GPIO, c.Function1.Name, c.Function2.Name)).Bold(),
				).Spacing(8),
			)
		}
	}

	if len(pwmWarnings) > 0 {
		sections = append(sections, ui.Divider(), ui.Text("PWM Sharing Warnings").Title())
		for _, w := range pwmWarnings {
			sections = append(sections,
				ui.HStack(
					ui.Icon(widget.IconWarning),
					ui.Text(w).Caption(),
				).Spacing(8),
			)
		}
	}

	// Per-pin remove buttons (sorted).
	sections = append(sections, ui.Divider(), ui.Text("Remove Pins").Title())
	for _, gp := range gpios {
		g := gp
		fn := sel[g]
		sections = append(sections,
			ui.HStack(
				ui.Badge(fn.Category),
				ui.Text(fmt.Sprintf("GP%d — %s", g, fn.Name)),
				ui.Button("Remove").OnClick(func() {
					cur := selections.Get()
					updated := make(map[int]pindata.Function, len(cur))
					for k, v := range cur {
						if k != g {
							updated[k] = v
						}
					}
					selections.Set(updated)
				}),
			).Spacing(8),
		)
	}

	sections = append(sections,
		ui.Divider(),
		ui.HStack(
			ui.Button("Clear All").OnClick(func() {
				selections.Set(map[int]pindata.Function{})
			}),
			ui.Button("Export to Console").OnClick(func() {
				exportPinList(sel)
			}),
		).Spacing(12),
	)

	return ui.VStack(sections...).Spacing(12)
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

// exportPinList prints the selected pin configuration to stdout.
func exportPinList(sel map[int]pindata.Function) {
	fmt.Println("// Pingo — Generated Pin Configuration")
	fmt.Println("// Board:", currentSpec().Name)
	fmt.Println("//")
	for _, gp := range sortedGPIOs(sel) {
		fn := sel[gp]
		fmt.Printf("// GP%-2d -> %s (%s)\n", gp, fn.Name, fn.Category)
	}
	fmt.Println()
}
