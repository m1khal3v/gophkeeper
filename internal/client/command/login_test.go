package command

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockAuthClient struct {
	loginFunc func(ctx context.Context, login, password string, masterPassword []byte) (string, error)
}

func (m *mockAuthClient) Login(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
	return m.loginFunc(ctx, login, password, masterPassword)
}

func TestLoginCommand_Execute_Success(t *testing.T) {
	client := &mockAuthClient{
		loginFunc: func(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
			return "token", nil
		},
	}

	cmd := NewLoginCommand(client, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"user", "pass"})
	assert.NoError(t, err)
	assert.Equal(t, "login successful", got)
}

func TestLoginCommand_Execute_MissingArgs(t *testing.T) {
	cmd := NewLoginCommand(nil, nil)
	got, err := cmd.Execute(context.Background(), []string{})
	assert.Error(t, err)
	assert.Equal(t, "", got)

	got, err = cmd.Execute(context.Background(), []string{"user"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestLoginCommand_Execute_LoginError(t *testing.T) {
	client := &mockAuthClient{
		loginFunc: func(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
			return "", errors.New("invalid credentials")
		},
	}
	cmd := NewLoginCommand(client, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"user", "pass"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}
