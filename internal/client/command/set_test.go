package command

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/client/model"
	"github.com/stretchr/testify/assert"
)

type mockDataUpserter struct {
	upsertFunc func(ctx context.Context, data *model.UserData) error
}

func (m *mockDataUpserter) Upsert(ctx context.Context, data *model.UserData) error {
	return m.upsertFunc(ctx, data)
}

func TestSetCommand_Execute_Success(t *testing.T) {
	dataManager := &mockDataUpserter{
		upsertFunc: func(ctx context.Context, data *model.UserData) error {
			assert.Equal(t, "test-key", data.DataKey)
			assert.NotNil(t, data.DataValue)
			assert.False(t, data.UpdatedAt.IsZero())
			assert.Equal(t, time.Unix(0, 0), data.DeletedAt)
			return nil
		},
	}

	cmd := NewSetCommand(dataManager, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"test-key", "text", "test-value"})
	assert.NoError(t, err)
	assert.Equal(t, "saved successful", got)
}

func TestSetCommand_Execute_MissingArgs(t *testing.T) {
	cmd := NewSetCommand(nil, nil)
	got, err := cmd.Execute(context.Background(), []string{})
	assert.Error(t, err)
	assert.Equal(t, "", got)

	got, err = cmd.Execute(context.Background(), []string{"key"})
	assert.Error(t, err)
	assert.Equal(t, "", got)

	got, err = cmd.Execute(context.Background(), []string{"key", "type"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestSetCommand_Execute_InvalidType(t *testing.T) {
	cmd := NewSetCommand(nil, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"key", "invalid-type", "value"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}

func TestSetCommand_Execute_UpsertError(t *testing.T) {
	dataManager := &mockDataUpserter{
		upsertFunc: func(ctx context.Context, data *model.UserData) error {
			return errors.New("upsert error")
		},
	}
	cmd := NewSetCommand(dataManager, []byte("1234567890abcdef"))
	got, err := cmd.Execute(context.Background(), []string{"key", "text", "value"})
	assert.Error(t, err)
	assert.Equal(t, "", got)
}
