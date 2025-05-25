package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/m1khal3v/gophkeeper/internal/client/command"
	"github.com/m1khal3v/gophkeeper/internal/common/logger"
	"go.uber.org/zap"
)

type CommandRegistry map[string]command.Command

func Run(ctx context.Context, registry CommandRegistry) {
	reader := bufio.NewScanner(os.Stdin)

	logger.Logger.Info("GophKeeper started")
	fmt.Print("> ")

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		if len(line) == 0 {
			fmt.Print("> ")

			continue
		}

		parts := strings.Fields(line)
		cmdName := parts[0]
		args := parts[1:]

		cmd, ok := registry[cmdName]
		if !ok {
			logger.Logger.Error("Unknown command", zap.String("command", cmdName))
			fmt.Print("> ")

			continue
		}

		result, err := cmd.Execute(ctx, args)
		if err != nil {
			logger.Logger.Error("Command execution error", zap.Error(err))
			fmt.Print("> ")

			continue
		}

		fmt.Println(result)
		fmt.Print("> ")
	}

	err := reader.Err()
	if err != nil {
		logger.Logger.Fatal("STDIN read error", zap.Error(err))
	}
}
