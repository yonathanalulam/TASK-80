package reviews

import (
	"context"
	"math"
	"time"

	"go.uber.org/zap"

	"travel-platform/apps/api/internal/common"
)

type ReviewService struct {
	repo   *Repository
	logger *zap.Logger
}

func NewReviewService(repo *Repository, logger *zap.Logger) *ReviewService {
	return &ReviewService{repo: repo, logger: logger}
}

func (s *ReviewService) SubmitReview(ctx context.Context, userID string, req CreateReviewRequest) error {
	if req.OverallRating < 1.0 || req.OverallRating > 5.0 {
		return common.NewBadRequestError("overall rating must be between 1.0 and 5.0")
	}

	if req.OverallRating < 3.0 && req.Comment == "" {
		return common.NewBadRequestError("a comment is required for ratings below 3.0")
	}

	if req.SubjectID == "" || req.OrderType == "" || req.OrderID == "" {
		return common.NewBadRequestError("subjectId, orderType, and orderId are required")
	}

	if userID == req.SubjectID {
		return common.NewBadRequestError("you cannot review yourself")
	}

	rev := &Review{
		ReviewerID:    userID,
		SubjectID:     req.SubjectID,
		OrderType:     req.OrderType,
		OrderID:       req.OrderID,
		OverallRating: req.OverallRating,
		Comment:       req.Comment,
		EditableUntil: time.Now().UTC().Add(48 * time.Hour),
	}

	reviewID, err := s.repo.CreateReview(ctx, rev)
	if err != nil {
		return err
	}

	for _, ds := range req.Scores {
		dim, err := s.repo.GetDimensionByName(ctx, ds.DimensionName)
		if err != nil {
			s.logger.Warn("skipping unknown review dimension",
				zap.String("dimension", ds.DimensionName),
				zap.Error(err),
			)
			continue
		}
		if !dim.Active {
			continue
		}
		if err := s.repo.CreateReviewScore(ctx, reviewID, dim.ID, ds.Score); err != nil {
			s.logger.Error("failed to insert review score",
				zap.String("reviewId", reviewID),
				zap.String("dimensionId", dim.ID),
				zap.Error(err),
			)
		}
	}

	go func() {
		bgCtx := context.Background()
		if _, err := s.ComputeCreditTier(bgCtx, req.SubjectID); err != nil {
			s.logger.Error("failed to recompute credit tier after review",
				zap.String("subjectId", req.SubjectID),
				zap.Error(err),
			)
		}
	}()

	s.logger.Info("review submitted",
		zap.String("reviewerId", userID),
		zap.String("subjectId", req.SubjectID),
		zap.String("reviewId", reviewID),
	)

	return nil
}

func (s *ReviewService) GetReviewsForSubject(ctx context.Context, subjectID string, page, pageSize int) ([]ReviewResponse, int, error) {
	revs, total, err := s.repo.GetReviewsBySubject(ctx, subjectID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var responses []ReviewResponse
	for _, rev := range revs {
		scores, err := s.repo.GetScoresForReview(ctx, rev.ID)
		if err != nil {
			s.logger.Warn("failed to load scores for review", zap.String("reviewId", rev.ID), zap.Error(err))
			scores = nil
		}
		responses = append(responses, ReviewResponse{
			ID:            rev.ID,
			ReviewerID:    rev.ReviewerID,
			SubjectID:     rev.SubjectID,
			OrderType:     rev.OrderType,
			OrderID:       rev.OrderID,
			OverallRating: rev.OverallRating,
			Comment:       rev.Comment,
			Scores:        scores,
			EditableUntil: rev.EditableUntil,
			CreatedAt:     rev.CreatedAt,
		})
	}

	return responses, total, nil
}

func (s *ReviewService) ComputeCreditTier(ctx context.Context, userID string) (*CreditSnapshotResponse, error) {
	avgRating, totalTx, err := s.repo.GetAvgRatingForSubject(ctx, userID)
	if err != nil {
		return nil, common.NewInternalError("compute avg rating", err)
	}

	violationCount, err := s.repo.GetViolationCount(ctx, userID)
	if err != nil {
		return nil, common.NewInternalError("count violations", err)
	}

	tiers, err := s.repo.GetAllCreditTiers(ctx)
	if err != nil {
		return nil, common.NewInternalError("load credit tiers", err)
	}

	matchedTier := TierBronze
	for _, tier := range tiers {
		if totalTx >= tier.MinTransactions &&
			avgRating >= tier.MinAvgRating &&
			violationCount <= tier.MaxViolations {
			matchedTier = tier.TierName
			break
		}
	}

	avgRating = math.Round(avgRating*10) / 10

	snap := &UserCreditSnapshot{
		UserID:            userID,
		Tier:              matchedTier,
		TotalTransactions: totalTx,
		AvgRating:         avgRating,
		ViolationCount:    violationCount,
	}

	if err := s.repo.CreateCreditSnapshot(ctx, snap); err != nil {
		return nil, common.NewInternalError("save credit snapshot", err)
	}

	return &CreditSnapshotResponse{
		UserID:            snap.UserID,
		Tier:              snap.Tier,
		TotalTransactions: snap.TotalTransactions,
		AvgRating:         snap.AvgRating,
		ViolationCount:    snap.ViolationCount,
		ComputedAt:        snap.ComputedAt,
	}, nil
}

func (s *ReviewService) RecordViolation(ctx context.Context, recordedBy string, req CreateViolationRequest) error {
	if req.UserID == "" || req.ViolationType == "" || req.Description == "" {
		return common.NewBadRequestError("userId, violationType, and description are required")
	}

	v := &ViolationRecord{
		UserID:        req.UserID,
		ViolationType: req.ViolationType,
		Description:   req.Description,
		Severity:      req.Severity,
		RecordedBy:    recordedBy,
	}

	if err := s.repo.CreateViolation(ctx, v); err != nil {
		return common.NewInternalError("record violation", err)
	}

	s.logger.Info("violation recorded",
		zap.String("userId", req.UserID),
		zap.String("type", req.ViolationType),
		zap.String("recordedBy", recordedBy),
	)

	return nil
}

func (s *ReviewService) RecordNoShow(ctx context.Context, recordedBy string, req CreateNoShowRequest) error {
	if req.UserID == "" || req.OrderType == "" || req.OrderID == "" {
		return common.NewBadRequestError("userId, orderType, and orderId are required")
	}

	ns := &NoShowRecord{
		UserID:     req.UserID,
		OrderType:  req.OrderType,
		OrderID:    req.OrderID,
		RecordedBy: recordedBy,
	}

	if err := s.repo.CreateNoShow(ctx, ns); err != nil {
		return common.NewInternalError("record no-show", err)
	}

	s.logger.Info("no-show recorded",
		zap.String("userId", req.UserID),
		zap.String("orderType", req.OrderType),
		zap.String("recordedBy", recordedBy),
	)

	return nil
}

func (s *ReviewService) FlagHarassment(ctx context.Context, reporterID string, req CreateHarassmentFlagRequest) error {
	if req.SubjectID == "" || req.Description == "" {
		return common.NewBadRequestError("subjectId and description are required")
	}

	hf := &HarassmentFlag{
		ReporterID:     reporterID,
		SubjectID:      req.SubjectID,
		Description:    req.Description,
		EvidenceFileID: req.EvidenceFileID,
		Status:         "pending",
	}

	if err := s.repo.CreateHarassmentFlag(ctx, hf); err != nil {
		return common.NewInternalError("flag harassment", err)
	}

	s.logger.Info("harassment flagged",
		zap.String("reporterId", reporterID),
		zap.String("subjectId", req.SubjectID),
	)

	return nil
}

func (s *ReviewService) GetCreditSnapshot(ctx context.Context, userID string) (*CreditSnapshotResponse, error) {
	snap, err := s.repo.GetLatestCreditSnapshot(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &CreditSnapshotResponse{
		UserID:            snap.UserID,
		Tier:              snap.Tier,
		TotalTransactions: snap.TotalTransactions,
		AvgRating:         snap.AvgRating,
		ViolationCount:    snap.ViolationCount,
		ComputedAt:        snap.ComputedAt,
	}, nil
}

func (s *ReviewService) GetRiskSummary(ctx context.Context, userID string) (*RiskSummaryResponse, error) {
	snap, err := s.repo.GetLatestCreditSnapshot(ctx, userID)
	var snapResp *CreditSnapshotResponse
	if err == nil {
		snapResp = &CreditSnapshotResponse{
			UserID:            snap.UserID,
			Tier:              snap.Tier,
			TotalTransactions: snap.TotalTransactions,
			AvgRating:         snap.AvgRating,
			ViolationCount:    snap.ViolationCount,
			ComputedAt:        snap.ComputedAt,
		}
	}

	violationCount, _ := s.repo.GetViolationCount(ctx, userID)
	noShowCount, _ := s.repo.GetNoShowCount(ctx, userID)
	harassmentCount, _ := s.repo.GetHarassmentFlagCount(ctx, userID)
	blacklisted, _ := s.repo.IsBlacklisted(ctx, userID)

	return &RiskSummaryResponse{
		UserID:          userID,
		CreditSnapshot:  snapResp,
		ViolationCount:  violationCount,
		NoShowCount:     noShowCount,
		HarassmentFlags: harassmentCount,
		IsBlacklisted:   blacklisted,
	}, nil
}
