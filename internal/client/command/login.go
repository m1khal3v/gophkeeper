package command

import (
	"context"
	"errors"

	"github.com/m1khal3v/gophkeeper/internal/client/grpc"
)

type LoginCommand struct {
	client         *grpc.Client
	masterPassword []byte
}

func NewLoginCommand(client *grpc.Client, masterPassword []byte) *LoginCommand {
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
