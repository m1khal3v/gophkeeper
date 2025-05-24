package command

import (
	"context"
	"errors"
)

type UserRegistrar interface {
	Register(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

type RegisterCommand struct {
	client         UserRegistrar
	masterPassword []byte
}

func NewRegisterCommand(client UserRegistrar, masterPassword []byte) *RegisterCommand {
	return &RegisterCommand{
		client:         client,
		masterPassword: masterPassword,
	}
}

func (c *RegisterCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("args: <login> <password>")
	}

	_, err := c.client.Register(ctx, args[0], args[1], c.masterPassword)
	if err == nil {
		return "register successful", nil
	}

	return "", err
}
