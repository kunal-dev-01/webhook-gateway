package services

import (
	"context"

	"authentication/internal/dto"
	"authentication/internal/repository"
	"authentication/internal/utils"

	"shared/auth"
	"shared/logger"
	"shared/config"
	"shared/errors"
	sharedUtils "shared/utils"
)

type AuthService struct {
	userRepo *repository.UserRepository
	organizationRepo *repository.OrganizationRepository
	tokenRepo *repository.TokenRepository
	sessionMgr *utils.Manager
	emailService *utils.EmailService
	jwtManager *auth.TokenManager
	cfg *config.Config
	log *logger.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	organizationRepo *repository.OrganizationRepository,
	tokenRepo *repository.TokenRepository,
	sessionMgr *utils.Manager,
	emailService *utils.EmailService,
	jwtManager *auth.TokenManager,
	cfg *config.Config,
	log *logger.Logger,
) *AuthService {
	return &AuthService{
		userRepo : userRepo,
		organizationRepo : organizationRepo,
		tokenRepo : tokenRepo,
		sessionMgr : sessionMgr,
		emailService : emailService,
		jwtManager : jwtManager,
		cfg : cfg,
		log : log,
	}
}

/* -------------------------- REGISTER ------------------------------ */

func (s *AuthService) Register(ctx context.Context, req *dto.RegisterRequest) error {
	if err := sharedUtils.ValidatePassword(req.Password); err != nil {
		return errors.BadRequest(err.Error())
	}

	if err := sharedUtils.ValidateEmail(req.Email); err != nil {
		return errors.BadRequest(err.Error())
	}

	if err := sharedUtils.ValidateMinLength(req.FullName, 3, "Full name"); err != nil {
		return errors.BadRequest(err.Error())
	}

	if err := sharedUtils.ValidateMinLength(req.Password, 50, "Full name"); err != nil {
		return errors.BadRequest(err.Error())
	}

	exists , err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return errors.Internal(err)
	}

	if exists {
		return errors.BadRequest("User already exists")
	}

	return nil

}