package command

import (
	"context"
	"errors"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/client/aes"
	"github.com/m1khal3v/gophkeeper/internal/client/manager"
	"github.com/m1khal3v/gophkeeper/internal/client/model"
	"github.com/m1khal3v/gophkeeper/internal/client/value"
)

type SetCommand struct {
	dataManager    *manager.UserDataManager
	masterPassword []byte
}

func NewSetCommand(dataManager *manager.UserDataManager, masterPassword []byte) *SetCommand {
	return &SetCommand{
		dataManager:    dataManager,
		masterPassword: masterPassword,
	}
}

func (c *SetCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) < 3 {
		return "", errors.New("args: <key> <type> <args...>")
	}

	val, err := value.FromUserInput(args[1], args[2:])
	if err != nil {
		return "", err
	}

	raw, err := val.ToBytes()
	if err != nil {
		return "", err
	}

	encRaw, err := aes.Encrypt(c.masterPassword, raw)
	if err != nil {
		return "", err
	}

	err = c.dataManager.Upsert(ctx, &model.UserData{
		DataKey:   args[0],
		DataValue: encRaw,
		UpdatedAt: time.Now(),
		DeletedAt: time.Unix(0, 0),
	})

	if err != nil {
		return "", err
	}

	return "saved successful", nil
}
