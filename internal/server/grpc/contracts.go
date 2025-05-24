package grpc

import (
	"context"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/server/jwt"
	"github.com/m1khal3v/gophkeeper/internal/server/model"
)

type UserManagerInterface interface {
	Register(login, password, masterPassword string) (string, error)
	Login(login, password, masterPassword string) (string, error)
	DecodeToken(token string) (*jwt.Claims, error)
}

type UserDataManagerInterface interface {
	Upsert(ctx context.Context, data *model.UserData) error
	GetUpdates(ctx context.Context, userID uint32, updatedAfter time.Time) ([]*model.UserData, error)
}
