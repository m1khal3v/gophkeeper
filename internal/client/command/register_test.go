package command

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockRegistrarClient struct {
	registerFunc func(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

func (m *mockRegistrarClient) Register(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
	return m.registerFunc(ctx, login, password, masterPassword)
}

func TestRegisterCommand_Execute_Success(t *testing.T) {
	client := &mockRegistrarClient{
		registerFunc: func(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
			return "token", nil
		},
	}

	cmd := NewRegisterCommand(client, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"user", "pass"})
	assert.NoError(t, err)
	assert.Equal(t, "register successful", got)
}

func TestRegisterCommand_Execute_MissingArgs(t *testing.T) {
	cmd := NewRegisterCommand(nil, nil)
	got, err := cmd.Execute(context.Background(), []string{})
	assert.Error(t, err)
	assert.Equal(t, "", got)

	got, err = cmd.Execute(context.Background(), []string{"user"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestRegisterCommand_Execute_RegisterError(t *testing.T) {
	client := &mockRegistrarClient{
		registerFunc: func(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
			return "", errors.New("user already exists")
		},
	}
	cmd := NewRegisterCommand(client, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"user", "pass"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}
