package repository

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	pkgerrors "github.com/pkg/errors"
	"github.com/ystkg/rest-example/entity"
	"gorm.io/gorm"
)

// ユーザテーブル操作
type UserRepository interface {
	Create(ctx context.Context, name, password string) (*entity.User, error)
	Find(ctx context.Context, name, password string) (*entity.User, error)
}

type userRepositoryGorm struct {
	logger *slog.Logger

	db *gorm.DB
}

func NewUserRepository(logger *slog.Logger, db *gorm.DB) UserRepository {
	return &userRepositoryGorm{logger, db}
}

func (r *userRepositoryGorm) Create(ctx context.Context, name, password string) (*entity.User, error) {
	r.logger.DebugContext(ctx, "userRepositoryGorm#Create start")
	defer r.logger.DebugContext(ctx, "userRepositoryGorm#Create end")

	tx := tx(ctx)

	user := &entity.User{
		Name:     name,
		Password: password,
	}

	if err := tx.Create(user).Error; err != nil {
		var pgerr *pgconn.PgError
		if errors.As(err, &pgerr); pgerr.Code == "23505" { // unique_violation
			return nil, errors.Join(pkgerrors.WithStack(ErrorDuplicated), err)
		}
		return nil, pkgerrors.WithStack(err)
	}

	return user, nil
}

func (r *userRepositoryGorm) Find(ctx context.Context, name, password string) (*entity.User, error) {
	r.logger.DebugContext(ctx, "userRepositoryGorm#Find start")
	defer r.logger.DebugContext(ctx, "userRepositoryGorm#Find end")

	tx := tx(ctx)
	if tx == nil {
		tx = r.db.WithContext(ctx)
	}

	user := &entity.User{}
	if db := tx.Select("id").Where("name = ? AND password = ?", name, password).First(user); db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, pkgerrors.WithStack(db.Error)
	}

	return user, nil
}
