package command

import (
	"context"
	"errors"

	"github.com/m1khal3v/gophkeeper/internal/client/aes"
	"github.com/m1khal3v/gophkeeper/internal/client/manager"
	"github.com/m1khal3v/gophkeeper/internal/client/value"
)

type GetCommand struct {
	dataManager    *manager.UserDataManager
	masterPassword []byte
}

func NewGetCommand(dataManager *manager.UserDataManager, masterPassword []byte) *GetCommand {
	return &GetCommand{
		dataManager:    dataManager,
		masterPassword: masterPassword,
	}
}

func (c *GetCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("args: <masterPassword>")
	}

	data, err := c.dataManager.Get(ctx, args[0])
	if err != nil {
		return "", err
	}

	raw, err := aes.Decrypt(c.masterPassword, data.DataValue)
	if err != nil {
		return "", err
	}

	val, err := value.FromBytes(raw)
	if err != nil {
		return "", err
	}

	return val.String(), nil
}
