package risk

import (
	"testing"
)

func TestRiskDecision_AllowedByDefault(t *testing.T) {
	// A risk decision with Allowed=true should pass.
	d := &RiskDecision{Allowed: true}
	if !d.Allowed {
		t.Error("default decision should be allowed")
	}
}

func TestRiskDecision_Blocked(t *testing.T) {
	d := &RiskDecision{Allowed: false, Reason: "blacklisted"}
	if d.Allowed {
		t.Error("blocked decision should not be allowed")
	}
	if d.Reason == "" {
		t.Error("blocked decision should have a reason")
	}
}

func TestRiskDecision_RequireApproval(t *testing.T) {
	d := &RiskDecision{Allowed: false, RequireApproval: true, Reason: "too many cancellations"}
	if d.Allowed {
		t.Error("approval-required decision should not be allowed")
	}
	if !d.RequireApproval {
		t.Error("should require approval")
	}
}

func TestDetermineSeverity(t *testing.T) {
	tests := []struct {
		eventType string
		want      string
	}{
		{"harassment_flag", "high"},
		{"cancellation", "medium"},
		{"rfq_creation", "low"},
		{"unknown_event", "low"},
	}
	for _, tt := range tests {
		got := determineSeverity(tt.eventType)
		if got != tt.want {
			t.Errorf("determineSeverity(%q) = %q, want %q", tt.eventType, got, tt.want)
		}
	}
}
