package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	pkgerrors "github.com/pkg/errors"
	"github.com/ystkg/rest-example/entity"
	"github.com/ystkg/rest-example/repository"
)

type Service interface {
	CreateUser(ctx context.Context, name, password string) (*uint, error)
	FindUser(ctx context.Context, name, password string) (*uint, error)

	CreatePrice(ctx context.Context, userId uint, dateTime time.Time, store, product string, price uint, inStock bool) (*entity.Price, error)
	FindPrices(ctx context.Context, userId uint) ([]entity.Price, error)
	FindPrice(ctx context.Context, priceId, userId uint) (*entity.Price, error)
	UpdatePrice(ctx context.Context, priceId, userId uint, dateTime time.Time, store, product string, price uint, inStock bool) (*entity.Price, error)
	DeletePrice(ctx context.Context, priceId, userId uint) error
}

type serviceImpl struct {
	logger *slog.Logger

	repository repository.Repository
}

func NewService(logger *slog.Logger, r repository.Repository) Service {
	return &serviceImpl{logger, r}
}

func (s *serviceImpl) beginTx(ctx context.Context) (context.Context, error) {
	return s.repository.BeginTx(ctx)
}

func (s *serviceImpl) rollback(ctx context.Context) error {
	return s.repository.Rollback(ctx)
}

func (s *serviceImpl) commit(ctx context.Context) error {
	return s.repository.Commit(ctx)
}

// ユーザの登録
func (s *serviceImpl) CreateUser(ctx context.Context, name, password string) (*uint, error) {
	s.logger.DebugContext(ctx, "serviceImpl#CreateUser start")
	defer s.logger.DebugContext(ctx, "serviceImpl#CreateUser end")

	// トランザクション開始
	ctx, err := s.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rollback(ctx)

	// ユーザの登録
	encPassword := encodePassword(name, password)
	user, err := s.repository.User().Create(ctx, name, encPassword)
	if err != nil {
		return nil, err
	}

	// コミット
	if err = s.commit(ctx); err != nil {
		return nil, err
	}

	return &user.ID, nil
}

// ユーザIDの取得
func (s *serviceImpl) FindUser(ctx context.Context, name, password string) (*uint, error) {
	s.logger.DebugContext(ctx, "serviceImpl#FindUser start")
	defer s.logger.DebugContext(ctx, "serviceImpl#FindUser end")

	encPassword := encodePassword(name, password)
	user, err := s.repository.User().Find(ctx, name, encPassword)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	return &user.ID, nil
}

func encodePassword(user, password string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s %s", user, password)))
	return hex.EncodeToString(hash[:])
}

// 価格の登録
func (s *serviceImpl) CreatePrice(ctx context.Context, userId uint, dateTime time.Time, store, product string, price uint, inStock bool) (*entity.Price, error) {
	s.logger.DebugContext(ctx, "serviceImpl#CreatePrice start")
	defer s.logger.DebugContext(ctx, "serviceImpl#CreatePrice end")

	// トランザクション開始
	ctx, err := s.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rollback(ctx)

	// 価格の登録
	priceEntity, err := s.repository.Price().Create(ctx, userId, dateTime, store, product, price, inStock)
	if err != nil {
		return nil, err
	}

	// コミット
	if err = s.commit(ctx); err != nil {
		return nil, err
	}

	return priceEntity, nil
}

// 価格の一覧
func (s *serviceImpl) FindPrices(ctx context.Context, userId uint) ([]entity.Price, error) {
	s.logger.DebugContext(ctx, "serviceImpl#FindPrices start")
	defer s.logger.DebugContext(ctx, "serviceImpl#FindPrices end")

	return s.repository.Price().FindByUserId(ctx, userId)
}

// 価格の取得
func (s *serviceImpl) FindPrice(ctx context.Context, priceId, userId uint) (*entity.Price, error) {
	s.logger.DebugContext(ctx, "serviceImpl#FindPrice start")
	defer s.logger.DebugContext(ctx, "serviceImpl#FindPrice end")

	return s.repository.Price().Find(ctx, priceId, userId)
}

// 価格の更新
func (s *serviceImpl) UpdatePrice(ctx context.Context, priceId, userId uint, dateTime time.Time, store, product string, price uint, inStock bool) (*entity.Price, error) {
	s.logger.DebugContext(ctx, "serviceImpl#UpdatePrice start")
	defer s.logger.DebugContext(ctx, "serviceImpl#UpdatePrice end")

	// トランザクション開始
	ctx, err := s.beginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.rollback(ctx)

	// 価格の更新
	priceEntity, rows, err := s.repository.Price().Update(
		ctx,
		priceId,
		userId,
		dateTime,
		store,
		product,
		price,
		inStock,
	)
	if err != nil {
		return nil, err
	}
	if rows != 1 {
		s.rollback(ctx)
		if rows == 0 {
			return nil, pkgerrors.WithStack(ErrorNotFound)
		}
		return nil, pkgerrors.Errorf("RowsAffected:%d", rows)
	}

	// コミット
	if err = s.commit(ctx); err != nil {
		return nil, err
	}

	return priceEntity, nil
}

// 価格の削除
func (s *serviceImpl) DeletePrice(ctx context.Context, priceId, userId uint) error {
	s.logger.DebugContext(ctx, "serviceImpl#DeletePrice start")
	defer s.logger.DebugContext(ctx, "serviceImpl#DeletePrice end")

	// トランザクション開始
	ctx, err := s.beginTx(ctx)
	if err != nil {
		return err
	}
	defer s.rollback(ctx)

	// 価格の削除
	rows, err := s.repository.Price().Delete(ctx, priceId, userId)
	if err != nil {
		return err
	}
	if rows != 1 {
		s.rollback(ctx)
		if rows == 0 {
			return pkgerrors.WithStack(ErrorNotFound)
		}
		return pkgerrors.Errorf("RowsAffected:%d", rows)
	}

	// コミット
	if err = s.commit(ctx); err != nil {
		return err
	}

	return nil
}
