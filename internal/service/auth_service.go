package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/falashlion/urban-sanctuary-api/internal/config"
	"github.com/falashlion/urban-sanctuary-api/internal/domain"
	"github.com/falashlion/urban-sanctuary-api/internal/dto"
	"github.com/falashlion/urban-sanctuary-api/internal/platform/cache"
	"github.com/falashlion/urban-sanctuary-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo *repository.UserRepository
	authRepo *repository.AuthRepository
	cache    *cache.RedisClient
	cfg      *config.Config
	log      zerolog.Logger
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepository,
	authRepo *repository.AuthRepository,
	cache *cache.RedisClient,
	cfg *config.Config,
	log zerolog.Logger,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		authRepo: authRepo,
		cache:    cache,
		cfg:      cfg,
		log:      log,
	}
}

// Register creates a new user account.
func (s *AuthService) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Validate at least one identifier is provided
	if req.Email == nil && req.Phone == nil {
		return nil, domain.ErrValidation([]domain.ErrDetail{
			{Field: "email", Message: "email or phone is required"},
		})
	}

	// Check for existing user
	if req.Email != nil {
		existing, err := s.userRepo.GetByEmail(ctx, *req.Email)
		if err != nil {
			return nil, domain.ErrInternal(err)
		}
		if existing != nil {
			return nil, domain.ErrConflict("A user with this email already exists")
		}
	}
	if req.Phone != nil {
		existing, err := s.userRepo.GetByPhone(ctx, *req.Phone)
		if err != nil {
			return nil, domain.ErrInternal(err)
		}
		if existing != nil {
			return nil, domain.ErrConflict("A user with this phone number already exists")
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, domain.ErrInternal(fmt.Errorf("failed to hash password: %w", err))
	}

	user := &domain.User{
		Email:        req.Email,
		Phone:        req.Phone,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         domain.RoleGuest,
		IsVerified:   false,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, domain.ErrInternal(fmt.Errorf("failed to create user: %w", err))
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID, nil)
	if err != nil {
		return nil, domain.ErrInternal(err)
	}

	_ = refreshToken // refresh token is set via cookie by the handler

	return &dto.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.cfg.JWT.AccessTTL.Seconds()),
		User:        toUserResponse(user),
	}, nil
}

// Login authenticates a user with email/phone and password.
func (s *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, string, error) {
	var user *domain.User
	var err error

	if req.Email != nil {
		user, err = s.userRepo.GetByEmail(ctx, *req.Email)
	} else if req.Phone != nil {
		user, err = s.userRepo.GetByPhone(ctx, *req.Phone)
	} else {
		return nil, "", domain.ErrValidation([]domain.ErrDetail{
			{Field: "email", Message: "email or phone is required"},
		})
	}

	if err != nil {
		return nil, "", domain.ErrInternal(err)
	}
	if user == nil {
		return nil, "", domain.ErrUnauthorized()
	}

	// Check account lockout
	lockKey := fmt.Sprintf("login_lockout:%s", user.ID.String())
	locked, _ := s.cache.Exists(ctx, lockKey)
	if locked {
		return nil, "", domain.ErrRateLimited()
	}

	// Check if account is active
	if !user.IsActive {
		return nil, "", domain.ErrForbidden()
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed attempts
		attemptsKey := fmt.Sprintf("login_attempts:%s", user.ID.String())
		attempts, _ := s.cache.Incr(ctx, attemptsKey)
		_ = s.cache.Expire(ctx, attemptsKey, 15*time.Minute)
		if attempts >= 5 {
			_ = s.cache.Set(ctx, lockKey, "locked", 15*time.Minute)
			_ = s.cache.Del(ctx, attemptsKey)
		}
		return nil, "", domain.ErrUnauthorized()
	}

	// Clear failed attempts on successful login
	attemptsKey := fmt.Sprintf("login_attempts:%s", user.ID.String())
	_ = s.cache.Del(ctx, attemptsKey)

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", domain.ErrInternal(err)
	}

	refreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID, nil)
	if err != nil {
		return nil, "", domain.ErrInternal(err)
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.cfg.JWT.AccessTTL.Seconds()),
		User:        toUserResponse(user),
	}, refreshToken, nil
}

// RefreshToken rotates the refresh token and issues a new access token.
func (s *AuthService) RefreshToken(ctx context.Context, oldToken string) (*dto.AuthResponse, string, error) {
	tokenHash := repository.HashToken(oldToken)

	storedToken, err := s.authRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, "", domain.ErrInternal(err)
	}
	if storedToken == nil {
		return nil, "", domain.ErrUnauthorized()
	}
	if time.Now().After(storedToken.ExpiresAt) {
		return nil, "", domain.ErrUnauthorized()
	}

	// Revoke old token
	if err := s.authRepo.RevokeRefreshToken(ctx, storedToken.ID); err != nil {
		return nil, "", domain.ErrInternal(err)
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, storedToken.UserID)
	if err != nil || user == nil {
		return nil, "", domain.ErrUnauthorized()
	}

	// Generate new tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, "", domain.ErrInternal(err)
	}

	newRefreshToken, err := s.generateAndStoreRefreshToken(ctx, user.ID, storedToken.DeviceInfo)
	if err != nil {
		return nil, "", domain.ErrInternal(err)
	}

	return &dto.AuthResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.cfg.JWT.AccessTTL.Seconds()),
		User:        toUserResponse(user),
	}, newRefreshToken, nil
}

// Logout revokes the refresh token.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	tokenHash := repository.HashToken(refreshToken)
	storedToken, err := s.authRepo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return domain.ErrInternal(err)
	}
	if storedToken == nil {
		return nil
	}
	return s.authRepo.RevokeRefreshToken(ctx, storedToken.ID)
}

// RequestOTP generates and stores a 6-digit OTP code.
func (s *AuthService) RequestOTP(ctx context.Context, req dto.OTPRequestDTO) error {
	code, err := generateOTPCode()
	if err != nil {
		return domain.ErrInternal(err)
	}

	hashedCode, err := bcrypt.GenerateFromPassword([]byte(code), 12)
	if err != nil {
		return domain.ErrInternal(err)
	}

	otp := &domain.OTPCode{
		Identifier: req.Identifier,
		CodeHash:   string(hashedCode),
		Purpose:    domain.OTPPurpose(req.Purpose),
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}

	if err := s.authRepo.CreateOTPCode(ctx, otp); err != nil {
		return domain.ErrInternal(err)
	}

	// TODO: Send OTP via SMS or email
	s.log.Info().
		Str("identifier", req.Identifier).
		Str("code", code). // Only log in development!
		Msg("OTP generated")

	return nil
}

// VerifyOTP validates an OTP code.
func (s *AuthService) VerifyOTP(ctx context.Context, req dto.OTPVerifyRequest) error {
	otp, err := s.authRepo.GetLatestOTP(ctx, req.Identifier, domain.OTPPurpose(req.Purpose))
	if err != nil {
		return domain.ErrInternal(err)
	}
	if otp == nil {
		return domain.ErrOTPExpired()
	}
	if otp.Attempts >= 3 {
		return domain.ErrRateLimited()
	}

	// Increment attempts
	_ = s.authRepo.IncrementOTPAttempts(ctx, otp.ID)

	// Verify code
	if err := bcrypt.CompareHashAndPassword([]byte(otp.CodeHash), []byte(req.Code)); err != nil {
		return domain.ErrOTPInvalid()
	}

	// Mark as used
	_ = s.authRepo.MarkOTPUsed(ctx, otp.ID)

	return nil
}

// PasswordResetRequest initiates a password reset flow.
func (s *AuthService) PasswordResetRequest(ctx context.Context, req dto.PasswordResetRequestDTO) error {
	var identifier string
	if req.Email != nil {
		identifier = *req.Email
	} else if req.Phone != nil {
		identifier = *req.Phone
	} else {
		return domain.ErrValidation([]domain.ErrDetail{
			{Field: "email", Message: "email or phone is required"},
		})
	}

	return s.RequestOTP(ctx, dto.OTPRequestDTO{
		Identifier: identifier,
		Purpose:    string(domain.OTPPurposePasswordReset),
	})
}

// ResetPassword sets a new password after OTP verification.
func (s *AuthService) ResetPassword(ctx context.Context, req dto.PasswordResetDTO) error {
	// The token here is used to identify the user and verify the reset was authorized
	// In a full implementation, this would be a signed JWT or verified OTP session

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		return domain.ErrInternal(err)
	}

	// Parse the token to get user ID (simplified — in production use a proper reset token)
	userID, err := uuid.Parse(req.Token)
	if err != nil {
		return domain.ErrBadRequest("Invalid reset token")
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword)); err != nil {
		return domain.ErrInternal(err)
	}

	// Revoke all refresh tokens
	_ = s.authRepo.RevokeAllUserTokens(ctx, userID)

	return nil
}

// --- Internal helpers ---

func (s *AuthService) generateAccessToken(user *domain.User) (string, error) {
	var emailStr string
	if user.Email != nil {
		emailStr = *user.Email
	}

	claims := jwt.MapClaims{
		"sub":      user.ID.String(),
		"role":     string(user.Role),
		"email":    emailStr,
		"verified": user.IsVerified,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(s.cfg.JWT.AccessTTL).Unix(),
		"jti":      uuid.New().String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.AccessSecret))
}

func (s *AuthService) generateAndStoreRefreshToken(ctx context.Context, userID uuid.UUID, deviceInfo *string) (string, error) {
	rawToken := uuid.New().String()
	tokenHash := repository.HashToken(rawToken)

	rt := &domain.RefreshToken{
		UserID:     userID,
		TokenHash:  tokenHash,
		ExpiresAt:  time.Now().Add(s.cfg.JWT.RefreshTTL),
		DeviceInfo: deviceInfo,
	}

	if err := s.authRepo.CreateRefreshToken(ctx, rt); err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return rawToken, nil
}

func generateOTPCode() (string, error) {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func toUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:            u.ID.String(),
		Email:         u.Email,
		Phone:         u.Phone,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		AvatarURL:     u.AvatarURL,
		Role:          string(u.Role),
		IsVerified:    u.IsVerified,
		TOTPEnabled:   u.TOTPEnabled,
		LoyaltyPoints: u.LoyaltyPoints,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
	}
}
