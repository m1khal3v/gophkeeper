package command

import (
	"context"
	"errors"

	"github.com/m1khal3v/gophkeeper/internal/client/grpc"
)

type RegisterCommand struct {
	client         *grpc.Client
	masterPassword []byte
}

func NewRegisterCommand(client *grpc.Client, masterPassword []byte) *RegisterCommand {
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
