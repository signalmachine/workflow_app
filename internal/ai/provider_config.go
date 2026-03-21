package ai

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrInvalidProviderConfig = errors.New("invalid ai provider config")

type ProviderConfig struct {
	OpenAIAPIKey string
	OpenAIModel  string
}

func LoadProviderConfigFromEnv() (ProviderConfig, error) {
	return LoadProviderConfig(os.LookupEnv)
}

func LoadProviderConfig(lookup func(string) (string, bool)) (ProviderConfig, error) {
	config := ProviderConfig{
		OpenAIAPIKey: lookupEnvTrimmed(lookup, "OPENAI_API_KEY"),
		OpenAIModel:  lookupEnvTrimmed(lookup, "OPENAI_MODEL"),
	}

	if config.OpenAIAPIKey == "" && config.OpenAIModel == "" {
		return config, nil
	}
	if config.OpenAIAPIKey == "" {
		return ProviderConfig{}, fmt.Errorf("%w: OPENAI_MODEL requires OPENAI_API_KEY", ErrInvalidProviderConfig)
	}
	if config.OpenAIModel == "" {
		return ProviderConfig{}, fmt.Errorf("%w: OPENAI_API_KEY requires OPENAI_MODEL", ErrInvalidProviderConfig)
	}

	return config, nil
}

func (c ProviderConfig) Enabled() bool {
	return c.OpenAIAPIKey != "" && c.OpenAIModel != ""
}

func lookupEnvTrimmed(lookup func(string) (string, bool), key string) string {
	value, ok := lookup(key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}
