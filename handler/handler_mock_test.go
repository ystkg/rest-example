package handler_test

import (
	"context"
	"log/slog"
	"time"

	"github.com/ystkg/rest-example/entity"
	"github.com/ystkg/rest-example/repository"
	"github.com/ystkg/rest-example/service"
)

type serviceMock struct {
	service.Service                 // 暗黙的な委譲
	base            service.Service // 明示的に委譲

	repository *repositoryMock
}

func newMockService(mock *repositoryMock) *serviceMock {
	s := service.NewService(slog.Default(), mock)
	return &serviceMock{s, s, mock}
}

type repositoryMock struct {
	repository.Repository                       // 暗黙的な委譲
	base                  repository.Repository // 明示的に委譲

	user  *userRepositoryMock
	price *priceRepositoryMock

	beginTxErr error
	commitErr  error
}

func newMockRepository(r repository.Repository) *repositoryMock {
	return &repositoryMock{r, r, newMockUserRepository(r.User()), newMockPriceRepository(r.Price()), nil, nil}
}

func (m *repositoryMock) BeginTx(ctx context.Context) (context.Context, error) {
	if m.beginTxErr != nil {
		return ctx, m.beginTxErr
	}
	return m.base.BeginTx(ctx)
}

func (m *repositoryMock) Commit(ctx context.Context) error {
	if m.commitErr != nil {
		return m.commitErr
	}
	return m.base.Commit(ctx)
}

func (m *repositoryMock) User() repository.UserRepository {
	return m.user
}

func (m *repositoryMock) Price() repository.PriceRepository {
	return m.price
}

type userRepositoryMock struct {
	repository.UserRepository                           // 暗黙的な委譲
	base                      repository.UserRepository // 明示的に委譲

	err error
}

func newMockUserRepository(r repository.UserRepository) *userRepositoryMock {
	return &userRepositoryMock{r, r, nil}
}

func (m *userRepositoryMock) Create(ctx context.Context, name, password string) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.base.Create(ctx, name, password)
}

func (m *userRepositoryMock) Find(ctx context.Context, name, password string) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.base.Find(ctx, name, password)
}

type priceRepositoryMock struct {
	repository.PriceRepository                            // 暗黙的な委譲
	base                       repository.PriceRepository // 明示的に委譲

	rowsAffected int
	overwirte    bool

	err error
}

func newMockPriceRepository(r repository.PriceRepository) *priceRepositoryMock {
	return &priceRepositoryMock{r, r, 0, false, nil}
}

func (m *priceRepositoryMock) Create(
	ctx context.Context,
	userId uint,
	dateTime time.Time,
	store string,
	product string,
	price uint,
	inStock bool,
) (*entity.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.base.Create(
		ctx,
		userId,
		dateTime,
		store,
		product,
		price,
		inStock,
	)
}

func (m *priceRepositoryMock) Find(ctx context.Context, id, userId uint) (*entity.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.base.Find(ctx, id, userId)
}

func (m *priceRepositoryMock) FindByUserId(ctx context.Context, userId uint) ([]entity.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.base.FindByUserId(ctx, userId)
}

func (m *priceRepositoryMock) Update(
	ctx context.Context,
	id uint,
	userId uint,
	dateTime time.Time,
	store string,
	product string,
	price uint,
	inStock bool,
) (*entity.Price, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	if m.overwirte {
		return nil, int64(m.rowsAffected), nil
	}
	return m.base.Update(
		ctx,
		id,
		userId,
		dateTime,
		store,
		product,
		price,
		inStock,
	)
}

func (m *priceRepositoryMock) Delete(ctx context.Context, id, userId uint) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.overwirte {
		return int64(m.rowsAffected), nil
	}
	return m.base.Delete(ctx, id, userId)
}
