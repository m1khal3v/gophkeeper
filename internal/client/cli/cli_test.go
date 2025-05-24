package cli

import (
	"bytes"
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/m1khal3v/gophkeeper/internal/common/logger"
)

type mockCommand struct {
	execute func(ctx context.Context, args []string) (string, error)
}

func (m *mockCommand) Execute(ctx context.Context, args []string) (string, error) {
	return m.execute(ctx, args)
}

func TestRun(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		registry          CommandRegistry
		expectedOutput    string
		expectedLogChecks func(t *testing.T, logs []observer.LoggedEntry)
	}{
		{
			name:  "valid command with args",
			input: "cmd1 arg1 arg2\n",
			registry: CommandRegistry{
				"cmd1": &mockCommand{
					execute: func(ctx context.Context, args []string) (string, error) {
						if len(args) != 2 || args[0] != "arg1" || args[1] != "arg2" {
							t.Errorf("unexpected args: %v", args)
						}
						return "success", nil
					},
				},
			},
			expectedOutput: "success\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:  "command with no args",
			input: "cmd2\n",
			registry: CommandRegistry{
				"cmd2": &mockCommand{
					execute: func(ctx context.Context, args []string) (string, error) {
						if len(args) != 0 {
							t.Errorf("unexpected args: %v", args)
						}
						return "executed", nil
					},
				},
			},
			expectedOutput: "executed\n> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
		{
			name:           "unknown command",
			input:          "unknowncmd\n",
			registry:       CommandRegistry{},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
				errorFound := false
				for _, entry := range logs {
					if entry.Message == "Unknown command" && entry.Level == zapcore.ErrorLevel {
						errorFound = true
					}
				}
				if !errorFound {
					t.Errorf("expected error log 'Unknown command', not found")
				}
			},
		},
		{
			name:  "command execution error",
			input: "cmd1 error\n",
			registry: CommandRegistry{
				"cmd1": &mockCommand{
					execute: func(ctx context.Context, args []string) (string, error) {
						return "", errors.New("execution failed")
					},
				},
			},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
				errorFound := false
				for _, entry := range logs {
					if entry.Message == "Command execution error" && entry.Level == zapcore.ErrorLevel {
						errorFound = true
					}
				}
				if !errorFound {
					t.Errorf("expected error log 'Command execution error', not found")
				}
			},
		},
		{
			name:  "empty line",
			input: "\n",
			registry: CommandRegistry{
				"cmd1": &mockCommand{
					execute: func(ctx context.Context, args []string) (string, error) {
						return "ignored", nil
					},
				},
			},
			expectedOutput: "> ",
			expectedLogChecks: func(t *testing.T, logs []observer.LoggedEntry) {
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rIn, wIn, _ := os.Pipe()
			rOut, wOut, _ := os.Pipe()

			stdin := os.Stdin
			stdout := os.Stdout
			defer func() { os.Stdin = stdin }()
			defer func() { os.Stdout = stdout }()

			os.Stdin = rIn
			os.Stdout = wOut

			go func() {
				wIn.Write([]byte(tt.input))
				wIn.Close()
			}()

			core, obs := observer.New(zapcore.DebugLevel)
			logger.Logger = zap.New(core)
			defer func() { _ = logger.Logger.Sync() }()

			ctx := context.Background()
			Run(ctx, tt.registry)

			wOut.Close()

			var writer bytes.Buffer
			_, _ = writer.ReadFrom(rOut)

			actualOutput := writer.String()
			if !strings.HasSuffix(actualOutput, tt.expectedOutput) {
				t.Errorf("unexpected output, got: %q, want suffix: %q", actualOutput, tt.expectedOutput)
			}

			tt.expectedLogChecks(t, obs.All())
		})
	}
}
