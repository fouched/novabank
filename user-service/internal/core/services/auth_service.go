package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"time"

	"github.com/fouched/novabank/users/internal/core/errdefs"
	"github.com/fouched/novabank/users/internal/core/models"
	"github.com/fouched/novabank/users/internal/repositories"
	"github.com/fouched/novabank/users/platform/config"
	"github.com/fouched/novabank/users/platform/security"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  repositories.UserRepository
	accRepo   repositories.AccountRepository
	roleRepo  repositories.RoleRepository
	publisher *AccountEventPublisher
	cfg       *config.Config
	logger    *slog.Logger
}

func NewAuthService(
	uRepo repositories.UserRepository,
	aRepo repositories.AccountRepository,
	rRepo repositories.RoleRepository,
	pub *AccountEventPublisher,
	cfg *config.Config,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		userRepo:  uRepo,
		accRepo:   aRepo,
		roleRepo:  rRepo,
		publisher: pub,
		cfg:       cfg,
		logger:    logger,
	}
}

func (s *AuthService) Register(ctx context.Context, req models.RegistrationRequest) (*models.RegistrationResponse, error) {
	// 1. Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed checking account uniqueness maps: %w", err)
	}
	if exists {
		return nil, errdefs.NewBadRequest("a user profile with that email address already exists")
	}

	// 2. Fetch default operational role domain reference mapping parameters
	roleName := "CUSTOMER"
	if req.Role != "" {
		roleName = req.Role
	}
	dbRole, err := s.roleRepo.FindByName(ctx, roleName)
	if err != nil || dbRole == nil {
		return nil, errdefs.NewBadRequest("requested security operational role assignment profile is invalid")
	}

	// 3. Encrypt the incoming user credential via secure Bcrypt routine hashes
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed encoding password safety sequence: %w", err)
	}

	// 4. Create User record
	newUser := &models.User{
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Enabled:   true,
		Roles:     []models.Role{*dbRole},
		Accounts:  make([]models.Account, 1),
	}

	// 5. Generate and validate a unique, strictly numeric bank account number
	var generatedAccNo string
	for {
		// Generates an 11-digit random number sequence padded with leading zeros if necessary
		// 10000000000 + random offset guarantees it always hits an exact 11-digit ceiling scale
		randomDigits := 10000000000 + rand.Int63n(8999999999)
		generatedAccNo = fmt.Sprintf("%011d", randomDigits)

		accExists, err := s.accRepo.ExistsByAccountNumber(ctx, generatedAccNo)
		if err != nil {
			return nil, fmt.Errorf("failed auditing account number uniqueness constraints: %w", err)
		}
		if !accExists {
			break
		}
		s.logger.Warn("Collision detected on random numeric account generation, recalculating sequence token...",
			slog.String("collided_number", generatedAccNo),
		)
	}

	// 6. Instantiate initial multi-portfolio tracking Account profile record state parameters
	newAccount := models.Account{
		AccountNumber: generatedAccNo,
		Balance:       0.0000,
		Currency:      models.CurrencyZAR,
		AccountType:   models.TypeCurrent,
		AccountStatus: models.AccountActive,
	}

	newUser.Accounts[0] = newAccount
	if err := s.userRepo.Insert(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed persisting target user identity profile: %w", err)
	}

	// 7. Fire and forget the event payload onto the streaming Kafka pipeline (Context Preservation)
	go func(asyncCtx context.Context, evt models.UserRegistrationEvent) {
		s.publisher.PublishUserRegistrationEvent(asyncCtx, evt)
	}(context.WithoutCancel(ctx), models.UserRegistrationEvent{
		Email:         newUser.Email,
		FirstName:     newUser.FirstName,
		LastName:      newUser.LastName,
		AccountNumber: newAccount.AccountNumber,
		BankName:      "Nova Bank, Ltd.",
	})

	// 8. Generate security validation access tokens to establish authentic sessions instantly
	token, err := security.GenerateToken(newUser, s.cfg.Authority.JWTSecret, time.Duration(s.cfg.Authority.TokenExpiry))
	if err != nil {
		return nil, fmt.Errorf("failed encoding verification session tokens: %w", err)
	}

	return &models.RegistrationResponse{
		Token: token,
		User:  *newUser,
	}, nil
}

// dummyHash is compared against when the email is unknown, so a failed lookup
// takes as long as a wrong password and response timing doesn't reveal which
// emails are registered.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("novabank-dummy-password"), bcrypt.DefaultCost)

func (s *AuthService) Login(ctx context.Context, req models.LoginRequest) (*models.LoginResponse, error) {
	invalidCredentials := errdefs.NewUnauthorized("invalid email or password")

	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, errdefs.ErrNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(req.Password))
			return nil, invalidCredentials
		}
		return nil, fmt.Errorf("login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, invalidCredentials
	}

	// Checked only after the password matches, so this doesn't leak account state to guessers.
	if !user.Enabled {
		return nil, errdefs.NewUnauthorized("account is disabled")
	}

	token, err := security.GenerateToken(user, s.cfg.Authority.JWTSecret, time.Duration(s.cfg.Authority.TokenExpiry))
	if err != nil {
		return nil, fmt.Errorf("login: generate token: %w", err)
	}

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}
