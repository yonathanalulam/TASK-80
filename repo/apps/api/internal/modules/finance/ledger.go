package finance

import (
	"context"
	"fmt"
	"math"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	CashOnHand                  = "1000"
	ManualTenderClearing        = "1100"
	EscrowLiability             = "2000"
	SupplierPayable             = "2100"
	CourierPayable              = "2200"
	CustomerWalletLiability     = "2300"
	RefundLiability             = "2400"
	SettlementAdjustmentReserve = "2500"
	Revenue                     = "4000"
	FeeRevenue                  = "4100"
	AdjustmentExpense           = "5000"
)

type Direction string

const (
	Debit  Direction = "debit"
	Credit Direction = "credit"
)

type JournalLine struct {
	AccountCode    string
	Direction      Direction
	Amount         float64
	CounterpartyID *string
}

func PostJournalEntry(
	ctx context.Context,
	tx pgx.Tx,
	entryType string,
	refType string,
	refID uuid.UUID,
	description string,
	createdBy string,
	lines []JournalLine,
) error {
	if len(lines) == 0 {
		return fmt.Errorf("journal entry must have at least one line")
	}

	var totalDebits, totalCredits float64
	for _, l := range lines {
		if l.Amount <= 0 {
			return fmt.Errorf("journal line amount must be positive, got %.2f", l.Amount)
		}
		switch l.Direction {
		case Debit:
			totalDebits += l.Amount
		case Credit:
			totalCredits += l.Amount
		default:
			return fmt.Errorf("invalid direction %q, must be debit or credit", l.Direction)
		}
	}

	totalDebits = math.Round(totalDebits*100) / 100
	totalCredits = math.Round(totalCredits*100) / 100
	if totalDebits != totalCredits {
		return fmt.Errorf("journal entry is unbalanced: debits=%.2f credits=%.2f", totalDebits, totalCredits)
	}

	entryID := uuid.New()
	_, err := tx.Exec(ctx,
		`INSERT INTO journal_entries (id, entry_type, reference_type, reference_id, description, effective_at, created_by, created_at)
		 VALUES ($1, $2, $3, $4, $5, NOW(), $6, NOW())`,
		entryID, entryType, refType, refID, description, createdBy,
	)
	if err != nil {
		return fmt.Errorf("insert journal entry: %w", err)
	}

	for _, l := range lines {
		lineID := uuid.New()
		_, err := tx.Exec(ctx,
			`INSERT INTO journal_lines (id, journal_entry_id, account_code, direction, amount, counterparty_id)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			lineID, entryID, l.AccountCode, string(l.Direction), l.Amount, l.CounterpartyID,
		)
		if err != nil {
			return fmt.Errorf("insert journal line: %w", err)
		}
	}

	return nil
}
