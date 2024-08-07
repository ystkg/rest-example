package repository

import (
	"context"

	"github.com/ystkg/rest-example/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type contextKey string

const (
	contextKeyTx = contextKey("TX")
)

type Repository interface {
	InitDb(ctx context.Context) error

	BeginTx(ctx context.Context) (context.Context, error)
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error

	User() UserRepository
	Price() PriceRepository
}

type repositoryGorm struct {
	db    *gorm.DB
	user  UserRepository
	price PriceRepository
}

func NewRepository(dburl string) (Repository, error) {
	db, err := gorm.Open(postgres.Open(dburl), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &repositoryGorm{
		db:    db,
		user:  NewUserRepository(db),
		price: NewPriceRepository(db),
	}, nil
}

func (r *repositoryGorm) InitDb(ctx context.Context) error {
	return r.db.WithContext(ctx).AutoMigrate(
		&entity.User{},
		&entity.Price{},
	)
}

func (r *repositoryGorm) BeginTx(ctx context.Context) (context.Context, error) {
	tx := r.db.Begin()
	if err := tx.Error; err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, contextKeyTx, tx.WithContext(ctx)), nil
}

func (r *repositoryGorm) Rollback(ctx context.Context) error {
	tx := tx(ctx)
	if tx == nil {
		return nil
	}
	return tx.Rollback().Error
}

func (r *repositoryGorm) Commit(ctx context.Context) error {
	return tx(ctx).Commit().Error
}

func tx(ctx context.Context) *gorm.DB {
	v := ctx.Value(contextKeyTx)
	if v == nil {
		return nil
	}
	tx, ok := v.(*gorm.DB)
	if !ok {
		return nil
	}
	return tx
}

func (r *repositoryGorm) User() UserRepository {
	return r.user
}

func (r *repositoryGorm) Price() PriceRepository {
	return r.price
}
