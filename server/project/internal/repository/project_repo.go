package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"project/internal/db"
	"shared/errors"
	"shared/logger"
	"shared/models"
)

// ============================================================================
// Project Repository
// ============================================================================

type ProjectRepository struct {
	db *db.DB
}

func NewProjectRepository(database *db.DB) *ProjectRepository {
	return &ProjectRepository{db: database}
}

func (r *ProjectRepository) withWriteContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func projectCols() string {
	return `id, organization_id, name, slug, description, signing_secret,
		environment, is_active, deleted_at, created_at, updated_at`
}

func (r *ProjectRepository) scanProject(row *sql.Row) (*models.Project, error) {
	p := &models.Project{}
	err := row.Scan(
		&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description, &p.SigningSecret,
		&p.Environment, &p.IsActive, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	return p, err
}

// ============================================================================
// CREATE
// ============================================================================

func (r *ProjectRepository) Create(p *models.Project) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		INSERT INTO projects (
			id, organization_id, name, slug, description,
			signing_secret, environment, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.Exec(ctx, query,
		p.ID, p.OrganizationID, p.Name, p.Slug, p.Description,
		p.SigningSecret, p.Environment, p.IsActive,
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		logger.Error("Failed to create project", err, "org_id", p.OrganizationID)
		return errors.Internal(err)
	}

	logger.Debug("Project created", "project_id", p.ID, "slug", p.Slug)
	return nil
}

// ============================================================================
// READ
// ============================================================================

func (r *ProjectRepository) GetByID(ctx context.Context, projectID uuid.UUID) (*models.Project, error) {
	query := fmt.Sprintf(`SELECT %s FROM projects WHERE id = $1 AND deleted_at IS NULL`, projectCols())

	p, err := r.scanProject(r.db.QueryRow(ctx, query, projectID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Project not found")
	}
	if err != nil {
		logger.Error("Failed to get project by ID", err, "project_id", projectID)
		return nil, errors.Internal(err)
	}
	return p, nil
}

// GetByIDAndOrg fetches project only if it belongs to the given org
func (r *ProjectRepository) GetByIDAndOrg(ctx context.Context, projectID, orgID uuid.UUID) (*models.Project, error) {
	query := fmt.Sprintf(`
		SELECT %s FROM projects
		WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL`, projectCols())

	p, err := r.scanProject(r.db.QueryRow(ctx, query, projectID, orgID))
	if err == sql.ErrNoRows {
		return nil, errors.NotFound("Project not found")
	}
	if err != nil {
		logger.Error("Failed to get project", err, "project_id", projectID, "org_id", orgID)
		return nil, errors.Internal(err)
	}
	return p, nil
}

func (r *ProjectRepository) ExistsBySlug(ctx context.Context, orgID uuid.UUID, slug string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM projects WHERE organization_id = $1 AND slug = $2 AND deleted_at IS NULL
	)`
	if err := r.db.QueryRow(ctx, query, orgID, slug).Scan(&exists); err != nil {
		return false, errors.Internal(err)
	}
	return exists, nil
}

func (r *ProjectRepository) ListByOrg(ctx context.Context, orgID uuid.UUID) ([]*models.Project, int, error) {
	countQuery := `SELECT COUNT(*) FROM projects WHERE organization_id = $1 AND deleted_at IS NULL`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, orgID).Scan(&total); err != nil {
		return nil, 0, errors.Internal(err)
	}

	listQuery := fmt.Sprintf(`
		SELECT %s FROM projects
		WHERE organization_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC`, projectCols())

	rows, err := r.db.Query(ctx, listQuery, orgID)
	if err != nil {
		logger.Error("Failed to list projects", err, "org_id", orgID)
		return nil, 0, errors.Internal(err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		p := &models.Project{}
		err := rows.Scan(
			&p.ID, &p.OrganizationID, &p.Name, &p.Slug, &p.Description, &p.SigningSecret,
			&p.Environment, &p.IsActive, &p.DeletedAt, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			logger.Error("Failed to scan project row", err)
			return nil, 0, errors.Internal(err)
		}
		projects = append(projects, p)
	}

	return projects, total, nil
}

// CountByOrg returns total active projects for an org (for plan limit checks)
func (r *ProjectRepository) CountByOrg(ctx context.Context, orgID uuid.UUID) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM projects WHERE organization_id = $1 AND deleted_at IS NULL`
	if err := r.db.QueryRow(ctx, query, orgID).Scan(&count); err != nil {
		return 0, errors.Internal(err)
	}
	return count, nil
}

// ============================================================================
// UPDATE
// ============================================================================

func (r *ProjectRepository) Update(projectID uuid.UUID, name string, description *string) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE projects
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.Exec(ctx, query, projectID, name, description, time.Now())
	if err != nil {
		logger.Error("Failed to update project", err, "project_id", projectID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Project not found")
	}
	return nil
}

func (r *ProjectRepository) SetActive(projectID uuid.UUID, isActive bool) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `UPDATE projects SET is_active = $2, updated_at = $3 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.Exec(ctx, query, projectID, isActive, time.Now())
	if err != nil {
		logger.Error("Failed to toggle project", err, "project_id", projectID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Project not found")
	}
	return nil
}

// ============================================================================
// DELETE
// ============================================================================

func (r *ProjectRepository) SoftDelete(projectID uuid.UUID, deletedSlug string) error {
	ctx, cancel := r.withWriteContext()
	defer cancel()

	query := `
		UPDATE projects
		SET deleted_at = $2, slug = $3, is_active = false, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`

	now := time.Now()
	result, err := r.db.Exec(ctx, query, projectID, now, deletedSlug, now)
	if err != nil {
		logger.Error("Failed to delete project", err, "project_id", projectID)
		return errors.Internal(err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NotFound("Project not found")
	}
	return nil
}