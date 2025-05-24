package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/common/logger"
	"github.com/m1khal3v/gophkeeper/internal/common/proto"
	"github.com/m1khal3v/gophkeeper/internal/server/config"
	grpcs "github.com/m1khal3v/gophkeeper/internal/server/grpc"
	"github.com/m1khal3v/gophkeeper/internal/server/jwt"
	"github.com/m1khal3v/gophkeeper/internal/server/manager"
	"github.com/m1khal3v/gophkeeper/internal/server/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type App struct {
	cfg      *config.Config
	db       *sql.DB
	services *services
}

func New() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	var level string
	if cfg.Debug {
		level = "debug"
	} else {
		level = "info"
	}
	logger.Init("server", level)

	db, err := initDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init db: %w", err)
	}

	services, err := initServices(db, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init services: %w", err)
	}

	return &App{
		cfg:      cfg,
		db:       db,
		services: services,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcs.NewAuthInterceptor(a.services.userManager).Unary(),
		),
	)
	grpcInternal := grpcs.NewServer(a.services.userManager, a.services.dataManager)
	proto.RegisterAuthServiceServer(grpcServer, grpcInternal)
	proto.RegisterDataServiceServer(grpcServer, grpcInternal)

	listener, err := net.Listen("tcp", a.cfg.Listen)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	go func() {
		logger.Logger.Info("Starting gRPC server", zap.String("addr", a.cfg.Listen))
		if err := grpcServer.Serve(listener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger.Logger.Error("gRPC server failed", zap.Error(err))
		}

		stop()
	}()

	<-ctx.Done()
	stop()

	logger.Logger.Info("Shutting down server...")

	grpcServer.GracefulStop()
	if err := a.db.Close(); err != nil {
		logger.Logger.Error("Failed to close db connection", zap.Error(err))
	}

	logger.Logger.Info("Server stopped gracefully")

	return nil
}

func initDB(cfg *config.Config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return db, nil
}

type services struct {
	userManager *manager.UserManager
	dataManager *manager.UserDataManager
}

func initServices(db *sql.DB, cfg *config.Config) (*services, error) {
	userRepo := repository.NewUserRepository(db)
	dataRepo := repository.NewUserDataRepository(db)

	return &services{
		userManager: manager.NewUserManager(userRepo, jwt.New(cfg.AppSecret)),
		dataManager: manager.NewUserDataManager(dataRepo),
	}, nil
}
