package postgres

import (
	"context"

	"go-echo-server-template/internal/auth"

	"github.com/google/uuid"
)

type Store struct {
	q *Queries
}

func NewStore(db DBTX) *Store {
	return &Store{q: New(db)}
}

func (s *Store) CreateUser(ctx context.Context, arg auth.CreateUserParams) (auth.User, error) {
	row, err := s.q.CreateUser(ctx, CreateUserParams{
		UserID: arg.UserID, Email: arg.Email, PasswordHash: arg.PasswordHash,
		Roles: arg.Roles, CreatedAt: arg.CreatedAt,
	})
	if err != nil {
		return auth.User{}, err
	}
	return mapUser(row), nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (auth.User, error) {
	row, err := s.q.GetUserByEmail(ctx, email)
	if err != nil {
		return auth.User{}, err
	}
	return mapUser(row), nil
}

func (s *Store) GetUserByID(ctx context.Context, id uuid.UUID) (auth.User, error) {
	row, err := s.q.GetUserByID(ctx, id)
	if err != nil {
		return auth.User{}, err
	}
	return mapUser(row), nil
}

func (s *Store) CreateSession(ctx context.Context, arg auth.CreateSessionParams) (auth.Session, error) {
	row, err := s.q.CreateSession(ctx, CreateSessionParams{
		SessionID: arg.SessionID, UserID: arg.UserID, ExpiresAt: arg.ExpiresAt, CreatedAt: arg.CreatedAt,
	})
	if err != nil {
		return auth.Session{}, err
	}
	return auth.Session{SessionID: row.SessionID, UserID: row.UserID, ExpiresAt: row.ExpiresAt}, nil
}

func (s *Store) GetSession(ctx context.Context, id uuid.UUID) (auth.Session, error) {
	row, err := s.q.GetSession(ctx, id)
	if err != nil {
		return auth.Session{}, err
	}
	return auth.Session{SessionID: row.SessionID, UserID: row.UserID, ExpiresAt: row.ExpiresAt}, nil
}

func (s *Store) DeleteSession(ctx context.Context, id uuid.UUID) error {
	return s.q.DeleteSession(ctx, id)
}

func (s *Store) CreateAPIToken(ctx context.Context, arg auth.CreateAPITokenParams) (auth.APIToken, error) {
	row, err := s.q.CreateAPIToken(ctx, CreateAPITokenParams{
		TokenID: arg.TokenID, UserID: arg.UserID, TokenHash: arg.TokenHash,
		ExpiresAt: arg.ExpiresAt, CreatedAt: arg.CreatedAt,
	})
	if err != nil {
		return auth.APIToken{}, err
	}
	return auth.APIToken{TokenID: row.TokenID, UserID: row.UserID, TokenHash: row.TokenHash, ExpiresAt: row.ExpiresAt}, nil
}

func (s *Store) GetAPITokenByHash(ctx context.Context, hash string) (auth.APIToken, error) {
	row, err := s.q.GetAPITokenByHash(ctx, hash)
	if err != nil {
		return auth.APIToken{}, err
	}
	return auth.APIToken{TokenID: row.TokenID, UserID: row.UserID, TokenHash: row.TokenHash, ExpiresAt: row.ExpiresAt}, nil
}

func mapUser(u AuthUser) auth.User {
	roles := u.Roles
	if roles == nil {
		roles = []string{}
	}
	return auth.User{
		UserID: u.UserID, Email: u.Email, PasswordHash: u.PasswordHash,
		Roles: roles, CreatedAt: u.CreatedAt,
	}
}
