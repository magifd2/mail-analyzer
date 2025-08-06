package config

import (
	"encoding/json"
	"os"

	"github.com/kelseyhightower/envconfig"
)

// Config holds the application configuration.
type Config struct {
	OpenAIAPIKey  string `json:"openai_api_key" envconfig:"OPENAI_API_KEY"`
	OpenAIBaseURL string `json:"openai_base_url" envconfig:"OPENAI_BASE_URL"`
	ModelName     string `json:"model_name" envconfig:"MODEL_NAME"`
}

// Load loads configuration from a file, then overrides with environment variables.
func Load(path string) (*Config, error) {
	var cfg Config

	// Load from file first.
	if path != "" {
		file, err := os.Open(path)
		if err != nil {
			// Ignore file not found errors, as the path may not always exist.
			if !os.IsNotExist(err) {
				return nil, err
			}
		} else {
			defer file.Close()
			if err := json.NewDecoder(file).Decode(&cfg); err != nil {
				return nil, err
			}
		}
	}

	// Now, process environment variables. This will override any fields with values
	// from the environment.
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	// Manually set default for ModelName if it's still empty.
	if cfg.ModelName == "" {
		cfg.ModelName = "gpt-4-turbo"
	}

	return &cfg, nil
}
