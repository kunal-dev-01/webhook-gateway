package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"shared/models"
)

// CompanyRepository handles database operations for companies
type CompanyRepository struct {
	db *sqlx.DB
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(db *sqlx.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

// ============================================================================
// SQL Queries
// ============================================================================

const (
	companySelectColumns = `
		id, owner_id, company_name, legal_name, industry, description,
		email, phone, website,
		address_line1, address_line2, city, state, country, pincode,
		gstin, pan, tan, cin,
		pf_registration_number, pf_sub_code, 
		esi_registration_number, pt_registration_number,
		lwf_registration_number, msme_registration_number,
		logo_url, logo_thumbnail_url, primary_color, secondary_color,
		settings, status, financial_year_start_month,
		created_at, updated_at, deleted_at
	`

	baseCompanyQuery = `SELECT ` + companySelectColumns + ` FROM companies WHERE deleted_at IS NULL`

	createCompanyQuery = `
		INSERT INTO companies (
			id, owner_id, company_name, legal_name, industry, description,
			email, phone, website,
			address_line1, address_line2, city, state, country, pincode,
			gstin, pan, tan, cin,
			pf_registration_number, pf_sub_code,
			esi_registration_number, pt_registration_number,
			lwf_registration_number, msme_registration_number,
			logo_url, logo_thumbnail_url, primary_color, secondary_color,
			settings, status, financial_year_start_month,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9,
			$10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25,
			$26, $27, $28, $29,
			$30, $31, $32,
			$33, $34
		)
		RETURNING id, created_at, updated_at
	`

	updateCompanyQuery = `
		UPDATE companies SET
			owner_id = $2,
			company_name = $3,
			legal_name = $4,
			industry = $5,
			description = $6,
			email = $7,
			phone = $8,
			website = $9,
			address_line1 = $10,
			address_line2 = $11,
			city = $12,
			state = $13,
			country = $14,
			pincode = $15,
			gstin = $16,
			pan = $17,
			tan = $18,
			cin = $19,
			pf_registration_number = $20,
			pf_sub_code = $21,
			esi_registration_number = $22,
			pt_registration_number = $23,
			lwf_registration_number = $24,
			msme_registration_number = $25,
			logo_url = $26,
			logo_thumbnail_url = $27,
			primary_color = $28,
			secondary_color = $29,
			settings = $30,
			status = $31,
			financial_year_start_month = $32,
			updated_at = $33
		WHERE id = $1 AND deleted_at IS NULL
		RETURNING updated_at
	`
)

var (
	ErrCompanyNotFound       = fmt.Errorf("company not found")
	ErrCompanyNameExists     = fmt.Errorf("company name already exists")
	ErrCompanyAlreadyDeleted = fmt.Errorf("company not found or already deleted")
	ErrCompanyNotDeleted     = fmt.Errorf("company not found or not deleted")
	ErrInvalidOwner          = fmt.Errorf("invalid owner_id")
)

// ============================================================================
// Create Operations
// ============================================================================

// Create creates a new company
func (r *CompanyRepository) Create(ctx context.Context, company *models.Company) error {
	if company.ID == uuid.Nil {
		company.ID = uuid.New()
	}

	// Validate before creation
	if err := company.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	now := time.Now()
	company.CreatedAt = now
	company.UpdatedAt = now

	// Set defaults if not provided
	if company.Country == "" {
		company.Country = "India"
	}
	if company.PrimaryColor == "" {
		company.PrimaryColor = "#1976D2"
	}
	if company.Status == "" {
		company.Status = models.CompanyStatusActive
	}
	if company.FinancialYearStartMonth == 0 {
		company.FinancialYearStartMonth = 4 // April
	}

	err := r.db.QueryRowContext(ctx, createCompanyQuery,
		company.ID, company.OwnerID, company.CompanyName, company.LegalName, company.Industry, company.Description,
		company.Email, company.Phone, company.Website,
		company.AddressLine1, company.AddressLine2, company.City, company.State, company.Country, company.Pincode,
		company.GSTIN, company.PAN, company.TAN, company.CIN,
		company.PFRegistrationNumber, company.PFSubCode,
		company.ESIRegistrationNumber, company.PTRegistrationNumber,
		company.LWFRegistrationNumber, company.MSMERegistrationNumber,
		company.LogoURL, company.LogoThumbnailURL, company.PrimaryColor, company.SecondaryColor,
		company.Settings, company.Status, company.FinancialYearStartMonth,
		company.CreatedAt, company.UpdatedAt,
	).Scan(&company.ID, &company.CreatedAt, &company.UpdatedAt)

	if err != nil {
		return r.handleDBError(err, "create company")
	}
	return nil
}

// ============================================================================
// Read Operations
// ============================================================================

// GetByID retrieves a company by ID
func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	return r.getCompany(ctx, "id = $1", id)
}

// GetByOwnerID retrieves a company by owner ID
func (r *CompanyRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*models.Company, error) {
	return r.getCompany(ctx, "owner_id = $1", ownerID)
}

// GetByName retrieves a company by name (case-insensitive)
func (r *CompanyRepository) GetByName(ctx context.Context, name string) (*models.Company, error) {
	return r.getCompany(ctx, "LOWER(company_name) = LOWER($1)", name)
}

// GetByGSTIN retrieves a company by GSTIN
func (r *CompanyRepository) GetByGSTIN(ctx context.Context, gstin string) (*models.Company, error) {
	return r.getCompany(ctx, "gstin = $1", strings.ToUpper(gstin))
}

// GetByPAN retrieves a company by PAN
func (r *CompanyRepository) GetByPAN(ctx context.Context, pan string) (*models.Company, error) {
	return r.getCompany(ctx, "pan = $1", strings.ToUpper(pan))
}

// List retrieves companies with filters and pagination
func (r *CompanyRepository) List(ctx context.Context, filters ListCompaniesFilter) ([]*models.Company, int, error) {
	whereClauses, args := r.buildFilters(filters)
	whereClause := strings.Join(whereClauses, " AND ")

	total, err := r.countCompanies(ctx, whereClause, args)
	if err != nil {
		return nil, 0, err
	}

	companies, err := r.fetchCompanies(ctx, whereClause, args, filters)
	if err != nil {
		return nil, 0, err
	}

	return companies, total, nil
}

// GetActiveCompanies retrieves all active companies
func (r *CompanyRepository) GetActiveCompanies(ctx context.Context) ([]*models.Company, error) {
	query := baseCompanyQuery + " AND status = 'active' ORDER BY company_name ASC"
	var companies []*models.Company
	err := r.db.SelectContext(ctx, &companies, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get active companies: %w", err)
	}
	return companies, nil
}

// GetCompaniesByStatus retrieves companies by status
func (r *CompanyRepository) GetCompaniesByStatus(ctx context.Context, status models.CompanyStatus) ([]*models.Company, error) {
	query := baseCompanyQuery + " AND status = $1 ORDER BY company_name ASC"
	var companies []*models.Company
	err := r.db.SelectContext(ctx, &companies, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by status: %w", err)
	}
	return companies, nil
}

// GetCompaniesByIndustry retrieves companies by industry
func (r *CompanyRepository) GetCompaniesByIndustry(ctx context.Context, industry string) ([]*models.Company, error) {
	query := baseCompanyQuery + " AND industry = $1 ORDER BY company_name ASC"
	var companies []*models.Company
	err := r.db.SelectContext(ctx, &companies, query, industry)
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by industry: %w", err)
	}
	return companies, nil
}

// ============================================================================
// Update Operations
// ============================================================================

// Update updates a company
func (r *CompanyRepository) Update(ctx context.Context, company *models.Company) error {
	// Validate before update
	if err := company.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	company.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, updateCompanyQuery,
		company.ID, company.OwnerID, company.CompanyName, company.LegalName, company.Industry, company.Description,
		company.Email, company.Phone, company.Website,
		company.AddressLine1, company.AddressLine2, company.City, company.State, company.Country, company.Pincode,
		company.GSTIN, company.PAN, company.TAN, company.CIN,
		company.PFRegistrationNumber, company.PFSubCode,
		company.ESIRegistrationNumber, company.PTRegistrationNumber,
		company.LWFRegistrationNumber, company.MSMERegistrationNumber,
		company.LogoURL, company.LogoThumbnailURL, company.PrimaryColor, company.SecondaryColor,
		company.Settings, company.Status, company.FinancialYearStartMonth,
		company.UpdatedAt,
	).Scan(&company.UpdatedAt)

	if err != nil {
		return r.handleDBError(err, "update company")
	}
	return nil
}

// UpdateStatus updates company status
func (r *CompanyRepository) UpdateStatus(ctx context.Context, companyID uuid.UUID, status models.CompanyStatus) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			status = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID, status, now, "update company status")
}

// UpdateOwner updates company owner
func (r *CompanyRepository) UpdateOwner(ctx context.Context, companyID, newOwnerID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			owner_id = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID, newOwnerID, now, "update company owner")
}

// UpdateSettings updates company settings
func (r *CompanyRepository) UpdateSettings(ctx context.Context, companyID uuid.UUID, settings models.CompanySettings) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			settings = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`
	settingsJSON, err := settings.Value()
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}
	return r.execAndCheck(ctx, query, companyID, settingsJSON, now, "update company settings")
}

// UpdateLogo updates company logo URLs
func (r *CompanyRepository) UpdateLogo(ctx context.Context, companyID uuid.UUID, logoURL, thumbnailURL string) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			logo_url = $2,
			logo_thumbnail_url = $3,
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID, logoURL, thumbnailURL, now, "update company logo")
}

// UpdateBranding updates company branding (colors)
func (r *CompanyRepository) UpdateBranding(ctx context.Context, companyID uuid.UUID, primaryColor, secondaryColor string) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			primary_color = $2,
			secondary_color = $3,
			updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID, primaryColor, secondaryColor, now, "update company branding")
}

// UpdateContactInfo updates company contact information
func (r *CompanyRepository) UpdateContactInfo(ctx context.Context, companyID uuid.UUID, email, phone, website *string) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			email = $2,
			phone = $3,
			website = $4,
			updated_at = $5
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID, email, phone, website, now, "update company contact info")
}

// UpdateAddress updates company address
func (r *CompanyRepository) UpdateAddress(ctx context.Context, companyID uuid.UUID, address models.Company) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			address_line1 = $2,
			address_line2 = $3,
			city = $4,
			state = $5,
			country = $6,
			pincode = $7,
			updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID,
		address.AddressLine1, address.AddressLine2, address.City,
		address.State, address.Country, address.Pincode,
		now, "update company address")
}

// UpdateTaxInfo updates company tax information
func (r *CompanyRepository) UpdateTaxInfo(ctx context.Context, companyID uuid.UUID, gstin, pan, tan, cin *string) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			gstin = $2,
			pan = $3,
			tan = $4,
			cin = $5,
			updated_at = $6
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID, gstin, pan, tan, cin, now, "update company tax info")
}

// UpdateStatutoryInfo updates company statutory registration numbers
func (r *CompanyRepository) UpdateStatutoryInfo(ctx context.Context, companyID uuid.UUID, statutory models.Company) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			pf_registration_number = $2,
			pf_sub_code = $3,
			esi_registration_number = $4,
			pt_registration_number = $5,
			lwf_registration_number = $6,
			msme_registration_number = $7,
			updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`
	return r.execAndCheck(ctx, query, companyID,
		statutory.PFRegistrationNumber, statutory.PFSubCode,
		statutory.ESIRegistrationNumber, statutory.PTRegistrationNumber,
		statutory.LWFRegistrationNumber, statutory.MSMERegistrationNumber,
		now, "update company statutory info")
}

// ============================================================================
// Delete Operations
// ============================================================================

// SoftDelete soft deletes a company
func (r *CompanyRepository) SoftDelete(ctx context.Context, companyID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			deleted_at = $2,
			status = 'inactive',
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`
	if err := r.execAndCheck(ctx, query, companyID, now, "soft delete company"); err != nil {
		if err == ErrCompanyNotFound {
			return ErrCompanyAlreadyDeleted
		}
		return err
	}
	return nil
}

// Restore restores a soft-deleted company
func (r *CompanyRepository) Restore(ctx context.Context, companyID uuid.UUID) error {
	now := time.Now()
	query := `
		UPDATE companies SET
			deleted_at = NULL,
			status = 'active',
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NOT NULL
	`
	if err := r.execAndCheck(ctx, query, companyID, now, "restore company"); err != nil {
		if err == ErrCompanyNotFound {
			return ErrCompanyNotDeleted
		}
		return err
	}
	return nil
}

// HardDelete permanently deletes a company (use with extreme caution)
func (r *CompanyRepository) HardDelete(ctx context.Context, companyID uuid.UUID) error {
	query := `DELETE FROM companies WHERE id = $1`
	return r.execAndCheck(ctx, query, companyID, "hard delete company")
}

// ============================================================================
// Utility Methods
// ============================================================================

// Exists checks if a company exists by ID
func (r *CompanyRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE id = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, id)
	if err != nil {
		return false, fmt.Errorf("failed to check company existence: %w", err)
	}
	return exists, nil
}

// ExistsByName checks if company name exists
func (r *CompanyRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE LOWER(company_name) = LOWER($1) AND deleted_at IS NULL)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, name)
	if err != nil {
		return false, fmt.Errorf("failed to check company name existence: %w", err)
	}
	return exists, nil
}

// ExistsByGSTIN checks if GSTIN exists
func (r *CompanyRepository) ExistsByGSTIN(ctx context.Context, gstin string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE gstin = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, strings.ToUpper(gstin))
	if err != nil {
		return false, fmt.Errorf("failed to check GSTIN existence: %w", err)
	}
	return exists, nil
}

// ExistsByPAN checks if PAN exists
func (r *CompanyRepository) ExistsByPAN(ctx context.Context, pan string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE pan = $1 AND deleted_at IS NULL)`
	var exists bool
	err := r.db.GetContext(ctx, &exists, query, strings.ToUpper(pan))
	if err != nil {
		return false, fmt.Errorf("failed to check PAN existence: %w", err)
	}
	return exists, nil
}

// Count returns total number of companies
func (r *CompanyRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM companies WHERE deleted_at IS NULL`
	var count int
	err := r.db.GetContext(ctx, &count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count companies: %w", err)
	}
	return count, nil
}

// CountByStatus returns count of companies by status
func (r *CompanyRepository) CountByStatus(ctx context.Context, status models.CompanyStatus) (int, error) {
	query := `SELECT COUNT(*) FROM companies WHERE status = $1 AND deleted_at IS NULL`
	var count int
	err := r.db.GetContext(ctx, &count, query, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count companies by status: %w", err)
	}
	return count, nil
}

// CountByIndustry returns count of companies by industry
func (r *CompanyRepository) CountByIndustry(ctx context.Context, industry string) (int, error) {
	query := `SELECT COUNT(*) FROM companies WHERE industry = $1 AND deleted_at IS NULL`
	var count int
	err := r.db.GetContext(ctx, &count, query, industry)
	if err != nil {
		return 0, fmt.Errorf("failed to count companies by industry: %w", err)
	}
	return count, nil
}

// GetEmployeeCount returns employee count for a company
// NOTE: This requires joining with users table
func (r *CompanyRepository) GetEmployeeCount(ctx context.Context, companyID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND deleted_at IS NULL`
	var count int
	err := r.db.GetContext(ctx, &count, query, companyID)
	if err != nil {
		return 0, fmt.Errorf("failed to count company employees: %w", err)
	}
	return count, nil
}

// GetCompanyWithEmployeeCount returns company with employee count
func (r *CompanyRepository) GetCompanyWithEmployeeCount(ctx context.Context, companyID uuid.UUID) (*models.Company, int, error) {
	company, err := r.GetByID(ctx, companyID)
	if err != nil {
		return nil, 0, err
	}

	count, err := r.GetEmployeeCount(ctx, companyID)
	if err != nil {
		return nil, 0, err
	}

	return company, count, nil
}

// ============================================================================
// Batch Operations
// ============================================================================

// GetMultipleByIDs retrieves multiple companies by IDs
func (r *CompanyRepository) GetMultipleByIDs(ctx context.Context, ids []uuid.UUID) ([]*models.Company, error) {
	if len(ids) == 0 {
		return []*models.Company{}, nil
	}

	query := baseCompanyQuery + " AND id = ANY($1) ORDER BY company_name ASC"
	var companies []*models.Company
	err := r.db.SelectContext(ctx, &companies, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple companies: %w", err)
	}
	return companies, nil
}

// UpdateMultipleStatus updates status for multiple companies
func (r *CompanyRepository) UpdateMultipleStatus(ctx context.Context, ids []uuid.UUID, status models.CompanyStatus) error {
	if len(ids) == 0 {
		return nil
	}

	now := time.Now()
	query := `
		UPDATE companies SET
			status = $2,
			updated_at = $3
		WHERE id = ANY($1) AND deleted_at IS NULL
	`
	result, err := r.db.ExecContext(ctx, query, pq.Array(ids), status, now)
	if err != nil {
		return fmt.Errorf("failed to update multiple company statuses: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rows == 0 {
		return ErrCompanyNotFound
	}

	return nil
}

// ============================================================================
// Statistics & Analytics
// ============================================================================

// GetCompanyStats returns statistics about companies
type CompanyStats struct {
	TotalCompanies    int                        `json:"total_companies"`
	ActiveCompanies   int                        `json:"active_companies"`
	InactiveCompanies int                        `json:"inactive_companies"`
	SuspendedCompanies int                       `json:"suspended_companies"`
	ByIndustry        map[string]int             `json:"by_industry"`
	RecentCompanies   int                        `json:"recent_companies"` // Created in last 30 days
}

func (r *CompanyRepository) GetStats(ctx context.Context) (*CompanyStats, error) {
	stats := &CompanyStats{
		ByIndustry: make(map[string]int),
	}

	// Total count
	total, err := r.Count(ctx)
	if err != nil {
		return nil, err
	}
	stats.TotalCompanies = total

	// Count by status
	active, _ := r.CountByStatus(ctx, models.CompanyStatusActive)
	inactive, _ := r.CountByStatus(ctx, models.CompanyStatusInactive)
	suspended, _ := r.CountByStatus(ctx, models.CompanyStatusSuspended)

	stats.ActiveCompanies = active
	stats.InactiveCompanies = inactive
	stats.SuspendedCompanies = suspended

	// Count by industry
	query := `
		SELECT industry, COUNT(*) as count
		FROM companies
		WHERE deleted_at IS NULL AND industry IS NOT NULL
		GROUP BY industry
		ORDER BY count DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return stats, nil // Return partial stats
	}
	defer rows.Close()

	for rows.Next() {
		var industry string
		var count int
		if err := rows.Scan(&industry, &count); err == nil {
			stats.ByIndustry[industry] = count
		}
	}

	// Recent companies (last 30 days)
	recentQuery := `
		SELECT COUNT(*) FROM companies
		WHERE deleted_at IS NULL AND created_at >= NOW() - INTERVAL '30 days'
	`
	_ = r.db.GetContext(ctx, &stats.RecentCompanies, recentQuery)

	return stats, nil
}

// ============================================================================
// Helper Methods
// ============================================================================

func (r *CompanyRepository) getCompany(ctx context.Context, condition string, args ...any) (*models.Company, error) {
	query := baseCompanyQuery + " AND " + condition
	var company models.Company
	err := r.db.GetContext(ctx, &company, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrCompanyNotFound
		}
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	return &company, nil
}

func (r *CompanyRepository) buildFilters(filters ListCompaniesFilter) ([]string, []any) {
	whereClauses := []string{"deleted_at IS NULL"}
	args := []any{}
	argCounter := 1

	addFilter := func(column string, value any) {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = $%d", column, argCounter))
		args = append(args, value)
		argCounter++
	}

	if filters.Status != nil {
		addFilter("status", *filters.Status)
	}
	if filters.Industry != nil {
		addFilter("industry", *filters.Industry)
	}
	if filters.Country != nil {
		addFilter("country", *filters.Country)
	}
	if filters.State != nil {
		addFilter("state", *filters.State)
	}
	if filters.OwnerID != nil {
		addFilter("owner_id", *filters.OwnerID)
	}
	if filters.Search != "" {
		whereClauses = append(whereClauses, fmt.Sprintf(
			"(company_name ILIKE $%d OR legal_name ILIKE $%d OR email ILIKE $%d)",
			argCounter, argCounter, argCounter,
		))
		args = append(args, "%"+filters.Search+"%")
		argCounter++
	}

	return whereClauses, args
}

func (r *CompanyRepository) countCompanies(ctx context.Context, whereClause string, args []any) (int, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM companies WHERE %s", whereClause)
	var total int
	err := r.db.GetContext(ctx, &total, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count companies: %w", err)
	}
	return total, nil
}

func (r *CompanyRepository) fetchCompanies(ctx context.Context, whereClause string, args []any, filters ListCompaniesFilter) ([]*models.Company, error) {
	orderBy := "created_at DESC"
	if filters.SortBy != "" {
		direction := "ASC"
		if filters.SortDesc {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf("%s %s", filters.SortBy, direction)
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 10
	}
	offset := max(filters.Offset, 0)

	argCounter := len(args) + 1
	query := fmt.Sprintf(`
		SELECT %s FROM companies 
		WHERE %s 
		ORDER BY %s 
		LIMIT $%d OFFSET $%d
	`, companySelectColumns, whereClause, orderBy, argCounter, argCounter+1)

	args = append(args, limit, offset)

	var companies []*models.Company
	err := r.db.SelectContext(ctx, &companies, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}
	return companies, nil
}

func (r *CompanyRepository) execAndCheck(ctx context.Context, query string, args ...any) error {
	operation := args[len(args)-1].(string)
	execArgs := args[:len(args)-1]

	result, err := r.db.ExecContext(ctx, query, execArgs...)
	if err != nil {
		return fmt.Errorf("failed to %s: %w", operation, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected for %s: %w", operation, err)
	}

	if rows == 0 {
		return ErrCompanyNotFound
	}
	return nil
}

func (r *CompanyRepository) handleDBError(err error, operation string) error {
	if err == sql.ErrNoRows {
		return ErrCompanyNotFound
	}
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
		// Unique constraint violation
		if strings.Contains(pqErr.Constraint, "company_name") {
			return ErrCompanyNameExists
		}
		if strings.Contains(pqErr.Constraint, "gstin") {
			return fmt.Errorf("GSTIN already exists")
		}
		if strings.Contains(pqErr.Constraint, "pan") {
			return fmt.Errorf("PAN already exists")
		}
		return fmt.Errorf("duplicate entry: %s", pqErr.Constraint)
	}
	if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
		// Foreign key violation
		if strings.Contains(pqErr.Constraint, "owner_id") {
			return ErrInvalidOwner
		}
		return fmt.Errorf("foreign key violation: %s", pqErr.Constraint)
	}
	return fmt.Errorf("failed to %s: %w", operation, err)
}

// ============================================================================
// Filter Struct
// ============================================================================

// ListCompaniesFilter represents filters for listing companies
type ListCompaniesFilter struct {
	Status   *models.CompanyStatus
	Industry *string
	Country  *string
	State    *string
	OwnerID  *uuid.UUID
	Search   string // Searches: company_name, legal_name, email
	SortBy   string
	SortDesc bool
	Limit    int
	Offset   int
}