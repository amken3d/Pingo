// ┌──────────────────────────────────────────────────────────────────────┐
// │  page_settings.go — AI & App Settings Page                         │
// │                                                                    │
// │  Non-secret settings (provider, model, host, temperature, tokens)  │
// │  are persisted to ~/.config/pingo/settings.json.                   │
// │  API keys are NEVER written to disk — use the ANTHROPIC_API_KEY    │
// │  environment variable or enter the key in the masked input         │
// │  (held in memory only for the current session).                    │
// └──────────────────────────────────────────────────────────────────────┘
package main

import (
	"fmt"
	"os"

	"github.com/amken3d/immygo/ai"
	"github.com/amken3d/immygo/ui"
)

// Persistent settings widgets — must live at package level.
var (
	providerDropdown = ui.Dropdown("auto", "ollama", "anthropic", "yzma", "simulation").
				Placeholder("Provider").
				OnSelect(func(i int, s string) {
			aiProviderChoice.Set(s)
		})

	modelInput = ui.Input().Placeholder("Model name (e.g. qwen2.5-coder)")

	ollamaHostInput = ui.Input().Placeholder("Ollama host (default: http://localhost:11434)")

	apiKeyInput = ui.Password().Placeholder("API Key (session only, not saved to disk)")

	temperatureSlider = ui.Slider(0, 1, 0.7).OnChange(func(v float32) {
		aiTemperature.Set(v)
	})

	maxTokensInput = ui.Input().Placeholder("Max tokens (default: 256)")
)

// loadSettingsIntoUI populates the UI widgets from the persisted config.
func loadSettingsIntoUI() {
	cfg := loadConfig()

	aiProviderChoice.Set(cfg.Provider)
	aiTemperature.Set(cfg.Temperature)

	// Set dropdown to match saved provider
	providers := []string{"auto", "ollama", "anthropic", "yzma", "simulation"}
	for i, p := range providers {
		if p == cfg.Provider {
			providerDropdown.SetSelected(i)
			break
		}
	}

	if cfg.Model != "" {
		modelInput.SetValue(cfg.Model)
	}
	if cfg.OllamaHost != "" {
		ollamaHostInput.SetValue(cfg.OllamaHost)
	}
	if cfg.MaxTokens > 0 {
		maxTokensInput.SetValue(fmt.Sprintf("%d", cfg.MaxTokens))
	}
}

func settingsPage() ui.View {
	temp := aiTemperature.Get()

	envKeySet := os.Getenv("ANTHROPIC_API_KEY") != ""
	keyCaption := "In-memory only — never saved to disk. Leave blank to use ANTHROPIC_API_KEY env var."
	if envKeySet {
		keyCaption = "ANTHROPIC_API_KEY env var is set. Override below for this session only (not saved)."
	}

	return ui.ScrollPersistent(settingsScrollList, ui.VStack(
		ui.Text("Settings").Headline(),
		ui.Divider(),

		// AI Provider
		ui.Text("AI Provider").Title(),
		ui.Text("Select the LLM backend. 'auto' detects available providers.").Caption(),
		providerDropdown,

		ui.Divider(),

		// Model
		ui.Text("Model").Title(),
		ui.Text("Model name for Ollama or Anthropic (e.g. qwen2.5-coder, claude-sonnet-4-20250514).").Caption(),
		modelInput,

		ui.Divider(),

		// Ollama Host
		ui.Text("Ollama Host").Title(),
		ui.Text("API base URL for Ollama server.").Caption(),
		ollamaHostInput,

		ui.Divider(),

		// API Key
		ui.Text("Anthropic API Key").Title(),
		ui.Text(keyCaption).Caption(),
		apiKeyInput,

		ui.Divider(),

		// Temperature
		ui.Text("Temperature").Title(),
		ui.Text(fmt.Sprintf("Controls randomness: 0.0 = deterministic, 1.0 = creative. Current: %.2f", temp)).Caption(),
		temperatureSlider,

		ui.Divider(),

		// Max Tokens
		ui.Text("Max Tokens").Title(),
		ui.Text("Maximum tokens per AI response.").Caption(),
		maxTokensInput,

		ui.Divider(),

		// Apply button
		ui.Button("Apply & Reinitialize AI").OnClick(func() {
			applyAISettings()
		}),
	).Spacing(8))
}

func applyAISettings() {
	cfg := ai.DefaultConfig()
	cfg.SystemPrompt = pingoSystemPrompt()

	cfg.Temperature = aiTemperature.Get()

	provider := aiProviderChoice.Get()
	if provider != "auto" {
		cfg.ProviderConfig.Provider = provider
	}

	if m := modelInput.Value(); m != "" {
		cfg.ProviderConfig.Model = m
	}

	if h := ollamaHostInput.Value(); h != "" {
		cfg.ProviderConfig.OllamaHost = h
	}

	// API key: use input value if provided, otherwise fall back to env var
	// (the ai package reads ANTHROPIC_API_KEY automatically if AnthropicKey is empty)
	if k := apiKeyInput.Value(); k != "" {
		cfg.ProviderConfig.AnthropicKey = k
	}

	var maxTokens int
	if t := maxTokensInput.Value(); t != "" {
		if _, err := fmt.Sscanf(t, "%d", &maxTokens); err == nil && maxTokens > 0 {
			cfg.MaxTokens = maxTokens
		}
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

	// Persist non-secret settings to disk
	appCfg := AppConfig{
		Provider:    provider,
		Model:       modelInput.Value(),
		OllamaHost:  ollamaHostInput.Value(),
		Temperature: aiTemperature.Get(),
		MaxTokens:   maxTokens,
	}
	_ = saveConfig(appCfg)
}
