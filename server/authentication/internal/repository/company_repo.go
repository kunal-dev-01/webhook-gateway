package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"authentication/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

type CompanyRepository struct {
	db *db.DB
}

func NewCompanyRepository(database *db.DB) *CompanyRepository {
	return &CompanyRepository{db: database}
}

// ============================================================================
// CREATE
// ============================================================================

func (r *CompanyRepository) Create(ctx context.Context, company *models.Company) error {
	query := `
		INSERT INTO companies (
			id, owner_id, company_name, legal_name, industry, description,
			email, phone, website,
			address_line1, address_line2, city, state, country, pincode,
			gstin, pan, tan, cin,
			pf_registration_number, pf_sub_code, esi_registration_number,
			pt_registration_number, lwf_registration_number, msme_registration_number,
			logo_url, logo_thumbnail_url, primary_color, secondary_color,
			settings, status, financial_year_start_month,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19,
			$20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34
		)`

	_, err := r.db.Exec(ctx, query,
		company.ID, company.OwnerID, company.CompanyName, company.LegalName, company.Industry, company.Description,
		company.Email, company.Phone, company.Website,
		company.AddressLine1, company.AddressLine2, company.City, company.State, company.Country, company.Pincode,
		company.GSTIN, company.PAN, company.TAN, company.CIN,
		company.PFRegistrationNumber, company.PFSubCode, company.ESIRegistrationNumber,
		company.PTRegistrationNumber, company.LWFRegistrationNumber, company.MSMERegistrationNumber,
		company.LogoURL, company.LogoThumbnailURL, company.PrimaryColor, company.SecondaryColor,
		company.Settings, company.Status, company.FinancialYearStartMonth,
		company.CreatedAt, company.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to create company", err, "company_name", company.CompanyName)
		return errors.Wrap(err, 500, "Failed to create company")
	}

	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	company := &models.Company{}
	query := `
		SELECT id, owner_id, company_name, legal_name, industry, description,
			email, phone, website,
			address_line1, address_line2, city, state, country, pincode,
			gstin, pan, tan, cin,
			pf_registration_number, pf_sub_code, esi_registration_number,
			pt_registration_number, lwf_registration_number, msme_registration_number,
			logo_url, logo_thumbnail_url, primary_color, secondary_color,
			settings, status, financial_year_start_month,
			created_at, updated_at, deleted_at
		FROM companies 
		WHERE id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&company.ID, &company.OwnerID, &company.CompanyName, &company.LegalName, &company.Industry, &company.Description,
		&company.Email, &company.Phone, &company.Website,
		&company.AddressLine1, &company.AddressLine2, &company.City, &company.State, &company.Country, &company.Pincode,
		&company.GSTIN, &company.PAN, &company.TAN, &company.CIN,
		&company.PFRegistrationNumber, &company.PFSubCode, &company.ESIRegistrationNumber,
		&company.PTRegistrationNumber, &company.LWFRegistrationNumber, &company.MSMERegistrationNumber,
		&company.LogoURL, &company.LogoThumbnailURL, &company.PrimaryColor, &company.SecondaryColor,
		&company.Settings, &company.Status, &company.FinancialYearStartMonth,
		&company.CreatedAt, &company.UpdatedAt, &company.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Company not found")
	}
	if err != nil {
		logger.Error("Failed to get company by ID", err, "company_id", id)
		return nil, errors.Internal(err)
	}

	return company, nil
}

func (r *CompanyRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*models.Company, error) {
	company := &models.Company{}
	query := `
		SELECT id, owner_id, company_name, legal_name, industry, description,
			email, phone, website,
			address_line1, address_line2, city, state, country, pincode,
			gstin, pan, tan, cin,
			pf_registration_number, pf_sub_code, esi_registration_number,
			pt_registration_number, lwf_registration_number, msme_registration_number,
			logo_url, logo_thumbnail_url, primary_color, secondary_color,
			settings, status, financial_year_start_month,
			created_at, updated_at, deleted_at
		FROM companies 
		WHERE owner_id = $1 AND deleted_at IS NULL`

	err := r.db.QueryRow(ctx, query, ownerID).Scan(
		&company.ID, &company.OwnerID, &company.CompanyName, &company.LegalName, &company.Industry, &company.Description,
		&company.Email, &company.Phone, &company.Website,
		&company.AddressLine1, &company.AddressLine2, &company.City, &company.State, &company.Country, &company.Pincode,
		&company.GSTIN, &company.PAN, &company.TAN, &company.CIN,
		&company.PFRegistrationNumber, &company.PFSubCode, &company.ESIRegistrationNumber,
		&company.PTRegistrationNumber, &company.LWFRegistrationNumber, &company.MSMERegistrationNumber,
		&company.LogoURL, &company.LogoThumbnailURL, &company.PrimaryColor, &company.SecondaryColor,
		&company.Settings, &company.Status, &company.FinancialYearStartMonth,
		&company.CreatedAt, &company.UpdatedAt, &company.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Company not found")
	}
	if err != nil {
		logger.Error("Failed to get company by owner ID", err, "owner_id", ownerID)
		return nil, errors.Internal(err)
	}

	return company, nil
}

func (r *CompanyRepository) List(ctx context.Context, limit, offset int) ([]*models.Company, int, error) {
	countQuery := `SELECT COUNT(*) FROM companies WHERE deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		logger.Error("Failed to count companies", err)
		return nil, 0, errors.Internal(err)
	}

	query := `
		SELECT id, owner_id, company_name, legal_name, industry, description,
			email, phone, website,
			address_line1, address_line2, city, state, country, pincode,
			gstin, pan, tan, cin,
			pf_registration_number, pf_sub_code, esi_registration_number,
			pt_registration_number, lwf_registration_number, msme_registration_number,
			logo_url, logo_thumbnail_url, primary_color, secondary_color,
			settings, status, financial_year_start_month,
			created_at, updated_at, deleted_at
		FROM companies 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		logger.Error("Failed to list companies", err)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var companies []*models.Company
	for rows.Next() {
		company := &models.Company{}
		if err := rows.Scan(
			&company.ID, &company.OwnerID, &company.CompanyName, &company.LegalName, &company.Industry, &company.Description,
			&company.Email, &company.Phone, &company.Website,
			&company.AddressLine1, &company.AddressLine2, &company.City, &company.State, &company.Country, &company.Pincode,
			&company.GSTIN, &company.PAN, &company.TAN, &company.CIN,
			&company.PFRegistrationNumber, &company.PFSubCode, &company.ESIRegistrationNumber,
			&company.PTRegistrationNumber, &company.LWFRegistrationNumber, &company.MSMERegistrationNumber,
			&company.LogoURL, &company.LogoThumbnailURL, &company.PrimaryColor, &company.SecondaryColor,
			&company.Settings, &company.Status, &company.FinancialYearStartMonth,
			&company.CreatedAt, &company.UpdatedAt, &company.DeletedAt,
		); err != nil {
			logger.Error("Failed to scan company", err)
			return nil, 0, errors.Internal(err)
		}
		companies = append(companies, company)
	}

	return companies, total, nil
}

func (r *CompanyRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE company_name = $1 AND deleted_at IS NULL)`
	if err := r.db.QueryRow(ctx, query, name).Scan(&exists); err != nil {
		logger.Error("Failed to check company name existence", err, "company_name", name)
		return false, errors.Internal(err)
	}
	return exists, nil
}

func (r *CompanyRepository) ExistsByGSTIN(ctx context.Context, gstin string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE gstin = $1 AND deleted_at IS NULL)`
	if err := r.db.QueryRow(ctx, query, gstin).Scan(&exists); err != nil {
		logger.Error("Failed to check GSTIN existence", err, "gstin", gstin)
		return false, errors.Internal(err)
	}
	return exists, nil
}

func (r *CompanyRepository) ExistsByPAN(ctx context.Context, pan string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM companies WHERE pan = $1 AND deleted_at IS NULL)`
	if err := r.db.QueryRow(ctx, query, pan).Scan(&exists); err != nil {
		logger.Error("Failed to check PAN existence", err, "pan", pan)
		return false, errors.Internal(err)
	}
	return exists, nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *CompanyRepository) Update(ctx context.Context, company *models.Company) error {
	company.UpdatedAt = time.Now()
	query := `
		UPDATE companies SET
			company_name = $2, legal_name = $3, industry = $4, description = $5,
			email = $6, phone = $7, website = $8,
			address_line1 = $9, address_line2 = $10, city = $11, state = $12, country = $13, pincode = $14,
			gstin = $15, pan = $16, tan = $17, cin = $18,
			pf_registration_number = $19, pf_sub_code = $20, esi_registration_number = $21,
			pt_registration_number = $22, lwf_registration_number = $23, msme_registration_number = $24,
			logo_url = $25, logo_thumbnail_url = $26, primary_color = $27, secondary_color = $28,
			settings = $29, status = $30, financial_year_start_month = $31,
			updated_at = $32
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query,
		company.ID, company.CompanyName, company.LegalName, company.Industry, company.Description,
		company.Email, company.Phone, company.Website,
		company.AddressLine1, company.AddressLine2, company.City, company.State, company.Country, company.Pincode,
		company.GSTIN, company.PAN, company.TAN, company.CIN,
		company.PFRegistrationNumber, company.PFSubCode, company.ESIRegistrationNumber,
		company.PTRegistrationNumber, company.LWFRegistrationNumber, company.MSMERegistrationNumber,
		company.LogoURL, company.LogoThumbnailURL, company.PrimaryColor, company.SecondaryColor,
		company.Settings, company.Status, company.FinancialYearStartMonth,
		company.UpdatedAt,
	)

	if err != nil {
		logger.Error("Failed to update company", err, "company_id", company.ID)
		return errors.Wrap(err, 500, "Failed to update company")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Company not found")
	}

	return nil
}

func (r *CompanyRepository) UpdateSettings(ctx context.Context, companyID uuid.UUID, settings models.CompanySettings) error {
	query := `UPDATE companies SET settings = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	now := time.Now()

	result, err := r.db.Exec(ctx, query, companyID, settings, now)
	if err != nil {
		logger.Error("Failed to update company settings", err, "company_id", companyID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Company not found")
	}

	return nil
}

func (r *CompanyRepository) UpdateStatus(ctx context.Context, companyID uuid.UUID, status models.CompanyStatus) error {
	query := `UPDATE companies SET status = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	now := time.Now()

	result, err := r.db.Exec(ctx, query, companyID, status, now)
	if err != nil {
		logger.Error("Failed to update company status", err, "company_id", companyID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Company not found")
	}

	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *CompanyRepository) SoftDelete(ctx context.Context, companyID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE companies SET deleted_at = $2, status = $3, updated_at = $4 WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, companyID, now, models.CompanyStatusInactive, now)
	if err != nil {
		logger.Error("Failed to soft delete company", err, "company_id", companyID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Company not found or already deleted")
	}

	return nil
}

func (r *CompanyRepository) Restore(ctx context.Context, companyID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE companies SET deleted_at = NULL, status = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NOT NULL`

	result, err := r.db.Exec(ctx, query, companyID, models.CompanyStatusActive, now)
	if err != nil {
		logger.Error("Failed to restore company", err, "company_id", companyID)
		return errors.Internal(err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Company not found or not deleted")
	}

	return nil
}

// ============================================================================
// STATISTICS
// ============================================================================

func (r *CompanyRepository) GetEmployeeCount(ctx context.Context, companyID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND deleted_at IS NULL`
	
	if err := r.db.QueryRow(ctx, query, companyID).Scan(&count); err != nil {
		logger.Error("Failed to get employee count", err, "company_id", companyID)
		return 0, errors.Internal(err)
	}

	return count, nil
}

func (r *CompanyRepository) GetActiveEmployeeCount(ctx context.Context, companyID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND status = $2 AND is_active = true AND deleted_at IS NULL`
	
	if err := r.db.QueryRow(ctx, query, companyID, models.UserStatusActive).Scan(&count); err != nil {
		logger.Error("Failed to get active employee count", err, "company_id", companyID)
		return 0, errors.Internal(err)
	}

	return count, nil
}