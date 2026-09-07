package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"go-echo-server-template/internal/platform/authn"
	"go-echo-server-template/internal/platform/config"
	"go-echo-server-template/internal/platform/httpx"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Store interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUserByEmail(ctx context.Context, email string) (User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (User, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) (Session, error)
	GetSession(ctx context.Context, id uuid.UUID) (Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	CreateAPIToken(ctx context.Context, arg CreateAPITokenParams) (APIToken, error)
	GetAPITokenByHash(ctx context.Context, hash string) (APIToken, error)
}

type User struct {
	UserID       uuid.UUID
	Email        string
	PasswordHash string
	Roles        []string
	CreatedAt    time.Time
}

type Session struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
}

type APIToken struct {
	TokenID   uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type CreateUserParams struct {
	UserID       uuid.UUID
	Email        string
	PasswordHash string
	Roles        []string
	CreatedAt    time.Time
}

type CreateSessionParams struct {
	SessionID uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	CreatedAt time.Time
}

type CreateAPITokenParams struct {
	TokenID   uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type Service struct {
	store Store
	cfg   config.AuthConfig
}

func NewService(store Store, cfg config.AuthConfig) *Service {
	return &Service{store: store, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	if email == "" || password == "" {
		return User{}, fmt.Errorf("%w: email and password required", httpx.ErrValidation)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.BcryptCost)
	if err != nil {
		return User{}, fmt.Errorf("%w: hash password", httpx.ErrInternal)
	}
	u, err := s.store.CreateUser(ctx, CreateUserParams{
		UserID: uuid.New(), Email: email, PasswordHash: string(hash),
		Roles: []string{"user"}, CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return User{}, fmt.Errorf("%w: create user", httpx.ErrConflict)
	}
	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, User, error) {
	u, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, User{}, fmt.Errorf("%w: invalid credentials", httpx.ErrUnauthorized)
		}
		return Session{}, User{}, fmt.Errorf("%w: lookup user", httpx.ErrInternal)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return Session{}, User{}, fmt.Errorf("%w: invalid credentials", httpx.ErrUnauthorized)
	}
	sess, err := s.store.CreateSession(ctx, CreateSessionParams{
		SessionID: uuid.New(), UserID: u.UserID,
		ExpiresAt: time.Now().UTC().Add(s.cfg.SessionTTL), CreatedAt: time.Now().UTC(),
	})
	if err != nil {
		return Session{}, User{}, fmt.Errorf("%w: create session", httpx.ErrInternal)
	}
	return sess, u, nil
}

func (s *Service) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.store.DeleteSession(ctx, sessionID)
}

func (s *Service) PrincipalFromSession(ctx context.Context, sessionID uuid.UUID) (authn.Principal, error) {
	sess, err := s.store.GetSession(ctx, sessionID)
	if err != nil {
		return authn.Anonymous(), fmt.Errorf("%w: session", httpx.ErrUnauthorized)
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		_ = s.store.DeleteSession(ctx, sessionID)
		return authn.Anonymous(), fmt.Errorf("%w: session expired", httpx.ErrUnauthorized)
	}
	u, err := s.store.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return authn.Anonymous(), fmt.Errorf("%w: user", httpx.ErrUnauthorized)
	}
	return authn.Principal{UserID: u.UserID.String(), Roles: u.Roles, AuthMethod: "cookie"}, nil
}

func (s *Service) CreateBearerToken(ctx context.Context, userID uuid.UUID) (plain string, err error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	plain = hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(plain))
	_, err = s.store.CreateAPIToken(ctx, CreateAPITokenParams{
		TokenID: uuid.New(), UserID: userID, TokenHash: hex.EncodeToString(sum[:]),
		ExpiresAt: time.Now().UTC().Add(s.cfg.TokenTTL), CreatedAt: time.Now().UTC(),
	})
	return plain, err
}

func (s *Service) PrincipalFromBearer(ctx context.Context, token string) (authn.Principal, error) {
	sum := sha256.Sum256([]byte(token))
	row, err := s.store.GetAPITokenByHash(ctx, hex.EncodeToString(sum[:]))
	if err != nil {
		return authn.Anonymous(), fmt.Errorf("%w: token", httpx.ErrUnauthorized)
	}
	if time.Now().UTC().After(row.ExpiresAt) {
		return authn.Anonymous(), fmt.Errorf("%w: token expired", httpx.ErrUnauthorized)
	}
	u, err := s.store.GetUserByID(ctx, row.UserID)
	if err != nil {
		return authn.Anonymous(), fmt.Errorf("%w: user", httpx.ErrUnauthorized)
	}
	return authn.Principal{UserID: u.UserID.String(), Roles: u.Roles, AuthMethod: "bearer"}, nil
}

func (s *Service) Disabled() bool { return s.cfg.Disabled }
