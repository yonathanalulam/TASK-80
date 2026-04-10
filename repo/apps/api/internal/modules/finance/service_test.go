package finance

import (
	"testing"
)

func TestMinRefundAmount(t *testing.T) {
	if MinRefundAmount != 1.00 {
		t.Errorf("MinRefundAmount = %f, want 1.00", MinRefundAmount)
	}
}

func TestMaxDailyWithdrawal(t *testing.T) {
	if MaxDailyWithdrawal != 2500.00 {
		t.Errorf("MaxDailyWithdrawal = %f, want 2500.00", MaxDailyWithdrawal)
	}
}

func TestRefundStatusConstants(t *testing.T) {
	// Verify refund status values exist (prevents typo drift)
	if RefundStatusApproved == "" {
		t.Error("RefundStatusApproved should not be empty")
	}
	if RefundStatusPending == "" {
		t.Error("RefundStatusPending should not be empty")
	}
	if RefundStatusRejected == "" {
		t.Error("RefundStatusRejected should not be empty")
	}
}

func TestWithdrawalStatusConstants(t *testing.T) {
	if WithdrawalStatusRequested == "" {
		t.Error("WithdrawalStatusRequested should not be empty")
	}
	if WithdrawalStatusApproved == "" {
		t.Error("WithdrawalStatusApproved should not be empty")
	}
	if WithdrawalStatusRejected == "" {
		t.Error("WithdrawalStatusRejected should not be empty")
	}
	if WithdrawalStatusSettled == "" {
		t.Error("WithdrawalStatusSettled should not be empty")
	}
}

func TestWalletTypeConstants(t *testing.T) {
	types := []WalletType{
		WalletTypeCustomer,
		WalletTypeSupplier,
		WalletTypeCourier,
		WalletTypeSystem,
	}
	for _, wt := range types {
		if wt == "" {
			t.Error("wallet type constant should not be empty")
		}
	}
}

func TestEscrowStatusConstants(t *testing.T) {
	statuses := []EscrowStatus{
		EscrowStatusHeld,
		EscrowStatusPartial,
		EscrowStatusReleased,
		EscrowStatusRefunded,
	}
	for _, s := range statuses {
		if s == "" {
			t.Error("escrow status constant should not be empty")
		}
	}
}

func TestDirectionConstants(t *testing.T) {
	if Debit == "" {
		t.Error("Debit direction should not be empty")
	}
	if Credit == "" {
		t.Error("Credit direction should not be empty")
	}
}
