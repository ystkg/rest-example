package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ystkg/rest-example/entity"
	"gorm.io/gorm"
)

// ユーザテーブル操作
type UserRepository interface {
	Create(ctx context.Context, name, password string) (*entity.User, error)
	Find(ctx context.Context, name, password string) (*entity.User, error)
}

type userRepositoryGorm struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryGorm{db}
}

func (r *userRepositoryGorm) Create(ctx context.Context, name, password string) (*entity.User, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)

	user := &entity.User{
		Name:     name,
		Password: password,
	}

	if err := tx.Create(user).Error; err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr); pgerr != nil && pgerr.Code == "23505" { // unique_violation
			return nil, errors.Join(wrap(ErrorDuplicated), err)
		}
		return nil, wrap(err)
	}

	return user, nil
}

func (r *userRepositoryGorm) Find(ctx context.Context, name, password string) (*entity.User, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)
	if tx == nil {
		tx = r.db.WithContext(ctx)
	}

	user := &entity.User{}
	if db := tx.Select("id").Where("name = ? AND password = ?", name, password).First(user); db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, wrap(db.Error)
	}

	return user, nil
}
