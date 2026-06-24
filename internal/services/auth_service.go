package services

import (
	"context"
	"errors"
	"time"

	"github.com/phyowaiyan-dev/goappmon/internal/models"
	"github.com/phyowaiyan-dev/goappmon/internal/repositories"
	"github.com/phyowaiyan-dev/goappmon/internal/utils"
)

var ErrInvalidCredentials = errors.New("invalid credentials")

type AuthService struct {
	admins        *repositories.AdminRepository
	sessionSecret []byte
	sessionTTL    time.Duration
}

func NewAuthService(admins *repositories.AdminRepository, sessionSecret []byte, ttl time.Duration) *AuthService {
	return &AuthService{admins: admins, sessionSecret: sessionSecret, sessionTTL: ttl}
}

func (s *AuthService) Authenticate(ctx context.Context, email, password string) (*models.Admin, error) {
	admin, err := s.admins.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repositories.ErrAdminNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}
	if err := utils.CheckPassword(admin.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return admin, nil
}

func (s *AuthService) SignSession(adminID int64) (string, error) {
	return utils.SignSession(s.sessionSecret, adminID, s.sessionTTL)
}

func (s *AuthService) VerifySession(token string) (int64, error) {
	claims, err := utils.VerifySession(s.sessionSecret, token)
	if err != nil {
		return 0, err
	}
	return claims.AdminID, nil
}
