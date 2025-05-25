package command

import (
	"context"
	"errors"
)

type UserAuthenticator interface {
	Login(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

type LoginCommand struct {
	client         UserAuthenticator
	masterPassword []byte
}

func NewLoginCommand(client UserAuthenticator, masterPassword []byte) *LoginCommand {
	return &LoginCommand{
		client:         client,
		masterPassword: masterPassword,
	}
}

func (c *LoginCommand) Execute(ctx context.Context, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("args: <login> <password>")
	}

	_, err := c.client.Login(ctx, args[0], args[1], c.masterPassword)
	if err == nil {
		return "login successful", nil
	}

	return "", err
}
