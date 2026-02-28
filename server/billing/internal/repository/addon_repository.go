// billing/internal/repository/addon_repository.go
package repository

import (
    "context"
    "database/sql"
    "time"

    "github.com/google/uuid"

    "billing/internal/db"
    "shared/errors"
    "shared/logger"
    "shared/models"
)

type AddonRepository struct {
    db *db.DB
}

func NewAddonRepository(database *db.DB) *AddonRepository {
    return &AddonRepository{db: database}
}

func (r *AddonRepository) withWriteContext() (context.Context, context.CancelFunc) {
    return context.WithTimeout(context.Background(), 5*time.Second)
}

// ── AddonPlan (catalog) ──────────────────────────────────────────────────────

const addonPlanCols = `
    id, name, description, addon_type, quantity,
    price, currency, lemon_squeezy_variant_id,
    compatible_plans, display_order, is_active,
    created_at, updated_at`

func (r *AddonRepository) scanAddonPlan(s interface{ Scan(...any) error }) (*models.AddonPlan, error) {
    a := &models.AddonPlan{}
    err := s.Scan(
        &a.ID, &a.Name, &a.Description, &a.AddonType, &a.Quantity,
        &a.Price, &a.Currency, &a.LemonSqueezyVariantID,
        &a.CompatiblePlans, &a.DisplayOrder, &a.IsActive,
        &a.CreatedAt, &a.UpdatedAt,
    )
    return a, err
}

func (r *AddonRepository) ListAddonPlans(ctx context.Context) ([]*models.AddonPlan, error) {
    query := `SELECT ` + addonPlanCols + `
        FROM addon_plans WHERE is_active = TRUE ORDER BY display_order ASC`

    rows, err := r.db.Query(ctx, query)
    if err != nil {
        return nil, errors.Internal(err)
    }
    defer rows.Close()

    var addons []*models.AddonPlan
    for rows.Next() {
        a, err := r.scanAddonPlan(rows)
        if err != nil {
            return nil, errors.Internal(err)
        }
        addons = append(addons, a)
    }
    return addons, rows.Err()
}

func (r *AddonRepository) GetAddonPlanByVariantID(ctx context.Context, variantID string) (*models.AddonPlan, error) {
    query := `SELECT ` + addonPlanCols + `
        FROM addon_plans WHERE lemon_squeezy_variant_id = $1 AND is_active = TRUE LIMIT 1`

    a, err := r.scanAddonPlan(r.db.QueryRow(ctx, query, variantID))
    if err == sql.ErrNoRows {
        return nil, errors.NotFound("Addon plan not found")
    }
    if err != nil {
        return nil, errors.Internal(err)
    }
    return a, nil
}

func (r *AddonRepository) GetAddonPlanByID(ctx context.Context, id uuid.UUID) (*models.AddonPlan, error) {
    query := `SELECT ` + addonPlanCols + ` FROM addon_plans WHERE id = $1`

    a, err := r.scanAddonPlan(r.db.QueryRow(ctx, query, id))
    if err == sql.ErrNoRows {
        return nil, errors.NotFound("Addon plan not found")
    }
    if err != nil {
        return nil, errors.Internal(err)
    }
    return a, nil
}

// ── OrganizationAddon (active addons per org) ────────────────────────────────

const orgAddonCols = `
    id, organization_id, addon_plan_id,
    lemon_squeezy_subscription_id, lemon_squeezy_order_id,
    addon_type, quantity, price_amount, currency,
    status, current_period_start, current_period_end,
    cancel_at_period_end, canceled_at,
    created_at, updated_at`

func (r *AddonRepository) scanOrgAddon(s interface{ Scan(...any) error }) (*models.OrganizationAddon, error) {
    a := &models.OrganizationAddon{}
    err := s.Scan(
        &a.ID, &a.OrganizationID, &a.AddonPlanID,
        &a.LemonSqueezySubscriptionID, &a.LemonSqueezyOrderID,
        &a.AddonType, &a.Quantity, &a.PriceAmount, &a.Currency,
        &a.Status, &a.CurrentPeriodStart, &a.CurrentPeriodEnd,
        &a.CancelAtPeriodEnd, &a.CanceledAt,
        &a.CreatedAt, &a.UpdatedAt,
    )
    return a, err
}

func (r *AddonRepository) CreateOrgAddon(ctx context.Context, addon *models.OrganizationAddon) error {
    writeCtx, cancel := r.withWriteContext()
    defer cancel()

    query := `
        INSERT INTO organization_addons (
            id, organization_id, addon_plan_id,
            lemon_squeezy_subscription_id, lemon_squeezy_order_id,
            addon_type, quantity, price_amount, currency,
            status, current_period_start, current_period_end,
            cancel_at_period_end, canceled_at,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9,
            $10, $11, $12, $13, $14, $15, $16
        )`

    _, err := r.db.Exec(writeCtx, query,
        addon.ID, addon.OrganizationID, addon.AddonPlanID,
        addon.LemonSqueezySubscriptionID, addon.LemonSqueezyOrderID,
        addon.AddonType, addon.Quantity, addon.PriceAmount, addon.Currency,
        addon.Status, addon.CurrentPeriodStart, addon.CurrentPeriodEnd,
        addon.CancelAtPeriodEnd, addon.CanceledAt,
        addon.CreatedAt, addon.UpdatedAt,
    )
    if err != nil {
        logger.Error("Failed to create org addon", err, "org_id", addon.OrganizationID)
        return errors.Wrap(err, 500, "Failed to create addon")
    }
    return nil
}

func (r *AddonRepository) UpdateOrgAddon(ctx context.Context, addon *models.OrganizationAddon) error {
    addon.UpdatedAt = time.Now()
    writeCtx, cancel := r.withWriteContext()
    defer cancel()

    query := `
        UPDATE organization_addons SET
            status = $2,
            current_period_start = $3,
            current_period_end = $4,
            cancel_at_period_end = $5,
            canceled_at = $6,
            updated_at = $7
        WHERE id = $1`

    result, err := r.db.Exec(writeCtx, query,
        addon.ID, addon.Status,
        addon.CurrentPeriodStart, addon.CurrentPeriodEnd,
        addon.CancelAtPeriodEnd, addon.CanceledAt,
        addon.UpdatedAt,
    )
    if err != nil {
        logger.Error("Failed to update org addon", err, "addon_id", addon.ID)
        return errors.Wrap(err, 500, "Failed to update addon")
    }

    rows, _ := result.RowsAffected()
    if rows == 0 {
        return errors.NotFound("Addon not found")
    }
    return nil
}

// GetActiveAddonsByOrgID — limit calculation ke liye
func (r *AddonRepository) GetActiveAddonsByOrgID(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationAddon, error) {
    query := `
        SELECT ` + orgAddonCols + `
        FROM organization_addons
        WHERE organization_id = $1 AND status = 'active'
        ORDER BY created_at ASC`

    rows, err := r.db.Query(ctx, query, orgID)
    if err != nil {
        return nil, errors.Internal(err)
    }
    defer rows.Close()

    var addons []*models.OrganizationAddon
    for rows.Next() {
        a, err := r.scanOrgAddon(rows)
        if err != nil {
            return nil, errors.Internal(err)
        }
        addons = append(addons, a)
    }
    return addons, rows.Err()
}

// ListOrgAddons — user dashboard ke liye (all statuses)
func (r *AddonRepository) ListOrgAddons(ctx context.Context, orgID uuid.UUID) ([]*models.OrganizationAddon, error) {
    query := `
        SELECT ` + orgAddonCols + `
        FROM organization_addons
        WHERE organization_id = $1
        ORDER BY created_at DESC`

    rows, err := r.db.Query(ctx, query, orgID)
    if err != nil {
        return nil, errors.Internal(err)
    }
    defer rows.Close()

    var addons []*models.OrganizationAddon
    for rows.Next() {
        a, err := r.scanOrgAddon(rows)
        if err != nil {
            return nil, errors.Internal(err)
        }
        addons = append(addons, a)
    }
    return addons, rows.Err()
}

func (r *AddonRepository) GetOrgAddonByLemonSqueezyID(ctx context.Context, lsSubID string) (*models.OrganizationAddon, error) {
    query := `SELECT ` + orgAddonCols + `
        FROM organization_addons WHERE lemon_squeezy_subscription_id = $1`

    a, err := r.scanOrgAddon(r.db.QueryRow(ctx, query, lsSubID))
    if err == sql.ErrNoRows {
        return nil, errors.NotFound("Addon subscription not found")
    }
    if err != nil {
        return nil, errors.Internal(err)
    }
    return a, nil
}

func (r *AddonRepository) GetOrgAddonByID(ctx context.Context, id uuid.UUID) (*models.OrganizationAddon, error) {
    query := `SELECT ` + orgAddonCols + ` FROM organization_addons WHERE id = $1`

    a, err := r.scanOrgAddon(r.db.QueryRow(ctx, query, id))
    if err == sql.ErrNoRows {
        return nil, errors.NotFound("Addon not found")
    }
    if err != nil {
        return nil, errors.Internal(err)
    }
    return a, nil
}

// SumActiveAddonQuantity — ek specific type ke saare active addon quantities add karo
func (r *AddonRepository) SumActiveAddonQuantity(ctx context.Context, orgID uuid.UUID, addonType models.AddonType) (int64, error) {
    var total int64
    err := r.db.QueryRow(ctx, `
        SELECT COALESCE(SUM(quantity), 0)
        FROM organization_addons
        WHERE organization_id = $1 AND addon_type = $2 AND status = 'active'
    `, orgID, addonType).Scan(&total)
    if err != nil {
        return 0, errors.Internal(err)
    }
    return total, nil
}