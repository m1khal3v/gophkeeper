package manager

import (
	"context"
	"errors"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/client/model"
	"github.com/m1khal3v/gophkeeper/internal/client/repository"
)

var (
	ErrNotFound = errors.New("data not found")
)

type UserDataRepository interface {
	Upsert(ctx context.Context, data *model.UserData) error
	Get(ctx context.Context, key string) (*model.UserData, error)
	GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error)
}

type UserDataManager struct {
	dataRepo UserDataRepository
}

func NewUserDataManager(repo *repository.UserDataRepository) *UserDataManager {
	return &UserDataManager{
		dataRepo: repo,
	}
}

func (m *UserDataManager) Upsert(ctx context.Context, data *model.UserData) error {
	return m.dataRepo.Upsert(ctx, data)
}

func (m *UserDataManager) Get(ctx context.Context, key string) (*model.UserData, error) {
	data, err := m.dataRepo.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *UserDataManager) GetUpdates(ctx context.Context, lastSync time.Time) ([]*model.UserData, error) {
	return m.dataRepo.GetUpdates(ctx, lastSync)
}
