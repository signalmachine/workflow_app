package ai

import (
	"errors"
	"testing"
)

func TestLoadProviderConfig(t *testing.T) {
	t.Run("disabled when env absent", func(t *testing.T) {
		config, err := LoadProviderConfig(func(string) (string, bool) {
			return "", false
		})
		if err != nil {
			t.Fatalf("LoadProviderConfig() error = %v", err)
		}
		if config.Enabled() {
			t.Fatal("expected provider config to be disabled")
		}
	})

	t.Run("loads trimmed openai settings", func(t *testing.T) {
		config, err := LoadProviderConfig(func(key string) (string, bool) {
			switch key {
			case "OPENAI_API_KEY":
				return "  sk-test  ", true
			case "OPENAI_MODEL":
				return "  gpt-4.1-mini  ", true
			default:
				return "", false
			}
		})
		if err != nil {
			t.Fatalf("LoadProviderConfig() error = %v", err)
		}
		if !config.Enabled() {
			t.Fatal("expected provider config to be enabled")
		}
		if config.OpenAIAPIKey != "sk-test" {
			t.Fatalf("unexpected api key: %q", config.OpenAIAPIKey)
		}
		if config.OpenAIModel != "gpt-4.1-mini" {
			t.Fatalf("unexpected model: %q", config.OpenAIModel)
		}
	})

	t.Run("rejects key without model", func(t *testing.T) {
		_, err := LoadProviderConfig(func(key string) (string, bool) {
			if key == "OPENAI_API_KEY" {
				return "sk-test", true
			}
			return "", false
		})
		if !errors.Is(err, ErrInvalidProviderConfig) {
			t.Fatalf("expected ErrInvalidProviderConfig, got %v", err)
		}
	})

	t.Run("rejects model without key", func(t *testing.T) {
		_, err := LoadProviderConfig(func(key string) (string, bool) {
			if key == "OPENAI_MODEL" {
				return "gpt-4.1-mini", true
			}
			return "", false
		})
		if !errors.Is(err, ErrInvalidProviderConfig) {
			t.Fatalf("expected ErrInvalidProviderConfig, got %v", err)
		}
	})
}
