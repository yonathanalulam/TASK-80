package procurement

import (
	"testing"
)

func TestRFQStatusConstants(t *testing.T) {
	statuses := []RFQStatus{
		RFQStatusDraft,
		RFQStatusIssued,
		RFQStatusResponded,
		RFQStatusComparisonReady,
		RFQStatusSelected,
		RFQStatusClosedNoAward,
		RFQStatusConvertedPO,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("RFQ status constant should not be empty")
		}
	}
}

func TestPOStatusConstants(t *testing.T) {
	statuses := []POStatus{
		POStatusDraft,
		POStatusIssued,
		POStatusAccepted,
		POStatusPartiallyDelivered,
		POStatusDelivered,
		POStatusInspectionPending,
		POStatusExceptionOpen,
		POStatusClosed,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("PO status constant should not be empty")
		}
	}
}

func TestInspectionStatusConstants(t *testing.T) {
	statuses := []InspectionStatus{
		InspectionStatusPending,
		InspectionStatusPassed,
		InspectionStatusFailed,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("Inspection status constant should not be empty")
		}
	}
}

func TestExceptionStatusConstants(t *testing.T) {
	statuses := []ExceptionStatus{
		ExceptionStatusOpen,
		ExceptionStatusClosed,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("Exception status constant should not be empty")
		}
	}
}
