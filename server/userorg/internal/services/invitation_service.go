package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"userorg/internal/dto"
	"userorg/internal/repository"
	"shared/errors"
	"shared/logger"
	"shared/models"
	"shared/utils"
)

const (
	invitationTTL    = 7 * 24 * time.Hour // 7 days
	invitationTokenLen = 32               // bytes → 64 hex chars
)

// ============================================================================
// Invitation Service
// ============================================================================

type InvitationService struct {
	orgRepo          *repository.OrganizationRepository
	invitationRepo   *repository.InvitationRepository
	notificationRepo *repository.NotificationRepository
	userRepo         *repository.UserRepository
	log              *logger.Logger
}

func NewInvitationService(
	orgRepo *repository.OrganizationRepository,
	invitationRepo *repository.InvitationRepository,
	notificationRepo *repository.NotificationRepository,
	userRepo *repository.UserRepository,
	log *logger.Logger,
) *InvitationService {
	return &InvitationService{
		orgRepo:          orgRepo,
		invitationRepo:   invitationRepo,
		notificationRepo: notificationRepo,
		userRepo:         userRepo,
		log:              log,
	}
}

// ============================================================================
// INVITE MEMBER
// ============================================================================

func (s *InvitationService) InviteMember(
	ctx context.Context,
	orgID, callerID uuid.UUID,
	req *dto.InviteMemberRequest,
) (*dto.InvitationResponse, error) {

	// 1. Normalize email
	email := utils.NormalizeEmail(req.Email)
	if err := utils.ValidateEmail(email); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// 2. Validate role — owner cannot be invited
	if req.Role == models.RoleOwner || !req.Role.IsValid() {
		return nil, errors.BadRequest("Invalid role — owner role cannot be assigned via invitation")
	}

	// 3. Load org + verify caller membership
	org, err := s.orgRepo.GetByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	callerMember, err := s.orgRepo.GetMembership(ctx, orgID, callerID)
	if err != nil {
		return nil, errors.Forbidden("You are not a member of this organization")
	}
	if !callerMember.CanInviteMembers() {
		return nil, errors.Forbidden("Insufficient permissions to invite members")
	}

	// 4. Check seat limit
	memberCount, err := s.orgRepo.GetMemberCount(ctx, orgID)
	if err != nil {
		return nil, err
	}
	if memberCount >= org.TeamSeatsLimit {
		return nil, errors.BadRequest(fmt.Sprintf(
			"Team seat limit reached (%d/%d). Upgrade your plan to add more members.",
			memberCount, org.TeamSeatsLimit,
		))
	}

	// 5. Check if email is already a member
	inviteeUser, _ := s.userRepo.GetByEmail(ctx, email) // nil if not registered — that's fine
	if inviteeUser != nil {
		already, err := s.orgRepo.ExistsMembership(ctx, orgID, inviteeUser.ID)
		if err != nil {
			return nil, err
		}
		if already {
			return nil, errors.Conflict("This user is already a member of the organization")
		}
	}

	// 6. Check for duplicate pending invitation
	pending, err := s.invitationRepo.ExistsPendingForEmail(ctx, orgID, email)
	if err != nil {
		return nil, err
	}
	if pending {
		return nil, errors.Conflict("A pending invitation for this email already exists")
	}

	// 7. Load caller user for metadata
	callerUser, err := s.userRepo.GetByID(ctx, callerID)
	if err != nil {
		return nil, err
	}

	// 8. Generate secure token
	token, err := generateSecureToken()
	if err != nil {
		s.log.Error("Failed to generate invitation token", err)
		return nil, errors.Internal(err)
	}

	// 9. Build invitation
	now := time.Now()
	inv := &models.Invitation{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Email:          email,
		Role:           req.Role,
		InvitedBy:      callerID,
		Token:          token,
		Status:         models.InvitationPending,
		ExpiresAt:      now.Add(invitationTTL),
		Metadata: models.InvitationMetadata{
			CustomMessage: req.CustomMessage,
			InviterName:   callerUser.FullName,
			OrgName:       org.Name,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := inv.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	if err := s.invitationRepo.Create(ctx, inv); err != nil {
		return nil, err
	}

	// 10. Send notification to invitee (if they have an account)
	if inviteeUser != nil {
		actionURL := fmt.Sprintf("/invitations/accept/%s", token)
		actionLabel := "Accept Invitation"
		notification := buildNotification(
			inviteeUser.ID,
			models.NotificationInvitation,
			models.PriorityHigh,
			fmt.Sprintf("You've been invited to join %s", org.Name),
			fmt.Sprintf("%s invited you to join %s as %s.", callerUser.FullName, org.Name, req.Role),
			&actionURL, &actionLabel,
			&orgID, &inv.ID, nil,
			models.NotificationMetadata{
				InviterName: callerUser.FullName,
				OrgName:     org.Name,
			},
			&inv.ExpiresAt,
		)
		if err := s.notificationRepo.Create(ctx, notification); err != nil {
			// Non-fatal: log but don't fail the invitation
			s.log.Error("Failed to create invitation notification", err, "user_id", inviteeUser.ID)
		}
	}

	s.log.Info("Invitation sent", "invitation_id", inv.ID, "email", email, "org_id", orgID)
	return dto.ToInvitationResponse(inv), nil
}

// ============================================================================
// ACCEPT INVITATION
// ============================================================================

func (s *InvitationService) AcceptInvitation(
	ctx context.Context,
	callerID uuid.UUID,
	token string,
) (*dto.AcceptInvitationResponse, error) {

	// 1. Load invitation
	inv, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.NotFound("Invitation not found or already used")
	}

	// 2. Validate invite can be accepted
	if !inv.CanBeAccepted() {
		return nil, errors.BadRequest(fmt.Sprintf(
			"Invitation cannot be accepted (status: %s, expired: %v)",
			inv.Status, inv.IsExpired(),
		))
	}

	// 3. Load accepting user and verify email matches
	callerUser, err := s.userRepo.GetByID(ctx, callerID)
	if err != nil {
		return nil, err
	}
	if callerUser.Email != inv.Email {
		return nil, errors.Forbidden("This invitation was not sent to your email address")
	}

	// 4. Check not already a member
	already, err := s.orgRepo.ExistsMembership(ctx, inv.OrganizationID, callerID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, errors.Conflict("You are already a member of this organization")
	}

	// 5. Accept invitation model
	if err := inv.Accept(callerID); err != nil {
		return nil, errors.BadRequest(err.Error())
	}

	// 6. Persist invitation status
	if err := s.invitationRepo.UpdateStatus(ctx, inv); err != nil {
		return nil, err
	}

	// 7. Create membership
	now := time.Now()
	member := &models.OrganizationMember{
		ID:                   uuid.New(),
		OrganizationID:       inv.OrganizationID,
		UserID:               callerID,
		Role:                 inv.Role,
		InvitedBy:            &inv.InvitedBy,
		InvitationAcceptedAt: inv.AcceptedAt,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := member.Validate(); err != nil {
		return nil, errors.BadRequest(err.Error())
	}
	if err := s.orgRepo.CreateMembership(ctx, member); err != nil {
		return nil, err
	}

	// 8. Load org for response
	org, err := s.orgRepo.GetByID(ctx, inv.OrganizationID)
	if err != nil {
		return nil, err
	}

	// 9. Notify the inviter that the invitation was accepted
	inviterNotif := buildNotification(
		inv.InvitedBy,
		models.NotificationInvitationAccepted,
		models.PriorityNormal,
		fmt.Sprintf("%s joined %s", callerUser.FullName, org.Name),
		fmt.Sprintf("%s accepted your invitation and joined %s as %s.", callerUser.FullName, org.Name, inv.Role),
		nil, nil,
		&inv.OrganizationID, &inv.ID, nil,
		models.NotificationMetadata{
			OrgName: org.Name,
		},
		nil,
	)
	if err := s.notificationRepo.Create(ctx, inviterNotif); err != nil {
		s.log.Error("Failed to notify inviter", err, "inviter_id", inv.InvitedBy)
	}

	s.log.Info("Invitation accepted", "invitation_id", inv.ID, "user_id", callerID, "org_id", inv.OrganizationID)
	return &dto.AcceptInvitationResponse{
		Organization: dto.ToOrganizationResponse(org, string(member.Role)),
		Member:       dto.ToMemberResponse(member),
	}, nil
}

// ============================================================================
// REVOKE INVITATION
// ============================================================================

func (s *InvitationService) RevokeInvitation(
	ctx context.Context,
	orgID, callerID, invitationID uuid.UUID,
) error {
	// 1. Verify caller membership + permissions
	callerMember, err := s.orgRepo.GetMembership(ctx, orgID, callerID)
	if err != nil {
		return errors.Forbidden("You are not a member of this organization")
	}
	if !callerMember.CanManageMembers() {
		return errors.Forbidden("Insufficient permissions to revoke invitations")
	}

	// 2. Load invitation
	inv, err := s.invitationRepo.GetByID(ctx, invitationID)
	if err != nil {
		return err
	}
	if inv.OrganizationID != orgID {
		return errors.NotFound("Invitation not found")
	}

	// 3. Revoke
	if err := inv.Revoke(); err != nil {
		return errors.BadRequest(err.Error())
	}
	if err := s.invitationRepo.UpdateStatus(ctx, inv); err != nil {
		return err
	}

	s.log.Info("Invitation revoked", "invitation_id", invitationID, "revoked_by", callerID)
	return nil
}

// ============================================================================
// LIST INVITATIONS
// ============================================================================

func (s *InvitationService) ListInvitations(
	ctx context.Context,
	orgID, callerID uuid.UUID,
) (*dto.ListInvitationsResponse, error) {

	// Verify caller is a member
	_, err := s.orgRepo.GetMembership(ctx, orgID, callerID)
	if err != nil {
		return nil, errors.Forbidden("You are not a member of this organization")
	}

	invitations, total, err := s.invitationRepo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		resp = append(resp, dto.ToInvitationResponse(inv))
	}

	return &dto.ListInvitationsResponse{Invitations: resp, Total: total}, nil
}

// ============================================================================
// HELPERS
// ============================================================================

func generateSecureToken() (string, error) {
	b := make([]byte, invitationTokenLen)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func buildNotification(
	userID uuid.UUID,
	notifType models.NotificationType,
	priority models.NotificationPriority,
	title, message string,
	actionURL, actionLabel *string,
	relatedOrgID, relatedInvID, relatedSubID *uuid.UUID,
	metadata models.NotificationMetadata,
	expiresAt *time.Time,
) *models.Notification {
	now := time.Now()
	return &models.Notification{
		ID:                    uuid.New(),
		UserID:                userID,
		Type:                  notifType,
		Title:                 title,
		Message:               message,
		ActionURL:             actionURL,
		ActionLabel:           actionLabel,
		RelatedOrganizationID: relatedOrgID,
		RelatedInvitationID:   relatedInvID,
		RelatedSubscriptionID: relatedSubID,
		Metadata:              metadata,
		IsRead:                false,
		Priority:              priority,
		ExpiresAt:             expiresAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}