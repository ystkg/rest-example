package repository

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/ystkg/rest-example/entity"
	"gorm.io/gorm"
)

// 価格テーブル操作
type PriceRepository interface {
	Create(ctx context.Context, userId uint, dateTime time.Time, store, product string, price uint) (*entity.Price, error)
	Find(ctx context.Context, id, userId uint) (*entity.Price, error)
	FindByUserId(ctx context.Context, userId uint) ([]entity.Price, error)
	Update(ctx context.Context, id, userId uint, dateTime time.Time, store, product string, price uint) (*entity.Price, int64, error)
	Delete(ctx context.Context, id, userId uint) (int64, error)
}

type priceRepositoryGorm struct {
	db *gorm.DB
}

func NewPriceRepository(db *gorm.DB) PriceRepository {
	return &priceRepositoryGorm{db}
}

func (r *priceRepositoryGorm) Create(
	ctx context.Context,
	userId uint,
	dateTime time.Time,
	store string,
	product string,
	price uint,
) (*entity.Price, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)

	priceEntity := &entity.Price{
		UserID:   userId,
		DateTime: dateTime,
		Store:    store,
		Product:  product,
		Price:    price,
	}

	if err := tx.Create(priceEntity).Error; err != nil {
		return nil, wrap(err)
	}

	return priceEntity, nil
}

func (r *priceRepositoryGorm) Find(ctx context.Context, id, userId uint) (*entity.Price, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)
	if tx == nil {
		tx = r.db.WithContext(ctx)
	}

	price := &entity.Price{
		Model: gorm.Model{
			ID: id,
		},
	}

	if err := tx.Where("user_id = ?", userId).First(price).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, wrap(err)
	}

	return price, nil
}

func (r *priceRepositoryGorm) FindByUserId(ctx context.Context, userId uint) ([]entity.Price, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)
	if tx == nil {
		tx = r.db.WithContext(ctx)
	}

	var entities []entity.Price
	if err := tx.Where("user_id = ?", userId).Find(&entities).Error; err != nil {
		return nil, wrap(err)
	}

	return entities, nil
}

func (r *priceRepositoryGorm) Update(
	ctx context.Context,
	id uint,
	userId uint,
	dateTime time.Time,
	store string,
	product string,
	price uint,
) (*entity.Price, int64, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)

	priceEntity := &entity.Price{
		Model: gorm.Model{
			ID: id,
		},
		UserID:   userId,
		DateTime: dateTime,
		Store:    store,
		Product:  product,
		Price:    price,
	}

	db := tx.Where("user_id = ? and deleted_at is null", userId).Updates(priceEntity)
	if db.Error != nil {
		return nil, 0, wrap(db.Error)
	}

	return priceEntity, db.RowsAffected, nil
}

func (r *priceRepositoryGorm) Delete(ctx context.Context, id, userId uint) (int64, error) {
	slog.DebugContext(ctx, "start")
	defer slog.DebugContext(ctx, "end")

	tx := tx(ctx)

	price := &entity.Price{
		Model: gorm.Model{
			ID: id,
		},
	}

	db := tx.Where("user_id = ?", userId).Delete(price)
	if db.Error != nil {
		return 0, wrap(db.Error)
	}

	return db.RowsAffected, nil
}
