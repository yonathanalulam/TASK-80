package procurement

import (
	"testing"
)

var canonicalRFQStatuses = []RFQStatus{
	RFQStatusDraft,
	RFQStatusIssued,
	RFQStatusResponded,
	RFQStatusComparisonReady,
	RFQStatusSelected,
	RFQStatusClosedNoAward,
	RFQStatusConvertedPO,
}

var canonicalPOStatuses = []POStatus{
	POStatusDraft,
	POStatusIssued,
	POStatusAccepted,
	POStatusPartiallyDelivered,
	POStatusDelivered,
	POStatusInspectionPending,
	POStatusExceptionOpen,
	POStatusClosed,
}

var dbRFQEnumValues = map[string]bool{
	"draft": true, "issued": true, "responded": true,
	"comparison_ready": true, "selected": true,
	"closed_no_award": true, "converted_to_po": true,
}

var dbPOEnumValues = map[string]bool{
	"draft": true, "issued": true, "accepted": true,
	"partially_delivered": true, "delivered": true,
	"inspection_pending": true, "exception_open": true, "closed": true,
}

func TestRFQStatusConstants_MatchDBEnum(t *testing.T) {
	for _, s := range canonicalRFQStatuses {
		if !dbRFQEnumValues[string(s)] {
			t.Errorf("RFQ status constant %q is not in DB rfq_status enum", s)
		}
	}
	if len(canonicalRFQStatuses) != len(dbRFQEnumValues) {
		t.Errorf("Go has %d RFQ statuses, DB enum has %d", len(canonicalRFQStatuses), len(dbRFQEnumValues))
	}
}

func TestPOStatusConstants_MatchDBEnum(t *testing.T) {
	for _, s := range canonicalPOStatuses {
		if !dbPOEnumValues[string(s)] {
			t.Errorf("PO status constant %q is not in DB po_status enum", s)
		}
	}
	if len(canonicalPOStatuses) != len(dbPOEnumValues) {
		t.Errorf("Go has %d PO statuses, DB enum has %d", len(canonicalPOStatuses), len(dbPOEnumValues))
	}
}

func TestRFQLifecycle_DraftToIssued(t *testing.T) {
	if RFQStatusDraft == RFQStatusIssued {
		t.Error("draft and issued should be different states")
	}
	if RFQStatusDraft != "draft" {
		t.Errorf("RFQStatusDraft = %q, want %q", RFQStatusDraft, "draft")
	}
	if RFQStatusIssued != "issued" {
		t.Errorf("RFQStatusIssued = %q, want %q", RFQStatusIssued, "issued")
	}
}

func TestPOLifecycle_DraftToAccepted(t *testing.T) {
	if POStatusDraft == POStatusAccepted {
		t.Error("draft and accepted should be different states")
	}
	if POStatusIssued != "issued" {
		t.Errorf("POStatusIssued = %q, want %q", POStatusIssued, "issued")
	}
}

var dbDiscrepancyEnumValues = map[string]bool{
	"shortage": true, "damage": true, "wrong_item": true,
	"late_delivery": true, "service_deviation": true, "other": true,
}

var canonicalDiscrepancyTypes = []DiscrepancyType{
	DiscrepancyTypeShortage,
	DiscrepancyTypeDamage,
	DiscrepancyTypeWrongItem,
	DiscrepancyTypeLateDelivery,
	DiscrepancyTypeServiceDeviation,
	DiscrepancyTypeOther,
}

func TestDiscrepancyTypeConstants_MatchDBEnum(t *testing.T) {
	for _, dt := range canonicalDiscrepancyTypes {
		if !dbDiscrepancyEnumValues[string(dt)] {
			t.Errorf("discrepancy type constant %q is not in DB discrepancy_type enum", dt)
		}
	}
	if len(canonicalDiscrepancyTypes) != len(dbDiscrepancyEnumValues) {
		t.Errorf("Go has %d discrepancy types, DB enum has %d", len(canonicalDiscrepancyTypes), len(dbDiscrepancyEnumValues))
	}
}

func TestNoStaleDiscrepancyTypes(t *testing.T) {
	staleValues := []string{"quantity", "quality"}
	for _, stale := range staleValues {
		for _, dt := range canonicalDiscrepancyTypes {
			if string(dt) == stale {
				t.Errorf("discrepancy type %q should not exist (stale)", stale)
			}
		}
	}
}

func TestNoStaleStatusConstants(t *testing.T) {
	staleValues := []string{"pending", "in_transit", "completed", "cancelled"}
	for _, stale := range staleValues {
		for _, s := range canonicalRFQStatuses {
			if string(s) == stale {
				t.Errorf("RFQ status %q should not exist (stale)", stale)
			}
		}
		for _, s := range canonicalPOStatuses {
			if string(s) == stale {
				t.Errorf("PO status %q should not exist (stale)", stale)
			}
		}
	}
}
