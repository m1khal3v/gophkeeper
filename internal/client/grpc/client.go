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

type Client struct {
	conn       *grpc.ClientConn
	AuthClient proto.AuthServiceClient
	DataClient proto.DataServiceClient
	authToken  string
}

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

func (c *Client) Close() error {
	return c.conn.Close()
}

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

func (c *Client) Upsert(ctx context.Context, data *model.UserData) (*proto.DataResponse, error) {
	ctx = c.withAuth(ctx)
	return c.DataClient.Upsert(ctx, &proto.UpsertRequest{
		DataKey:   data.DataKey,
		DataValue: data.DataValue,
		UpdatedAt: timestamppb.New(data.UpdatedAt),
		DeletedAt: timestamppb.New(data.DeletedAt),
	})
}

func (c *Client) GetUpdates(ctx context.Context, updatedAfter time.Time) (*proto.DataListResponse, error) {
	ctx = c.withAuth(ctx)
	return c.DataClient.GetUpdates(ctx, &proto.GetUpdatesRequest{
		UpdatedAfter: timestamppb.New(updatedAfter),
	})
}

func (c *Client) withAuth(ctx context.Context) context.Context {
	if c.authToken != "" {
		return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+c.authToken)
	}
	return ctx
}
