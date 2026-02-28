package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"project/internal/dto"
	"project/internal/repository"

	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

const (
	maxAPIKeysPerOrg   = 50
	bcryptCost         = 12
	keyRandomBytes     = 32
	keyPrefixLen       = 12 // chars shown after "whk_live_" for identification
)

// ============================================================================
// API Key Service
// ============================================================================

type APIKeyService struct {
	apiKeyRepo  *repository.APIKeyRepository
	projectRepo *repository.ProjectRepository
	log         *logger.Logger
}

func NewAPIKeyService(
	apiKeyRepo *repository.APIKeyRepository,
	projectRepo *repository.ProjectRepository,
	log *logger.Logger,
) *APIKeyService {
	return &APIKeyService{
		apiKeyRepo:  apiKeyRepo,
		projectRepo: projectRepo,
		log:         log,
	}
}

// ============================================================================
// CREATE API KEY
// ============================================================================

func (s *APIKeyService) CreateAPIKey(ctx context.Context, orgID, userID uuid.UUID, req *dto.CreateAPIKeyRequest) (*dto.CreateAPIKeyResponse, error) {
	// Validate
	if err := sharedUtils.ValidateRequired(req.Name, "name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if err := sharedUtils.ValidateLength(req.Name, 2, 100, "name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if !req.KeyType.IsValid() {
		return nil, errors.BadRequest("invalid key_type: must be api_key, cli_pat, or ci_cd")
	}
	if !req.Environment.IsValid() {
		return nil, errors.BadRequest("environment must be 'production' or 'development'")
	}
	if len(req.Scopes) == 0 {
		return nil, errors.BadRequest("at least one scope is required")
	}
	if err := validateScopes(req.Scopes); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// If project-scoped, verify project belongs to org
	if req.ProjectID != nil {
		if _, err := s.projectRepo.GetByIDAndOrg(ctx, *req.ProjectID, orgID); err != nil {
			return nil, errors.BadRequest("project not found in your organization")
		}
	}

	// Plan limit check
	count, err := s.apiKeyRepo.CountActiveByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if count >= maxAPIKeysPerOrg {
		return nil, errors.BadRequest(fmt.Sprintf("API key limit reached (%d/%d)", count, maxAPIKeysPerOrg))
	}

	// Generate key
	plainKey, prefix, hash, err := generateAPIKey(req.Environment)
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to generate API key: %w", err))
	}

	rateLimitPerMinute := 1000
	if req.RateLimitPerMinute != nil && *req.RateLimitPerMinute > 0 {
		rateLimitPerMinute = *req.RateLimitPerMinute
	}

	now := time.Now()
	key := &models.APIKey{
		ID:                 uuid.New(),
		OrganizationID:     orgID,
		ProjectID:          req.ProjectID,
		CreatedBy:          userID,
		Name:               req.Name,
		Description:        req.Description,
		KeyType:            req.KeyType,
		Environment:        req.Environment,
		KeyPrefix:          prefix,
		KeyHash:            hash,
		Scopes:             req.Scopes,
		RateLimitPerMinute: rateLimitPerMinute,
		ExpiresAt:          req.ExpiresAt,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := key.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.apiKeyRepo.Create(key); err != nil {
		return nil, err
	}

	s.log.Info("API key created",
		"key_id", key.ID,
		"key_type", key.KeyType,
		"prefix", prefix,
		"org_id", orgID,
		"created_by", userID,
	)

	return &dto.CreateAPIKeyResponse{
		APIKey:   dto.ToAPIKeyResponse(key),
		PlainKey: plainKey,
	}, nil
}

// ============================================================================
// LIST API KEYS
// ============================================================================

func (s *APIKeyService) ListAPIKeys(ctx context.Context, orgID uuid.UUID) (*dto.ListAPIKeysResponse, error) {
	keys, total, err := s.apiKeyRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.APIKeyResponse, 0, len(keys))
	for _, k := range keys {
		responses = append(responses, dto.ToAPIKeyResponse(k))
	}

	return &dto.ListAPIKeysResponse{
		APIKeys: responses,
		Total:   total,
	}, nil
}

// ============================================================================
// GET API KEY
// ============================================================================

func (s *APIKeyService) GetAPIKey(ctx context.Context, keyID, orgID uuid.UUID) (*dto.GetAPIKeyResponse, error) {
	key, err := s.apiKeyRepo.GetByIDAndOrg(ctx, keyID, orgID)
	if err != nil {
		return nil, err
	}

	return &dto.GetAPIKeyResponse{
		APIKey: dto.ToAPIKeyResponse(key),
	}, nil
}

// ============================================================================
// REVOKE API KEY
// ============================================================================

func (s *APIKeyService) RevokeAPIKey(ctx context.Context, keyID, orgID, userID uuid.UUID, req *dto.RevokeAPIKeyRequest) (*dto.RevokeAPIKeyResponse, error) {
	key, err := s.apiKeyRepo.GetByIDAndOrg(ctx, keyID, orgID)
	if err != nil {
		return nil, err
	}

	if key.IsRevoked() {
		return nil, errors.BadRequest("this API key is already revoked")
	}

	reason := req.Reason
	if reason == "" {
		reason = "Revoked by user"
	}

	if err := key.Revoke(userID, reason); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.apiKeyRepo.Revoke(key); err != nil {
		return nil, err
	}

	s.log.Info("API key revoked",
		"key_id", keyID,
		"org_id", orgID,
		"revoked_by", userID,
	)

	return &dto.RevokeAPIKeyResponse{
		APIKey:  dto.ToAPIKeyResponse(key),
		Message: "API key revoked successfully",
	}, nil
}

// ============================================================================
// DELETE API KEY
// ============================================================================

func (s *APIKeyService) DeleteAPIKey(ctx context.Context, keyID, orgID uuid.UUID) (*dto.DeleteAPIKeyResponse, error) {
	if _, err := s.apiKeyRepo.GetByIDAndOrg(ctx, keyID, orgID); err != nil {
		return nil, err
	}

	if err := s.apiKeyRepo.SoftDelete(keyID); err != nil {
		return nil, err
	}

	s.log.Info("API key deleted", "key_id", keyID, "org_id", orgID)

	return &dto.DeleteAPIKeyResponse{
		Message: "API key deleted successfully",
	}, nil
}

// ============================================================================
// KEY GENERATION
// ============================================================================

// generateAPIKey generates a new API key, returning (plainKey, prefix, bcryptHash, error)
// Format: whk_live_<hex> or whk_test_<hex>
func generateAPIKey(env models.ProjectEnvironment) (string, string, string, error) {
	b := make([]byte, keyRandomBytes)
	if _, err := rand.Read(b); err != nil {
		return "", "", "", err
	}

	random := hex.EncodeToString(b) // 64 chars

	var envPart string
	if env == models.ProjectEnvironmentProduction {
		envPart = "live"
	} else {
		envPart = "test"
	}

	// Full key: whk_live_<64 hex chars>
	plainKey := fmt.Sprintf("whk_%s_%s", envPart, random)

	// Prefix shown to user: "whk_live_" + first 12 chars + "..."
	prefix := fmt.Sprintf("whk_%s_%s...", envPart, random[:keyPrefixLen])

	// bcrypt hash of full key
	hash, err := bcrypt.GenerateFromPassword([]byte(plainKey), bcryptCost)
	if err != nil {
		return "", "", "", err
	}

	return plainKey, prefix, string(hash), nil
}

// ============================================================================
// SCOPE VALIDATION
// ============================================================================

func validateScopes(scopes []string) error {
	valid := map[string]bool{}
	for _, s := range models.AllScopes() {
		valid[string(s)] = true
	}
	for _, scope := range scopes {
		if !valid[scope] {
			return fmt.Errorf("invalid scope: %q", scope)
		}
	}
	return nil
}