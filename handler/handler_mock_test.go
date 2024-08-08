package handler_test

import (
	"context"
	"time"

	"github.com/ystkg/rest-example/entity"
	"github.com/ystkg/rest-example/repository"
	"github.com/ystkg/rest-example/service"
)

type serviceMock struct {
	service.Service
	base service.Service

	repository repository.Repository
}

func newMockService(r repository.Repository) *serviceMock {
	s := service.NewService(r)
	return &serviceMock{s, s, r}
}

type repositoryMock struct {
	repository.Repository
	base repository.Repository

	user  repository.UserRepository
	price repository.PriceRepository

	beginTxErr error
	commitErr  error
}

func newMockRepository(s *serviceMock) *repositoryMock {
	r := s.repository
	return &repositoryMock{r, r, r.User(), r.Price(), nil, nil}
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
	repository.UserRepository
	base repository.UserRepository

	err error
}

func newMockUserRepository(s *serviceMock) *userRepositoryMock {
	r := s.repository.User()
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
	repository.PriceRepository
	base repository.PriceRepository

	rowsAffected int
	overwirte    bool

	err error
}

func newMockPriceRepository(s *serviceMock) *priceRepositoryMock {
	r := s.repository.Price()
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
