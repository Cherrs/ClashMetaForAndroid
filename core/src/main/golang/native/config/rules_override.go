package config

import (
	"encoding/json"
	"strings"

	M "github.com/metacubex/mihomo/config"
)

type rawOverride struct {
	*M.RawConfig
	PrependRules []string `json:"prepend-rules"`
}

func applyOverrideContent(cfg *M.RawConfig, content string) ([]string, error) {
	override := rawOverride{RawConfig: cfg}
	err := json.NewDecoder(strings.NewReader(content)).Decode(&override)

	return override.PrependRules, err
}

func prependOverrideRules(cfg *M.RawConfig, persistRules, sessionRules []string) {
	// Keep the feature-disabled path identical to the original rule pipeline:
	// no new slice and no additional rules in the runtime matcher.
	if len(persistRules) == 0 && len(sessionRules) == 0 {
		return
	}

	rules := make([]string, 0, len(sessionRules)+len(persistRules)+len(cfg.Rule))
	rules = append(rules, sessionRules...)
	rules = append(rules, persistRules...)
	cfg.Rule = append(rules, cfg.Rule...)
}
