// billing/internal/dto/addon_dto.go
package dto

import (
    "github.com/google/uuid"
    "shared/models"
)

type ListAddonsResponse struct {
    Addons []*models.PublicAddonPlan `json:"addons"`
}

type ListOrgAddonsResponse struct {
    Addons []*models.PublicOrganizationAddon `json:"addons"`
}

type CreateAddonCheckoutRequest struct {
    AddonPlanID uuid.UUID `json:"addon_plan_id"`
}

// EffectiveLimitsResponse — base plan + addons ka total
type EffectiveLimitsResponse struct {
    OrganizationID uuid.UUID      `json:"organization_id"`
    BasePlanTier   models.PlanTier `json:"base_plan_tier"`

    // Base limits (from plan)
    BaseMonthlyEventsLimit int64 `json:"base_monthly_events_limit"`
    BaseRetentionDays      int   `json:"base_retention_days"`
    BaseTeamSeatsLimit     int   `json:"base_team_seats_limit"`
    BaseProjectsLimit      int   `json:"base_projects_limit"`

    // Addon bonuses
    AddonEventsBonus    int64 `json:"addon_events_bonus"`
    AddonRetentionBonus int   `json:"addon_retention_bonus"`
    AddonSeatsBonus     int   `json:"addon_seats_bonus"`
    AddonProjectsBonus  int   `json:"addon_projects_bonus"`

    // Effective totals (what actually enforces)
    EffectiveMonthlyEventsLimit int64 `json:"effective_monthly_events_limit"`
    EffectiveRetentionDays      int   `json:"effective_retention_days"`
    EffectiveTeamSeatsLimit     int   `json:"effective_team_seats_limit"`
    EffectiveProjectsLimit      int   `json:"effective_projects_limit"`

    ActiveAddons []*models.PublicOrganizationAddon `json:"active_addons"`
}