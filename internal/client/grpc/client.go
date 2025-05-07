package grpc

import (
	"context"
	"time"

	"github.com/m1khal3v/gophkeeper/internal/client/model"
	"github.com/m1khal3v/gophkeeper/internal/common/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client инкапсулирует gRPC-клиентов обоих сервисов и токен авторизации.
type Client struct {
	conn       *grpc.ClientConn
	AuthClient proto.AuthServiceClient
	DataClient proto.DataServiceClient
	authToken  string // Bearer-токен пользователя
}

// NewClient устанавливает соединение с gRPC сервером.
func NewClient(serverAddr string) (*Client, error) {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:       conn,
		AuthClient: proto.NewAuthServiceClient(conn),
		DataClient: proto.NewDataServiceClient(conn),
	}, nil
}

// Close закрывает соединение.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Login выполняет логин и сохраняет токен.
func (c *Client) Login(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
	resp, err := c.AuthClient.Login(ctx, &proto.LoginRequest{
		Login:          login,
		Password:       password,
		MasterPassword: string(masterPassword),
	})
	if err != nil {
		return "", err
	}
	c.authToken = resp.Token
	return resp.Token, nil
}

// Register выполняет регистрацию и возвращает токен.
func (c *Client) Register(ctx context.Context, login, password string, masterPassword []byte) (string, error) {
	resp, err := c.AuthClient.Register(ctx, &proto.RegisterRequest{
		Login:          login,
		Password:       password,
		MasterPassword: string(masterPassword),
	})
	if err != nil {
		return "", err
	}
	c.authToken = resp.Token
	return resp.Token, nil
}

// Upsert отправляет данные пользователя с авторизацией по токену.
func (c *Client) Upsert(ctx context.Context, data *model.UserData) (*proto.DataResponse, error) {
	ctx = c.withAuth(ctx)
	return c.DataClient.Upsert(ctx, &proto.UpsertRequest{
		DataKey:   data.DataKey,
		DataValue: data.DataValue,
		UpdatedAt: timestamppb.New(data.UpdatedAt),
		DeletedAt: timestamppb.New(data.DeletedAt),
	})
}

// GetUpdates запрашивает обновления данных пользователя с авторизацией по токену.
func (c *Client) GetUpdates(ctx context.Context, updatedAfter time.Time) (*proto.DataListResponse, error) {
	ctx = c.withAuth(ctx)
	return c.DataClient.GetUpdates(ctx, &proto.GetUpdatesRequest{
		UpdatedAfter: timestamppb.New(updatedAfter),
	})
}

// withAuth добавляет токен авторизации к контексту, если он есть.
func (c *Client) withAuth(ctx context.Context) context.Context {
	if c.authToken != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.authToken)
	}
	return ctx
}
