package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/client/cli"
	"github.com/m1khal3v/gophkeeper/internal/client/command"
	"github.com/m1khal3v/gophkeeper/internal/client/grpc"
	"github.com/m1khal3v/gophkeeper/internal/common/logger"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"github.com/m1khal3v/gophkeeper/internal/client/config"
	"github.com/m1khal3v/gophkeeper/internal/client/manager"
	"github.com/m1khal3v/gophkeeper/internal/client/repository"
	"github.com/m1khal3v/gophkeeper/internal/client/synchronizer"
)

type App struct {
	syncer   *synchronizer.Synchronizer
	registry cli.CommandRegistry
	db       *sql.DB
}

func New() (*App, error) {
	conf, err := config.ParseArgs()
	if err != nil {
		return nil, fmt.Errorf("can`t parse arguments: %w", err)
	}

	logger.Init("client", zap.InfoLevel.String())

	dbPath := conf.DBPath
	if err := touchFilepath(dbPath); err != nil {
		return nil, fmt.Errorf("can`t touch filepath: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("can`t open db: %w", err)
	}

	userDataRepo, err := repository.NewUserDataRepository(db)
	if err != nil {
		return nil, fmt.Errorf("can`t create data repo: %w", err)
	}

	metaRepo, err := repository.NewMetaRepository(db)
	if err != nil {
		return nil, fmt.Errorf("can`t create meta repo: %w", err)
	}

	userDataManager := manager.NewUserDataManager(userDataRepo)
	metaManager := manager.NewMetaManager(metaRepo)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	ok, err := metaManager.MasterPasswordHashDefined(ctx)
	if err != nil {
		return nil, err
	}
	if !ok {
		if err := metaManager.SetMasterPassword(ctx, conf.MasterPassword); err != nil {
			return nil, err
		}
	} else {
		if err := metaManager.ValidateMasterPassword(ctx, conf.MasterPassword); err != nil {
			return nil, fmt.Errorf("invalid master password: %w", err)
		}
	}

	client, err := grpc.NewClient(conf.ServerAddr)
	if err != nil {
		return nil, fmt.Errorf("can`t create client: %w", err)
	}

	syncer := synchronizer.New(client, userDataManager, metaManager, time.Duration(conf.SyncIntervalSec)*time.Second)

	return &App{
		syncer: syncer,
		registry: cli.CommandRegistry{
			"get":      command.NewGetCommand(userDataManager, []byte(conf.MasterPassword)),
			"set":      command.NewSetCommand(userDataManager, []byte(conf.MasterPassword)),
			"login":    command.NewLoginCommand(client, []byte(conf.MasterPassword)),
			"register": command.NewRegisterCommand(client, []byte(conf.MasterPassword)),
		},
		db: db,
	}, nil
}

func touchFilepath(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}

		return file.Close()
	}

	return nil
}

func (a *App) Run() {
	defer a.db.Close()
	defer a.syncer.Stop()

	var wg sync.WaitGroup
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg.Add(2)
	go func() {
		defer wg.Done()
		a.syncer.Start(ctx)
		stop()
	}()
	go func() {
		defer wg.Done()
		cli.Run(ctx, a.registry)
		stop()
	}()

	wg.Wait()
}
