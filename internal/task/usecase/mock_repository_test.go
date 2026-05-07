package usecase_test

import (
	"context"

	"github.com/stretchr/testify/mock"
	"github.com/yourorg/cell-arch/internal/task/entity"
)

// mockRepository is a testify mock implementing entity.Repository.
// Generated manually for Phase 1; Phase 5 will use mockery/mockgen.
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) Get(ctx context.Context, id string) (*entity.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *mockRepository) Create(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *mockRepository) Update(ctx context.Context, t *entity.Task) (*entity.Task, error) {
	args := m.Called(ctx, t)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Task), args.Error(1)
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRepository) List(ctx context.Context) ([]*entity.Task, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Task), args.Error(1)
}
