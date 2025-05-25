package command

import (
	"context"
	"errors"
	"testing"

	"github.com/m1khal3v/gophkeeper/internal/client/aes"
	"github.com/m1khal3v/gophkeeper/internal/client/model"
	"github.com/m1khal3v/gophkeeper/internal/client/value"
	"github.com/stretchr/testify/assert"
)

type mockDataManager struct {
	getFunc func(ctx context.Context, key string) (*model.UserData, error)
}

func (m *mockDataManager) Get(ctx context.Context, key string) (*model.UserData, error) {
	return m.getFunc(ctx, key)
}

func TestGetCommand_Execute_Success(t *testing.T) {
	password := []byte("1234567890abcdef")
	want := "some-value"
	val, err := value.FromUserInput("text", []string{want})
	assert.NoError(t, err)
	bytes, err := val.ToBytes()
	assert.NoError(t, err)

	cipherBytes, err := aes.Encrypt(password, bytes)
	assert.NoError(t, err)

	dataManager := &mockDataManager{
		getFunc: func(ctx context.Context, key string) (*model.UserData, error) {
			return &model.UserData{
				DataKey:   key,
				DataValue: cipherBytes,
			}, nil
		},
	}

	cmd := NewGetCommand(dataManager, password)
	got, err := cmd.Execute(context.Background(), []string{"some-key"})
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestGetCommand_Execute_MissingArgs(t *testing.T) {
	cmd := NewGetCommand(nil, nil)
	got, err := cmd.Execute(context.Background(), []string{})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestGetCommand_Execute_GetError(t *testing.T) {
	dataManager := &mockDataManager{
		getFunc: func(ctx context.Context, key string) (*model.UserData, error) {
			return nil, errors.New("not found")
		},
	}
	cmd := NewGetCommand(dataManager, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"some-key"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestGetCommand_Execute_DecryptError(t *testing.T) {
	dataManager := &mockDataManager{
		getFunc: func(ctx context.Context, key string) (*model.UserData, error) {
			return &model.UserData{
				DataKey:   key,
				DataValue: []byte("corrupted-cipher"),
			}, nil
		},
	}
	cmd := NewGetCommand(dataManager, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"test"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}
