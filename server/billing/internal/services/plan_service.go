package services

import (
	"context"

	"billing/internal/dto"
	"billing/internal/repository"
	"shared/logger"
)

// ============================================================================
// PLAN SERVICE
// ============================================================================

type PlanService struct {
	planRepo *repository.PlanRepository
	log      *logger.Logger
}

func NewPlanService(planRepo *repository.PlanRepository, log *logger.Logger) *PlanService {
	return &PlanService{planRepo: planRepo, log: log}
}

// ListPlans returns all active plans for the pricing page
func (s *PlanService) ListPlans(ctx context.Context) (*dto.ListPlansResponse, error) {
	plans, err := s.planRepo.ListAll(ctx)
	if err != nil {
		s.log.Error("Failed to list plans", err)
		return nil, err
	}

	publicPlans := make([]*dto.PublicPlanWithSavings, 0, len(plans))
	for _, p := range plans {
		publicPlans = append(publicPlans, &dto.PublicPlanWithSavings{
			PublicSubscriptionPlan: p.ToPublic(),
			FormattedPrice:         p.GetFormattedPrice(),
			FormattedPricePerMonth: p.GetFormattedPricePerMonth(),
			SavingsText:            p.GetSavingsText(),
		})
	}

	return &dto.ListPlansResponse{Plans: publicPlans}, nil
}