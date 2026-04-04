package shared

import (
	"encoding/json"
	"os"
)

var DEFAULTS = RunnerConfig{
	MaxConcurrent:       DEFAULT_MAX_CONCURRENT,
	DefaultTimeout:      DEFAULT_TIMEOUT,
	DefaultApprovalMode: DEFAULT_APPROVAL_MODE,
	DefaultBackend:      DEFAULT_BACKEND,
}

func LoadConfig() RunnerConfig {
	if _, err := os.Stat(CONFIG_PATH); os.IsNotExist(err) {
		return DEFAULTS
	}

	data, err := os.ReadFile(CONFIG_PATH)
	if err != nil {
		return DEFAULTS
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return DEFAULTS
	}

	cfg := DEFAULTS

	if v, ok := raw["maxConcurrent"].(float64); ok {
		cfg.MaxConcurrent = int(v)
	}
	if v, ok := raw["defaultTimeout"].(float64); ok {
		cfg.DefaultTimeout = int(v)
	}
	if v, ok := raw["defaultApprovalMode"].(string); ok {
		cfg.DefaultApprovalMode = ApprovalMode(v)
	}
	if v, ok := raw["defaultBackend"].(string); ok {
		cfg.DefaultBackend = Backend(v)
	}
	if v, ok := raw["defaultThinking"].(bool); ok {
		cfg.DefaultThinking = v
	}

	return cfg
}
