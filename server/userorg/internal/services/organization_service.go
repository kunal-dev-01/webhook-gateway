package services

import (
	"context"

	"github.com/google/uuid"

	"userorg/internal/dto"
	"userorg/internal/repository"

	"shared/errors"
	"shared/logger"
	"shared/models"
	sharedUtils "shared/utils"
)

type OrganizationService struct {
	orgRepo *repository.OrganizationRepository
	log     *logger.Logger
}

func NewOrganizationService(
	orgRepo *repository.OrganizationRepository,
	log *logger.Logger,
) *OrganizationService {
	return &OrganizationService{
		orgRepo: orgRepo,
		log:     log,
	}
}

// ============================================================================
// LIST MY ORGANIZATIONS
// ============================================================================

func (s *OrganizationService) ListMyOrganizations(ctx context.Context, userID uuid.UUID) (*dto.ListOrganizationsResponse, error) {
	orgs, err := s.orgRepo.ListByUser(ctx, userID)
	if err != nil {
		s.log.Error("Failed to list organizations", err, "user_id", userID)
		return nil, err
	}

	// For each org, get the user's role in it
	responses := make([]*dto.OrganizationResponse, 0, len(orgs))
	for _, org := range orgs {
		member, err := s.orgRepo.GetMembership(ctx, org.ID, userID)
		if err != nil {
			s.log.Error("Failed to get membership for org", err, "org_id", org.ID, "user_id", userID)
			continue
		}
		responses = append(responses, dto.ToOrganizationResponse(org, string(member.Role)))
	}

	return &dto.ListOrganizationsResponse{
		Organizations: responses,
		Total:         len(responses),
	}, nil
}

// ============================================================================
// GET ORGANIZATION DETAILS
// ============================================================================

func (s *OrganizationService) GetOrganization(ctx context.Context, orgID, userID uuid.UUID) (*dto.GetOrganizationResponse, error) {
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	member, err := s.orgRepo.GetMembership(ctx, orgID, userID)
	if err != nil {
		return nil, errors.Forbidden("You are not a member of this organization")
	}

	return &dto.GetOrganizationResponse{
		Organization: dto.ToOrganizationResponse(org, string(member.Role)),
	}, nil
}

// ============================================================================
// RENAME ORGANIZATION
// ============================================================================
func (s *OrganizationService) UpdateName(ctx context.Context, orgID, userID uuid.UUID, name string) (*dto.GetOrganizationResponse, error) {
	if err := sharedUtils.ValidateLength(name, 2, 255, "organization name"); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// Caller must be owner to rename
	member, err := s.orgRepo.GetMembership(ctx, orgID, userID)
	if err != nil {
		return nil, errors.Forbidden("You are not a member of this organization")
	}
	if member.Role != models.RoleOwner {
		return nil, errors.Forbidden("Only the owner can rename the organization")
	}

	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.Name = name
	if err := s.orgRepo.Update(ctx, org); err != nil {
		s.log.Error("Failed to update organization name", err, "org_id", orgID)
		return nil, err
	}

	return &dto.GetOrganizationResponse{
		Organization: dto.ToOrganizationResponse(org, string(member.Role)),
	}, nil
}

// ============================================================================
// LIST MEMBERS
// ============================================================================

func (s *OrganizationService) ListMembers(ctx context.Context, orgID, userID uuid.UUID, req *dto.ListMembersRequest) (*dto.ListMembersResponse, error) {
	// Verify caller is a member
	if _, err := s.orgRepo.GetMembership(ctx, orgID, userID); err != nil {
		return nil, errors.Forbidden("You are not a member of this organization")
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := max(req.Offset, 0)

	members, total, err := s.orgRepo.ListMembers(ctx, orgID, limit, offset)
	if err != nil {
		s.log.Error("Failed to list members", err, "org_id", orgID)
		return nil, err
	}

	responses := make([]*dto.MemberResponse, 0, len(members))
	for _, m := range members {
		responses = append(responses, dto.ToMemberResponse(m))
	}

	return &dto.ListMembersResponse{
		Members: responses,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}, nil
}

// ============================================================================
// REMOVE MEMBER
// ============================================================================

func (s *OrganizationService) RemoveMember(ctx context.Context, orgID, callerID, targetUserID uuid.UUID) error {
	// Caller must be admin or owner
	callerMember, err := s.orgRepo.GetMembership(ctx, orgID, callerID)
	if err != nil {
		return errors.Forbidden("You are not a member of this organization")
	}

	if callerMember.Role != models.RoleOwner && callerMember.Role != models.RoleAdmin {
		return errors.Forbidden("Only owners and admins can remove members")
	}

	// Cannot remove yourself
	if callerID == targetUserID {
		return errors.BadRequest("You cannot remove yourself from the organization")
	}

	// Check target membership exists
	targetMember, err := s.orgRepo.GetMembership(ctx, orgID, targetUserID)
	if err != nil {
		return errors.NotFound("Member not found in this organization")
	}

	// Admin cannot remove owner
	if targetMember.Role == models.RoleOwner {
		return errors.Forbidden("Owner cannot be removed from the organization")
	}

	// Admin cannot remove another admin (only owner can)
	if callerMember.Role == models.RoleAdmin && targetMember.Role == models.RoleAdmin {
		return errors.Forbidden("Admins cannot remove other admins")
	}

	if err := s.orgRepo.RemoveMember(ctx, orgID, targetUserID); err != nil {
		s.log.Error("Failed to remove member", err, "org_id", orgID, "target_user_id", targetUserID)
		return err
	}

	s.log.Info("Member removed", "org_id", orgID, "removed_by", callerID, "removed_user", targetUserID)
	return nil
}