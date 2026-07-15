package config

import (
	"reflect"
	"testing"

	M "github.com/metacubex/mihomo/config"
	// Register the linkname implementation required by config.ParseRawConfig.
	_ "github.com/metacubex/mihomo/hub/executor"
)

func TestApplyOverrideContentKeepsSubscriptionRules(t *testing.T) {
	cfg := &M.RawConfig{
		MixedPort: 7890,
		Rule:      []string{"MATCH,Subscription"},
	}

	rules, err := applyOverrideContent(cfg, `{
		"mixed-port": 9090,
		"prepend-rules": ["DOMAIN,example.com,DIRECT"]
	}`)
	if err != nil {
		t.Fatalf("apply override: %v", err)
	}

	if cfg.MixedPort != 9090 {
		t.Fatalf("mixed port = %d, want 9090", cfg.MixedPort)
	}
	if want := []string{"MATCH,Subscription"}; !reflect.DeepEqual(cfg.Rule, want) {
		t.Fatalf("subscription rules = %#v, want %#v", cfg.Rule, want)
	}
	if want := []string{"DOMAIN,example.com,DIRECT"}; !reflect.DeepEqual(rules, want) {
		t.Fatalf("prepend rules = %#v, want %#v", rules, want)
	}
}

func TestPrependOverrideRulesPriority(t *testing.T) {
	cfg := &M.RawConfig{
		Rule: []string{
			"DOMAIN,subscription.example,Proxy",
			"MATCH,Proxy",
		},
	}

	persistRules := []string{
		"DOMAIN,persist.example,DIRECT",
		"DOMAIN-SUFFIX,example,DIRECT",
	}
	sessionRules := []string{"DOMAIN,session.example,REJECT"}

	prependOverrideRules(cfg, persistRules, sessionRules)

	want := []string{
		"DOMAIN,session.example,REJECT",
		"DOMAIN,persist.example,DIRECT",
		"DOMAIN-SUFFIX,example,DIRECT",
		"DOMAIN,subscription.example,Proxy",
		"MATCH,Proxy",
	}
	if !reflect.DeepEqual(cfg.Rule, want) {
		t.Fatalf("rules = %#v, want %#v", cfg.Rule, want)
	}
}

func TestPatchOverrideEndToEnd(t *testing.T) {
	useTemporaryOverrideHome(t)

	WriteOverride(OverrideSlotPersist, `{
		"mixed-port": 9090,
		"prepend-rules": ["DOMAIN,persist.example,DIRECT"]
	}`)
	WriteOverride(OverrideSlotSession, `{
		"ipv6": true,
		"prepend-rules": ["DOMAIN,session.example,REJECT"]
	}`)

	cfg := &M.RawConfig{
		MixedPort: 7890,
		Rule:      []string{"MATCH,Subscription"},
	}
	if err := patchOverride(cfg, ""); err != nil {
		t.Fatalf("patch override: %v", err)
	}

	if cfg.MixedPort != 9090 {
		t.Fatalf("mixed port = %d, want 9090", cfg.MixedPort)
	}
	if !cfg.IPv6 {
		t.Fatal("session override should still patch regular fields")
	}
	want := []string{
		"DOMAIN,session.example,REJECT",
		"DOMAIN,persist.example,DIRECT",
		"MATCH,Subscription",
	}
	if !reflect.DeepEqual(cfg.Rule, want) {
		t.Fatalf("rules = %#v, want %#v", cfg.Rule, want)
	}
}

func TestApplyOverrideContentPreservesLegacyRuleOverride(t *testing.T) {
	cfg := &M.RawConfig{Rule: []string{"MATCH,Subscription"}}

	persistRules, err := applyOverrideContent(cfg, `{
		"rule": ["MATCH,DIRECT"],
		"prepend-rules": ["DOMAIN,example.com,REJECT"]
	}`)
	if err != nil {
		t.Fatalf("apply override: %v", err)
	}

	prependOverrideRules(cfg, persistRules, nil)
	want := []string{
		"DOMAIN,example.com,REJECT",
		"MATCH,DIRECT",
	}
	if !reflect.DeepEqual(cfg.Rule, want) {
		t.Fatalf("rules = %#v, want %#v", cfg.Rule, want)
	}
}

func TestMergedRulesReachMihomoParserInPriorityOrder(t *testing.T) {
	raw := M.DefaultRawConfig()
	raw.Rule = []string{
		"DOMAIN,subscription.example,REJECT",
		"MATCH,DIRECT",
	}
	prependOverrideRules(
		raw,
		[]string{"DOMAIN,custom.example,DIRECT"},
		nil,
	)

	parsed, err := M.ParseRawConfig(raw)
	if err != nil {
		t.Fatalf("parse merged config: %v", err)
	}
	if len(parsed.Rules) != 3 {
		t.Fatalf("parsed rule count = %d, want 3", len(parsed.Rules))
	}

	wantPayloads := []string{"custom.example", "subscription.example", ""}
	wantAdapters := []string{"DIRECT", "REJECT", "DIRECT"}
	for i, rule := range parsed.Rules {
		if rule.Payload() != wantPayloads[i] || rule.Adapter() != wantAdapters[i] {
			t.Fatalf(
				"rule[%d] = payload %q adapter %q, want payload %q adapter %q",
				i,
				rule.Payload(),
				rule.Adapter(),
				wantPayloads[i],
				wantAdapters[i],
			)
		}
	}
}

func TestPrependOverrideRulesNoOp(t *testing.T) {
	cfg := &M.RawConfig{Rule: []string{"MATCH,DIRECT"}}
	original := &cfg.Rule[0]

	prependOverrideRules(cfg, nil, nil)

	if &cfg.Rule[0] != original {
		t.Fatal("empty overrides should not allocate a new rule slice")
	}
}

func BenchmarkPrependOverrideRulesDisabled(b *testing.B) {
	cfg := &M.RawConfig{Rule: []string{"MATCH,DIRECT"}}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		prependOverrideRules(cfg, nil, nil)
	}
}
