// billing/internal/services/addon_service.go
package services

import (
    "context"
    "fmt"
    "strconv"
    "time"

    "github.com/google/uuid"

    "billing/internal/dto"
    "billing/internal/lemonsqueezy"
    "billing/internal/repository"
    "shared/errors"
    "shared/logger"
    "shared/models"
)

type AddonService struct {
    orgRepo     *repository.OrgRepository
    subRepo     *repository.SubscriptionRepository
    addonRepo   *repository.AddonRepository
    planRepo    *repository.PlanRepository
    lsClient    *lemonsqueezy.Client
    log         *logger.Logger
}

func NewAddonService(
    orgRepo *repository.OrgRepository,
    subRepo *repository.SubscriptionRepository,
    addonRepo *repository.AddonRepository,
    planRepo *repository.PlanRepository,
    lsClient *lemonsqueezy.Client,
    log *logger.Logger,
) *AddonService {
    return &AddonService{
        orgRepo:   orgRepo,
        subRepo:   subRepo,
        addonRepo: addonRepo,
        planRepo:  planRepo,
        lsClient:  lsClient,
        log:       log,
    }
}

// ListAvailableAddons — pricing page + dashboard ke liye
func (s *AddonService) ListAvailableAddons(ctx context.Context, orgID uuid.UUID) (*dto.ListAddonsResponse, error) {
    // Org ka current plan tier nikalo
    org, err := s.orgRepo.GetByID(ctx, orgID)
    if err != nil {
        return nil, err
    }

    allAddons, err := s.addonRepo.ListAddonPlans(ctx)
    if err != nil {
        return nil, err
    }

    // Sirf compatible addons dikhao
    publicAddons := make([]*models.PublicAddonPlan, 0)
    for _, a := range allAddons {
        if a.IsCompatibleWith(org.PlanTier) {
            publicAddons = append(publicAddons, a.ToPublic())
        }
    }

    return &dto.ListAddonsResponse{Addons: publicAddons}, nil
}

// CreateAddonCheckout — user addon purchase karna chahta hai
func (s *AddonService) CreateAddonCheckout(
    ctx context.Context,
    orgID, userID uuid.UUID,
    userEmail, userName string,
    req *dto.CreateAddonCheckoutRequest,
) (*dto.CreateCheckoutResponse, error) {
    // 1. Verify org has active paid subscription
    org, err := s.orgRepo.GetByID(ctx, orgID)
    if err != nil {
        return nil, err
    }

    if org.SubscriptionStatus != models.SubscriptionActive {
        return nil, errors.BadRequest("Active subscription required to purchase add-ons")
    }

    // 2. Get addon plan
    addonPlan, err := s.addonRepo.GetAddonPlanByID(ctx, req.AddonPlanID)
    if err != nil {
        return nil, err
    }

    // 3. Check compatibility with current plan
    if !addonPlan.IsCompatibleWith(org.PlanTier) {
        return nil, errors.BadRequest(
            fmt.Sprintf("This add-on is not available on your current plan (%s)", org.PlanTier),
        )
    }

    // 4. Parse variant ID
    variantID, err := strconv.Atoi(addonPlan.LemonSqueezyVariantID)
    if err != nil {
        return nil, errors.Internal(fmt.Errorf("invalid addon variant ID configuration"))
    }

    // 5. Build checkout — same flow as main subscription
    expiresAt := time.Now().Add(1 * time.Hour)

    checkoutReq := &lemonsqueezy.CreateCheckoutRequest{}
    checkoutReq.Data.Type = "checkouts"
    checkoutReq.Data.Attributes.StoreID = s.lsClient.GetStoreIDInt()
    checkoutReq.Data.Attributes.VariantID = variantID
    checkoutReq.Data.Attributes.ExpiresAt = &expiresAt
    checkoutReq.Data.Attributes.CheckoutData.Email = userEmail
    checkoutReq.Data.Attributes.CheckoutData.Name = userName
    checkoutReq.Data.Attributes.CheckoutData.Custom = map[string]string{
        "org_id":        orgID.String(),
        "user_id":       userID.String(),
        "purchase_type": "addon",                    // ← important: webhook mein distinguish karne ke liye
        "addon_plan_id": addonPlan.ID.String(),
    }
    checkoutReq.Data.Attributes.CheckoutOptions.SubscriptionPreview = true

    checkout, err := s.lsClient.CreateCheckout(ctx, checkoutReq)
    if err != nil {
        s.log.Error("Failed to create addon checkout", err, "org_id", orgID, "addon_plan_id", addonPlan.ID)
        return nil, errors.Internal(err)
    }

    s.log.Info("Addon checkout created", "org_id", orgID, "addon", addonPlan.Name)

    return &dto.CreateCheckoutResponse{
        CheckoutURL: checkout.Attributes.URL,
        ExpiresAt:   checkout.Attributes.ExpiresAt,
    }, nil
}

// CancelAddon — addon cancel karo (period end pe)
func (s *AddonService) CancelAddon(ctx context.Context, orgID, addonID uuid.UUID) (*models.PublicOrganizationAddon, error) {
    addon, err := s.addonRepo.GetOrgAddonByID(ctx, addonID)
    if err != nil {
        return nil, err
    }

    // Verify belongs to this org
    if addon.OrganizationID != orgID {
        return nil, errors.Forbidden("Addon does not belong to this organization")
    }

    if !addon.IsActive() {
        return nil, errors.BadRequest("Only active add-ons can be cancelled")
    }

    // Cancel via LemonSqueezy
    _, err = s.lsClient.CancelSubscription(ctx, addon.LemonSqueezySubscriptionID)
    if err != nil {
        s.log.Error("Failed to cancel addon on LS", err, "addon_id", addonID)
        return nil, errors.Internal(err)
    }

    // Mark locally
    addon.CancelAtPeriodEnd = true
    now := time.Now()
    addon.CanceledAt = &now

    if err := s.addonRepo.UpdateOrgAddon(ctx, addon); err != nil {
        return nil, err
    }

    s.log.Info("Addon cancelled", "org_id", orgID, "addon_id", addonID)
    return addon.ToPublic(), nil
}

// ListOrgAddons — org ke saare addons
func (s *AddonService) ListOrgAddons(ctx context.Context, orgID uuid.UUID) (*dto.ListOrgAddonsResponse, error) {
    addons, err := s.addonRepo.ListOrgAddons(ctx, orgID)
    if err != nil {
        return nil, err
    }

    public := make([]*models.PublicOrganizationAddon, 0, len(addons))
    for _, a := range addons {
        public = append(public, a.ToPublic())
    }

    return &dto.ListOrgAddonsResponse{Addons: public}, nil
}

// GetEffectiveLimits — base plan + all active addons = actual limits
func (s *AddonService) GetEffectiveLimits(ctx context.Context, orgID uuid.UUID) (*dto.EffectiveLimitsResponse, error) {
    org, err := s.orgRepo.GetByID(ctx, orgID)
    if err != nil {
        return nil, err
    }

    activeAddons, err := s.addonRepo.GetActiveAddonsByOrgID(ctx, orgID)
    if err != nil {
        return nil, err
    }

    // Start with base plan limits
    resp := &dto.EffectiveLimitsResponse{
        OrganizationID:          orgID,
        BasePlanTier:            org.PlanTier,
        BaseMonthlyEventsLimit:  org.MonthlyEventsLimit,
        BaseRetentionDays:       org.RetentionDays,
        BaseTeamSeatsLimit:      org.TeamSeatsLimit,
        BaseProjectsLimit:       org.ProjectsLimit,
        AddonEventsBonus:        0,
        AddonRetentionBonus:     0,
        AddonSeatsBonus:         0,
        AddonProjectsBonus:      0,
        ActiveAddons:            make([]*models.PublicOrganizationAddon, 0),
    }

    // Add up addon quantities
    for _, addon := range activeAddons {
        resp.ActiveAddons = append(resp.ActiveAddons, addon.ToPublic())
        switch addon.AddonType {
        case models.AddonTypeEvents:
            resp.AddonEventsBonus += addon.Quantity
        case models.AddonTypeRetention:
            resp.AddonRetentionBonus += int(addon.Quantity)
        case models.AddonTypeSeats:
            resp.AddonSeatsBonus += int(addon.Quantity)
        case models.AddonTypeProjects:
            resp.AddonProjectsBonus += int(addon.Quantity)
        }
    }

    // Effective totals
    resp.EffectiveMonthlyEventsLimit = org.MonthlyEventsLimit + resp.AddonEventsBonus
    resp.EffectiveRetentionDays      = org.RetentionDays + resp.AddonRetentionBonus
    resp.EffectiveTeamSeatsLimit     = org.TeamSeatsLimit + resp.AddonSeatsBonus
    resp.EffectiveProjectsLimit      = org.ProjectsLimit + resp.AddonProjectsBonus

    return resp, nil
}