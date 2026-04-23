package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/gjtiquia/vimance/internal/db"
)

func (s *Service) CreateUser(ctx context.Context, username string) (db.User, error) {
	now := time.Now().Unix()
	return s.queries.CreateUser(ctx, db.CreateUserParams{
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	})
}

func (s *Service) GetUser(ctx context.Context, id int64) (db.User, error) {
	return s.queries.GetUser(ctx, id)
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (db.User, error) {
	return s.queries.GetUserByUsername(ctx, username)
}

func (s *Service) ListUsers(ctx context.Context) ([]db.User, error) {
	return s.queries.ListUsers(ctx)
}

func (s *Service) ListActiveUsers(ctx context.Context) ([]db.ActiveUser, error) {
	return s.queries.ListActiveUsers(ctx)
}

func (s *Service) UpdateUser(ctx context.Context, id int64, username string) (db.User, error) {
	return s.queries.UpdateUser(ctx, db.UpdateUserParams{
		ID:        id,
		Username:  username,
		UpdatedAt: time.Now().Unix(),
	})
}

func (s *Service) SoftDeleteUser(ctx context.Context, id int64, deletedBy int64) error {
	return s.queries.SoftDeleteUser(ctx, db.SoftDeleteUserParams{
		ID:        id,
		DeletedAt: sql.NullInt64{Int64: time.Now().Unix(), Valid: true},
		DeletedBy: sql.NullInt64{Int64: deletedBy, Valid: true},
	})
}

func (s *Service) RestoreUser(ctx context.Context, id int64) (db.User, error) {
	return s.queries.RestoreUser(ctx, id)
}

func (s *Service) HardDeleteUser(ctx context.Context, id int64) error {
	return s.queries.HardDeleteUser(ctx, id)
}
