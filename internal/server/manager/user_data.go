package manager

import (
	"context"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/server/model"
	"github.com/m1khal3v/gophkeeper/internal/server/repository"
)

type UserDataRepository interface {
	Upsert(ctx context.Context, data *model.UserData) error
	GetUpdates(ctx context.Context, userID uint32, since time.Time) ([]*model.UserData, error)
}

type UserDataManager struct {
	dataRepo UserDataRepository
}

func NewUserDataManager(dataRepo *repository.UserDataRepository) *UserDataManager {
	return &UserDataManager{
		dataRepo: dataRepo,
	}
}

func (m *UserDataManager) Upsert(ctx context.Context, data *model.UserData) error {
	return m.dataRepo.Upsert(ctx, data)
}

func (m *UserDataManager) GetUpdates(ctx context.Context, userID uint32, since time.Time) ([]*model.UserData, error) {
	return m.dataRepo.GetUpdates(ctx, userID, since)
}
