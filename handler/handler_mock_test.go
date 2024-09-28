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

	repository *repositoryMock
}

func newMockService(mock *repositoryMock) *serviceMock {
	s := service.NewService(mock)
	return &serviceMock{s, mock}
}

type repositoryMock struct {
	repository.Repository

	user  *userRepositoryMock
	price *priceRepositoryMock

	beginTxErr error
	commitErr  error
}

func newMockRepository(r repository.Repository) *repositoryMock {
	return &repositoryMock{r, newMockUserRepository(r.User()), newMockPriceRepository(r.Price()), nil, nil}
}

func (m *repositoryMock) BeginTx(ctx context.Context) (context.Context, error) {
	if m.beginTxErr != nil {
		return ctx, m.beginTxErr
	}
	return m.Repository.BeginTx(ctx)
}

func (m *repositoryMock) Commit(ctx context.Context) error {
	if m.commitErr != nil {
		return m.commitErr
	}
	return m.Repository.Commit(ctx)
}

func (m *repositoryMock) User() repository.UserRepository {
	return m.user
}

func (m *repositoryMock) Price() repository.PriceRepository {
	return m.price
}

type userRepositoryMock struct {
	repository.UserRepository

	err error
}

func newMockUserRepository(r repository.UserRepository) *userRepositoryMock {
	return &userRepositoryMock{r, nil}
}

func (m *userRepositoryMock) Create(ctx context.Context, name, password string) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.UserRepository.Create(ctx, name, password)
}

func (m *userRepositoryMock) Find(ctx context.Context, name, password string) (*entity.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.UserRepository.Find(ctx, name, password)
}

type priceRepositoryMock struct {
	repository.PriceRepository

	rowsAffected int
	overwirte    bool

	err error
}

func newMockPriceRepository(r repository.PriceRepository) *priceRepositoryMock {
	return &priceRepositoryMock{r, 0, false, nil}
}

func (m *priceRepositoryMock) Create(
	ctx context.Context,
	userId uint,
	dateTime time.Time,
	store string,
	product string,
	price uint,
) (*entity.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.PriceRepository.Create(
		ctx,
		userId,
		dateTime,
		store,
		product,
		price,
	)
}

func (m *priceRepositoryMock) Find(ctx context.Context, id, userId uint) (*entity.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.PriceRepository.Find(ctx, id, userId)
}

func (m *priceRepositoryMock) FindByUserId(ctx context.Context, userId uint) ([]entity.Price, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.PriceRepository.FindByUserId(ctx, userId)
}

func (m *priceRepositoryMock) Update(
	ctx context.Context,
	id uint,
	userId uint,
	dateTime time.Time,
	store string,
	product string,
	price uint,
) (*entity.Price, int64, error) {
	if m.err != nil {
		return nil, 0, m.err
	}
	if m.overwirte {
		return nil, int64(m.rowsAffected), nil
	}
	return m.PriceRepository.Update(
		ctx,
		id,
		userId,
		dateTime,
		store,
		product,
		price,
	)
}

func (m *priceRepositoryMock) Delete(ctx context.Context, id, userId uint) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}
	if m.overwirte {
		return int64(m.rowsAffected), nil
	}
	return m.PriceRepository.Delete(ctx, id, userId)
}
