// ┌──────────────────────────────────────────────────────────────────────┐
// │  ai_suggest.go — Parse & Apply AI Pin Suggestions                   │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/amken3d/Pingo/pindata"
	"github.com/amken3d/immygo/ai"
)

// pinSuggestion represents a single AI-suggested pin assignment.
type pinSuggestion struct {
	GPIO     int
	Function string // e.g. "SPI0 RX"
}

// parseSuggestions extracts pin assignments from an AI response.
// It looks for structured "PIN: GPx -> function" lines first,
// then falls back to common patterns like "GP0 for SPI0 RX".
func parseSuggestions(text string) []pinSuggestion {
	var suggestions []pinSuggestion
	seen := map[int]bool{}

	// Pattern 1: Structured format "PIN: GP0 -> SPI0 RX"
	structured := regexp.MustCompile(`(?i)PIN:\s*GP(\d+)\s*[-=~>]+\s*(.+)`)
	for _, m := range structured.FindAllStringSubmatch(text, -1) {
		gpio, _ := strconv.Atoi(m[1])
		fn := strings.TrimSpace(m[2])
		if !seen[gpio] && fn != "" {
			suggestions = append(suggestions, pinSuggestion{GPIO: gpio, Function: fn})
			seen[gpio] = true
		}
	}

	if len(suggestions) > 0 {
		return suggestions
	}

	// Pattern 2: Fallback — "GP0 for/as/: SPI0 RX" or "GP0 → SPI0 RX"
	fallback := regexp.MustCompile(`GP(\d+)\s*(?:for|as|:|->|→)\s+([A-Z][A-Z0-9]*\s+[A-Z][A-Za-z0-9]*)`)
	for _, m := range fallback.FindAllStringSubmatch(text, -1) {
		gpio, _ := strconv.Atoi(m[1])
		fn := strings.TrimSpace(m[2])
		if !seen[gpio] && fn != "" {
			suggestions = append(suggestions, pinSuggestion{GPIO: gpio, Function: fn})
			seen[gpio] = true
		}
	}

	return suggestions
}

// lastAssistantMessage returns the content of the most recent AI response.
func lastAssistantMessage() string {
	msgs := chatPanel.Messages
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == ai.RoleAssistant {
			return msgs[i].Content
		}
	}
	return ""
}

// applySuggestions parses the last AI response and applies valid pin assignments.
// Returns a human-readable summary of what was applied.
func applySuggestions() string {
	text := lastAssistantMessage()
	if text == "" {
		return "No AI response to parse."
	}

	suggestions := parseSuggestions(text)
	if len(suggestions) == 0 {
		return "No pin suggestions found in the last AI response."
	}

	spec := currentSpec()
	sel := selections.Get()
	updated := make(map[int]pindata.Function, len(sel)+len(suggestions))
	for k, v := range sel {
		updated[k] = v
	}

	var applied []string
	var skipped []string

	for _, s := range suggestions {
		// Find the matching function on this board.
		matched := false
		for _, pin := range spec.Pins {
			if pin.GPIO != s.GPIO || !pin.IsGPIO {
				continue
			}
			for _, fn := range pin.Functions {
				if strings.EqualFold(fn.Name, s.Function) {
					// Remove any existing assignment of this function.
					for k, v := range updated {
						if v.Name == fn.Name {
							delete(updated, k)
						}
					}
					updated[s.GPIO] = fn
					applied = append(applied, fmt.Sprintf("GP%d -> %s", s.GPIO, fn.Name))
					matched = true
					break
				}
			}
			break
		}
		if !matched {
			skipped = append(skipped, fmt.Sprintf("GP%d -> %s", s.GPIO, s.Function))
		}
	}

	if len(applied) == 0 {
		return fmt.Sprintf("Could not match any suggestions to this board. Skipped: %s",
			strings.Join(skipped, ", "))
	}

	selections.Set(updated)

	result := fmt.Sprintf("Applied %d pins: %s", len(applied), strings.Join(applied, ", "))
	if len(skipped) > 0 {
		result += fmt.Sprintf(" | Skipped: %s", strings.Join(skipped, ", "))
	}
	return result
}
