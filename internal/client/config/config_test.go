package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseArgs_Success(t *testing.T) {
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"app", "test.db", "secret123"}
	cfg, err := ParseArgs()

	assert.NoError(t, err)
	assert.Equal(t, "test.db", cfg.DBPath)
	assert.Equal(t, "secret123", cfg.MasterPassword)
	assert.Equal(t, "localhost:50501", cfg.ServerAddr)
	assert.Equal(t, 60, cfg.SyncIntervalSec)
}

func TestParseArgs_WithFlags(t *testing.T) {
	oldArgs := os.Args
	oldFlagCommandLine := flag.CommandLine
	defer func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagCommandLine
	}()

	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	os.Args = []string{"app", "-addr", "127.0.0.1:8080", "-interval", "30", "custom.db", "pass456"}
	cfg, err := ParseArgs()

	assert.NoError(t, err)
	assert.Equal(t, "custom.db", cfg.DBPath)
	assert.Equal(t, "pass456", cfg.MasterPassword)
	assert.Equal(t, "127.0.0.1:8080", cfg.ServerAddr)
	assert.Equal(t, 30, cfg.SyncIntervalSec)
}

func TestParseArgs_InvalidArguments(t *testing.T) {
	testCases := []struct {
		name string
		args []string
	}{
		{
			name: "no arguments",
			args: []string{"app"},
		},
		{
			name: "one argument",
			args: []string{"app", "test.db"},
		},
		{
			name: "too many arguments",
			args: []string{"app", "test.db", "secret123", "extra_arg"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			oldArgs := os.Args
			oldFlagCommandLine := flag.CommandLine
			defer func() {
				os.Args = oldArgs
				flag.CommandLine = oldFlagCommandLine
			}()

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			os.Args = tc.args
			cfg, err := ParseArgs()

			assert.Error(t, err)
			assert.Nil(t, cfg)
		})
	}
}
