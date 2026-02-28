package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"

	"project/internal/dto"
	"project/internal/repository"

	"shared/errors"
	"shared/logger"
	"shared/models"
)

// BaseDomain is the platform's base domain used for CNAME targets.
// This should be loaded from config in production.
const baseDomain = "hooks.yourdomain.com"

// maxDomainsPerOrg is the limit of custom domains per organization
const maxDomainsPerOrg = 10

// ============================================================================
// Domain Service
// ============================================================================

type DomainService struct {
	domainRepo  *repository.DomainRepository
	projectRepo *repository.ProjectRepository
	log         *logger.Logger
}

func NewDomainService(
	domainRepo *repository.DomainRepository,
	projectRepo *repository.ProjectRepository,
	log *logger.Logger,
) *DomainService {
	return &DomainService{
		domainRepo:  domainRepo,
		projectRepo: projectRepo,
		log:         log,
	}
}

// ============================================================================
// ADD DOMAIN
// ============================================================================

func (s *DomainService) AddDomain(ctx context.Context, orgID uuid.UUID, req *dto.AddDomainRequest) (*dto.AddDomainResponse, error) {
	// Normalize domain
	domain := strings.ToLower(strings.TrimSpace(req.Domain))

	// Basic format validation
	if domain == "" {
		return nil, errors.BadRequest("domain is required")
	}
	if strings.HasPrefix(domain, "http") {
		return nil, errors.BadRequest("domain must not include protocol (no http/https)")
	}
	if !strings.Contains(domain, ".") {
		return nil, errors.BadRequest("invalid domain format")
	}

	// Check global uniqueness — a domain can only be claimed by one org
	exists, err := s.domainRepo.ExistsByDomain(ctx, domain)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Conflict("this domain is already registered on this platform")
	}

	// Per-org limit check
	_, total, err := s.domainRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if total >= maxDomainsPerOrg {
		return nil, errors.BadRequest(fmt.Sprintf("domain limit reached (%d/%d)", total, maxDomainsPerOrg))
	}

	// If projectID provided, verify it belongs to this org
	if req.ProjectID != nil {
		if _, err := s.projectRepo.GetByIDAndOrg(ctx, *req.ProjectID, orgID); err != nil {
			return nil, errors.BadRequest("project not found in your organization")
		}
	}

	// Generate verification token
	token, err := generateVerificationToken()
	if err != nil {
		return nil, errors.Internal(fmt.Errorf("failed to generate verification token: %w", err))
	}

	now := time.Now()
	d := &models.CustomDomain{
		ID:                 uuid.New(),
		OrganizationID:     orgID,
		ProjectID:          req.ProjectID,
		Domain:             domain,
		DomainType:         models.DomainTypeInbound,
		VerificationToken:  token,
		VerificationStatus: models.DomainVerificationPending,
		SSLStatus:          models.SSLPending,
		SSLProvider:        "letsencrypt",
		SSLAutoRenew:       true,
		IsActive:           false,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := d.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.domainRepo.Create(d); err != nil {
		return nil, err
	}

	s.log.Info("Custom domain added",
		"domain_id", d.ID,
		"domain", domain,
		"org_id", orgID,
	)

	return &dto.AddDomainResponse{
		Domain: dto.ToDomainResponse(d),
	}, nil
}

// ============================================================================
// LIST DOMAINS
// ============================================================================

func (s *DomainService) ListDomains(ctx context.Context, orgID uuid.UUID) (*dto.ListDomainsResponse, error) {
	domains, total, err := s.domainRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	responses := make([]*dto.DomainResponse, 0, len(domains))
	for _, d := range domains {
		responses = append(responses, dto.ToDomainResponse(d))
	}

	return &dto.ListDomainsResponse{
		Domains: responses,
		Total:   total,
	}, nil
}

// ============================================================================
// GET DOMAIN
// ============================================================================

func (s *DomainService) GetDomain(ctx context.Context, domainID, orgID uuid.UUID) (*dto.GetDomainResponse, error) {
	d, err := s.domainRepo.GetByIDAndOrg(ctx, domainID, orgID)
	if err != nil {
		return nil, err
	}

	return &dto.GetDomainResponse{
		Domain:      dto.ToDomainResponse(d),
		CNAMETarget: d.GetCNAMETarget(baseDomain),
		Instruction: d.GetCNAMEInstruction(baseDomain),
	}, nil
}

// ============================================================================
// DELETE DOMAIN
// ============================================================================

func (s *DomainService) DeleteDomain(ctx context.Context, domainID, orgID uuid.UUID) (*dto.DeleteDomainResponse, error) {
	// Verify ownership before deleting
	if _, err := s.domainRepo.GetByIDAndOrg(ctx, domainID, orgID); err != nil {
		return nil, err
	}

	if err := s.domainRepo.SoftDelete(domainID); err != nil {
		return nil, err
	}

	s.log.Info("Custom domain deleted", "domain_id", domainID, "org_id", orgID)

	return &dto.DeleteDomainResponse{
		Message: "Domain removed successfully",
	}, nil
}

// ============================================================================
// TRIGGER VERIFICATION (manual DNS check)
// ============================================================================

// TriggerVerification allows the user to manually trigger a DNS check
// without waiting for the cron worker. Useful for immediate feedback.
func (s *DomainService) TriggerVerification(ctx context.Context, domainID, orgID uuid.UUID) (*dto.TriggerVerificationResponse, error) {
	d, err := s.domainRepo.GetByIDAndOrg(ctx, domainID, orgID)
	if err != nil {
		return nil, err
	}

	if d.IsDNSVerified() {
		return &dto.TriggerVerificationResponse{
			Domain:  dto.ToDomainResponse(d),
			Message: "Domain DNS is already verified",
		}, nil
	}

	// Perform actual DNS lookup
	verified, checkErr := s.checkCNAME(d)

	if verified {
		d.MarkDNSVerified()
		s.log.Info("DNS verified via manual trigger", "domain", d.Domain, "domain_id", d.ID)
	} else {
		errMsg := "CNAME record not found or does not point to the expected target"
		if checkErr != nil {
			errMsg = checkErr.Error()
		}
		d.MarkDNSFailed(errMsg)
		s.log.Info("DNS verification failed", "domain", d.Domain, "error", errMsg)
	}

	// Persist state change
	if err := s.domainRepo.UpdateVerificationStatus(d); err != nil {
		return nil, err
	}

	msg := "DNS check completed — CNAME not found yet. Please ensure your DNS record is set correctly."
	if verified {
		msg = "DNS verified! SSL provisioning has started — this may take a few minutes."
	}

	return &dto.TriggerVerificationResponse{
		Domain:  dto.ToDomainResponse(d),
		Message: msg,
	}, nil
}

// ============================================================================
// DNS CHECK LOGIC
// ============================================================================

// checkCNAME performs a real DNS CNAME lookup to verify domain ownership
func (s *DomainService) checkCNAME(d *models.CustomDomain) (bool, error) {
	expectedTarget := d.GetCNAMETarget(baseDomain)

	cname, err := net.LookupCNAME(d.Domain)
	if err != nil {
		return false, fmt.Errorf("DNS lookup failed: %w", err)
	}

	// net.LookupCNAME returns FQDN with trailing dot — normalize
	cname = strings.TrimSuffix(strings.ToLower(cname), ".")
	expectedTarget = strings.ToLower(expectedTarget)

	return cname == expectedTarget, nil
}

// ============================================================================
// HELPERS
// ============================================================================

func generateVerificationToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}