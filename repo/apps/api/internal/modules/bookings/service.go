package bookings

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"travel-platform/apps/api/internal/common"
	"travel-platform/apps/api/internal/modules/finance"
	"travel-platform/apps/api/internal/modules/pricing"
	"travel-platform/apps/api/internal/modules/risk"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Service struct {
	pool        *pgxpool.Pool
	repo        *Repository
	pricingRepo *pricing.Repository
	riskSvc     *risk.Service
	logger      *zap.Logger
}

func NewService(pool *pgxpool.Pool, repo *Repository, pricingRepo *pricing.Repository, riskSvc *risk.Service, logger *zap.Logger) *Service {
	return &Service{
		pool:        pool,
		repo:        repo,
		pricingRepo: pricingRepo,
		riskSvc:     riskSvc,
		logger:      logger,
	}
}

func (s *Service) CreateBooking(ctx context.Context, userID string, req CreateBookingRequest) (*BookingResponse, error) {
	if req.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one item is required")
	}

	var totalAmount float64
	items := make([]BookingItem, len(req.Items))
	for i, ri := range req.Items {
		subtotal := roundMoney(ri.UnitPrice * float64(ri.Quantity))
		totalAmount += subtotal
		items[i] = BookingItem{
			ItemType:    ri.ItemType,
			ItemName:    ri.ItemName,
			Description: ri.Description,
			UnitPrice:   ri.UnitPrice,
			Quantity:    ri.Quantity,
			Subtotal:    subtotal,
			Category:    ri.Category,
		}
	}
	totalAmount = roundMoney(totalAmount)

	booking := &Booking{
		OrganizerID: userID,
		ItineraryID: req.ItineraryID,
		Title:       req.Title,
		Description: req.Description,
		Status:      StatusDraft,
		TotalAmount: totalAmount,
	}

	bookingID, err := s.repo.Create(ctx, booking, items)
	if err != nil {
		return nil, fmt.Errorf("create booking: %w", err)
	}

	created, err := s.repo.GetByID(ctx, bookingID)
	if err != nil || created == nil {
		return nil, fmt.Errorf("fetch created booking: %w", err)
	}

	dbItems, err := s.repo.GetItems(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("fetch booking items: %w", err)
	}

	return toBookingResponse(created, dbItems), nil
}

func (s *Service) GetBooking(ctx context.Context, id, userID string) (*BookingResponse, error) {
	booking, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.OrganizerID != userID {
		return nil, fmt.Errorf("forbidden: not the organizer")
	}

	items, err := s.repo.GetItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get booking items: %w", err)
	}

	return toBookingResponse(booking, items), nil
}

func (s *Service) PricePreview(ctx context.Context, bookingID, userID string, couponCodes []string, isNewUser bool, membershipTier string) (*pricing.PricingResult, error) {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.OrganizerID != userID {
		return nil, fmt.Errorf("forbidden: not the organizer")
	}

	dbItems, err := s.repo.GetItems(ctx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get items: %w", err)
	}

	pricingItems := toPricingItems(dbItems)
	result, err := pricing.EvaluateCheckout(ctx, s.pricingRepo, pricingItems, couponCodes, userID, isNewUser, membershipTier, bookingID)
	if err != nil {
		return nil, fmt.Errorf("evaluate checkout: %w", err)
	}

	return result, nil
}

func (s *Service) Checkout(ctx context.Context, bookingID, userID string, req CheckoutRequest) (*CheckoutResponse, error) {
	if req.IdempotencyKey == "" {
		return nil, fmt.Errorf("idempotency key is required")
	}

	route := fmt.Sprintf("POST /api/v1/bookings/%s/checkout", bookingID)
	requestHash := hashRequest(req)

	found, cached, err := pricing.CheckIdempotencyKey(ctx, s.pool, userID, route, req.IdempotencyKey, requestHash)
	if err != nil {
		if found {
			return nil, fmt.Errorf("idempotency key reused with different request body")
		}
		return nil, fmt.Errorf("check idempotency: %w", err)
	}
	if found && cached != nil {
		var resp CheckoutResponse
		if err := json.Unmarshal(cached.ResponseBody, &resp); err != nil {
			return nil, fmt.Errorf("unmarshal cached response: %w", err)
		}
		return &resp, nil
	}

	if err := pricing.LockIdempotencyKey(ctx, s.pool, userID, route, req.IdempotencyKey, requestHash); err != nil {
		return nil, fmt.Errorf("lock idempotency: %w", err)
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	booking, err := s.repo.GetByIDTx(ctx, tx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get booking in tx: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.OrganizerID != userID {
		return nil, fmt.Errorf("forbidden: not the organizer")
	}
	if booking.Status != StatusDraft && booking.Status != StatusPendingPricing {
		return nil, fmt.Errorf("booking status must be draft or pending_pricing, got %s", booking.Status)
	}

	dbItems, err := s.repo.GetItemsTx(ctx, tx, bookingID)
	if err != nil {
		return nil, fmt.Errorf("get items in tx: %w", err)
	}
	if len(dbItems) == 0 {
		return nil, fmt.Errorf("booking has no items")
	}

	pricingItems := toPricingItems(dbItems)
	pricingResult, err := pricing.EvaluateCheckout(ctx, s.pricingRepo, pricingItems, req.CouponCodes, userID, req.IsNewUser, req.MembershipTier, bookingID)
	if err != nil {
		return nil, fmt.Errorf("evaluate checkout: %w", err)
	}
	for _, applied := range pricingResult.EligibleCoupons {
		redemption := &CouponRedemption{
			CouponID:           applied.CouponID,
			UserID:             userID,
			BookingID:          &bookingID,
			RedemptionScopeKey: bookingID,
			DiscountAmount:     applied.DiscountAmount,
		}
		if err := s.repo.CreateCouponRedemptionTx(ctx, tx, redemption); err != nil {
			return nil, fmt.Errorf("create coupon redemption: %w", err)
		}
	}

	escrow := &Escrow{
		OrderType:  "booking",
		OrderID:    bookingID,
		AmountHeld: pricingResult.EscrowHoldAmount,
		Status:     EscrowHeld,
	}
	if err := s.repo.CreateEscrowTx(ctx, tx, escrow); err != nil {
		return nil, fmt.Errorf("create escrow: %w", err)
	}

	bookingUUID, err := uuid.Parse(bookingID)
	if err != nil {
		return nil, fmt.Errorf("parse booking UUID: %w", err)
	}

	err = finance.PostJournalEntry(ctx, tx,
		"escrow_hold", "booking", bookingUUID,
		fmt.Sprintf("Escrow hold for booking %s", bookingID),
		userID,
		[]finance.JournalLine{
			{
				AccountCode: finance.CashOnHand,
				Direction:   finance.Debit,
				Amount:      pricingResult.EscrowHoldAmount,
			},
			{
				AccountCode: finance.EscrowLiability,
				Direction:   finance.Credit,
				Amount:      pricingResult.EscrowHoldAmount,
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("post journal entry: %w", err)
	}

	if err := s.repo.UpdateStatusTx(ctx, tx, bookingID, StatusPaidHeldInEscrow); err != nil {
		return nil, fmt.Errorf("update booking status: %w", err)
	}

	snapshotJSON, err := pricing.CreateSnapshot(pricingResult, pricingItems)
	if err != nil {
		return nil, fmt.Errorf("create snapshot: %w", err)
	}

	snapshotID, err := s.pricingRepo.SavePricingSnapshotTx(ctx, tx, &bookingID, snapshotJSON)
	if err != nil {
		return nil, fmt.Errorf("save snapshot: %w", err)
	}

	if err := s.repo.UpdateBookingAmountsTx(ctx, tx, bookingID, pricingResult.Subtotal, pricingResult.TotalDiscount, pricingResult.EscrowHoldAmount, snapshotID); err != nil {
		return nil, fmt.Errorf("update booking amounts: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	resp := &CheckoutResponse{
		BookingID:      bookingID,
		Status:         string(StatusPaidHeldInEscrow),
		TotalAmount:    pricingResult.Subtotal,
		DiscountAmount: pricingResult.TotalDiscount,
		EscrowAmount:   pricingResult.EscrowHoldAmount,
		SnapshotID:     snapshotID,
	}

	respBody, _ := json.Marshal(resp)
	if err := pricing.CompleteIdempotencyKey(ctx, s.pool, userID, route, req.IdempotencyKey, 200, respBody); err != nil {
		s.logger.Error("failed to complete idempotency key", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) RecordTender(ctx context.Context, bookingID, userID string, req TenderRecordRequest) error {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}
	if booking.OrganizerID != userID {
		return fmt.Errorf("forbidden: not the organizer")
	}

	record := &PaymentRecord{
		OrderType:     "booking",
		OrderID:       bookingID,
		TenderType:    req.TenderType,
		Amount:        req.Amount,
		Currency:      "USD",
		ReferenceText: req.ReferenceText,
		RecordedBy:    userID,
	}
	return s.repo.RecordTender(ctx, record)
}

func (s *Service) CancelBooking(ctx context.Context, bookingID, userID string) error {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}
	if booking.OrganizerID != userID {
		return fmt.Errorf("forbidden: not the organizer")
	}
	if booking.Status == StatusCancelled || booking.Status == StatusCompleted {
		return fmt.Errorf("booking cannot be cancelled in status %s", booking.Status)
	}

	_ = s.riskSvc.RecordEvent(ctx, userID, "cancellation", "booking cancelled: "+bookingID, "medium")
	decision, err := s.riskSvc.EvaluateAction(ctx, userID, "cancel_booking")
	if err == nil && !decision.Allowed {
		return common.NewForbiddenError("action blocked by risk engine: " + decision.Reason)
	}

	if booking.Status == StatusPaidHeldInEscrow {
		escrow, err := s.repo.GetEscrow(ctx, "booking", bookingID)
		if err != nil {
			return fmt.Errorf("get escrow: %w", err)
		}
		if escrow != nil && escrow.Status == EscrowHeld {
			if err := s.repo.UpdateEscrowStatus(ctx, escrow.ID, EscrowRefunded, 0, escrow.AmountHeld); err != nil {
				return fmt.Errorf("update escrow: %w", err)
			}
		}
	}

	return s.repo.UpdateStatus(ctx, bookingID, StatusCancelled)
}

func (s *Service) CompleteBooking(ctx context.Context, bookingID, userID string) error {
	booking, err := s.repo.GetByID(ctx, bookingID)
	if err != nil {
		return fmt.Errorf("get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}
	if booking.OrganizerID != userID {
		return fmt.Errorf("forbidden: not the organizer")
	}
	if booking.Status != StatusPaidHeldInEscrow {
		return fmt.Errorf("booking must be in paid_held_in_escrow status to complete, got %s", booking.Status)
	}

	escrow, err := s.repo.GetEscrow(ctx, "booking", bookingID)
	if err != nil {
		return fmt.Errorf("get escrow: %w", err)
	}
	if escrow != nil && escrow.Status == EscrowHeld {
		if err := s.repo.UpdateEscrowStatus(ctx, escrow.ID, EscrowReleased, escrow.AmountHeld, 0); err != nil {
			return fmt.Errorf("update escrow: %w", err)
		}
	}

	return s.repo.UpdateStatus(ctx, bookingID, StatusCompleted)
}

func (s *Service) ListBookings(ctx context.Context, userID string, page, pageSize int) (map[string]interface{}, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var total int
	err := s.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM bookings WHERE organizer_id = $1`, userID).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, organizer_id, itinerary_id, title, description, status, total_amount, discount_amount, escrow_amount, created_at, updated_at
		 FROM bookings WHERE organizer_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []map[string]interface{}
	for rows.Next() {
		var id, organizerID, title, status string
		var itineraryID, description *string
		var totalAmount, discountAmount, escrowAmount float64
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&id, &organizerID, &itineraryID, &title, &description, &status, &totalAmount, &discountAmount, &escrowAmount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, map[string]interface{}{
			"id": id, "organizerId": organizerID, "itineraryId": itineraryID,
			"title": title, "description": description, "status": status,
			"totalAmount": totalAmount, "discountAmount": discountAmount, "escrowAmount": escrowAmount,
			"createdAt": createdAt, "updatedAt": updatedAt,
		})
	}
	if bookings == nil {
		bookings = []map[string]interface{}{}
	}

	return map[string]interface{}{
		"items": bookings, "total": total, "page": page,
		"pageSize": pageSize, "totalPages": (total + pageSize - 1) / pageSize,
	}, rows.Err()
}
func toBookingResponse(b *Booking, items []BookingItem) *BookingResponse {
	resp := &BookingResponse{
		ID:                b.ID,
		OrganizerID:       b.OrganizerID,
		ItineraryID:       b.ItineraryID,
		Title:             b.Title,
		Description:       b.Description,
		Status:            b.Status,
		TotalAmount:       b.TotalAmount,
		DiscountAmount:    b.DiscountAmount,
		EscrowAmount:      b.EscrowAmount,
		PricingSnapshotID: b.PricingSnapshotID,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
	resp.Items = make([]BookingItemResponse, len(items))
	for i, item := range items {
		resp.Items[i] = BookingItemResponse{
			ID:          item.ID,
			ItemType:    item.ItemType,
			ItemName:    item.ItemName,
			Description: item.Description,
			UnitPrice:   item.UnitPrice,
			Quantity:    item.Quantity,
			Subtotal:    item.Subtotal,
			Category:    item.Category,
		}
	}
	return resp
}

func toPricingItems(items []BookingItem) []pricing.BookingItem {
	result := make([]pricing.BookingItem, len(items))
	for i, item := range items {
		result[i] = pricing.BookingItem{
			ID:        item.ID,
			BookingID: item.BookingID,
			ItemType:  item.ItemType,
			ItemName:  item.ItemName,
			UnitPrice: item.UnitPrice,
			Quantity:  item.Quantity,
			Subtotal:  item.Subtotal,
			Category:  item.Category,
		}
	}
	return result
}

func hashRequest(req CheckoutRequest) string {
	data, _ := json.Marshal(req)
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}
