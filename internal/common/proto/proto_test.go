package proto

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockClientConn struct {
	mock.Mock
}

func (m *mockClientConn) Invoke(ctx context.Context, method string, args, reply interface{}, opts ...grpc.CallOption) error {
	mockArgs := m.Called(ctx, method, args, reply)
	return mockArgs.Error(0)
}

func (m *mockClientConn) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	mockArgs := m.Called(ctx, desc, method)
	return mockArgs.Get(0).(grpc.ClientStream), mockArgs.Error(1)
}

// MockClientStream is a mock implementation of grpc.ClientStream
type MockClientStream struct {
	mock.Mock
}

func (m *MockClientStream) Header() (metadata.MD, error) {
	args := m.Called()
	return args.Get(0).(metadata.MD), args.Error(1)
}

func (m *MockClientStream) Trailer() metadata.MD {
	args := m.Called()
	return args.Get(0).(metadata.MD)
}

func (m *MockClientStream) CloseSend() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockClientStream) Context() context.Context {
	args := m.Called()
	return args.Get(0).(context.Context)
}

func (m *MockClientStream) SendMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

func (m *MockClientStream) RecvMsg(msg interface{}) error {
	args := m.Called(msg)
	return args.Error(0)
}

type mockServiceRegistrar struct {
	mock.Mock
}

func (m *mockServiceRegistrar) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	m.Called(desc, impl)
}

func TestRegisterAuthServiceServer(t *testing.T) {
	mockRegistrar := new(mockServiceRegistrar)
	mockServer := new(mockAuthServer)

	mockRegistrar.On("RegisterService", &AuthService_ServiceDesc, mockServer).Return()

	RegisterAuthServiceServer(mockRegistrar, mockServer)
	mockRegistrar.AssertExpectations(t)
}

func TestRegisterDataServiceServer(t *testing.T) {
	mockRegistrar := new(mockServiceRegistrar)
	mockServer := new(mockDataServer)

	mockRegistrar.On("RegisterService", &DataService_ServiceDesc, mockServer).Return()

	RegisterDataServiceServer(mockRegistrar, mockServer)
	mockRegistrar.AssertExpectations(t)
}

// Test panic condition for nil server embedded by pointer
type badAuthServer struct {
	*UnimplementedAuthServiceServer
}

func (s *badAuthServer) mustEmbedUnimplementedAuthServiceServer() {}

type badDataServer struct {
	*UnimplementedDataServiceServer
}

func (s *badDataServer) mustEmbedUnimplementedDataServiceServer() {}

func TestRegisterNilAuthServerPanic(t *testing.T) {
	mockRegistrar := new(mockServiceRegistrar)
	mockServer := &badAuthServer{nil}

	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()

	RegisterAuthServiceServer(mockRegistrar, mockServer)
}

func TestRegisterNilDataServerPanic(t *testing.T) {
	mockRegistrar := new(mockServiceRegistrar)
	mockServer := &badDataServer{nil}

	defer func() {
		r := recover()
		assert.NotNil(t, r)
	}()

	RegisterDataServiceServer(mockRegistrar, mockServer)
}

func TestAuthServiceClient_Register(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewAuthServiceClient(mockConn)

	req := &RegisterRequest{
		Login:          "testuser",
		Password:       "password123",
		MasterPassword: "master123",
	}
	expectedResp := &TokenResponse{Token: "test-token"}

	mockConn.On("Invoke", mock.Anything, AuthService_Register_FullMethodName, req, mock.Anything).
		Run(func(args mock.Arguments) {
			resp := args.Get(3).(*TokenResponse)
			resp.Token = expectedResp.Token
		}).
		Return(nil)

	resp, err := client.Register(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp.Token, resp.Token)
	mockConn.AssertExpectations(t)
}

func TestAuthServiceClient_Register_Error(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewAuthServiceClient(mockConn)

	req := &RegisterRequest{
		Login:          "testuser",
		Password:       "password123",
		MasterPassword: "master123",
	}
	expectedErr := errors.New("connection error")

	mockConn.On("Invoke", mock.Anything, AuthService_Register_FullMethodName, req, mock.Anything).
		Return(expectedErr)

	resp, err := client.Register(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	mockConn.AssertExpectations(t)
}

func TestAuthServiceClient_Login(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewAuthServiceClient(mockConn)

	req := &LoginRequest{
		Login:          "testuser",
		Password:       "password123",
		MasterPassword: "master123",
	}
	expectedResp := &TokenResponse{Token: "test-token"}

	mockConn.On("Invoke", mock.Anything, AuthService_Login_FullMethodName, req, mock.Anything).
		Run(func(args mock.Arguments) {
			resp := args.Get(3).(*TokenResponse)
			resp.Token = expectedResp.Token
		}).
		Return(nil)

	resp, err := client.Login(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp.Token, resp.Token)
	mockConn.AssertExpectations(t)
}

func TestAuthServiceClient_Login_Error(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewAuthServiceClient(mockConn)

	req := &LoginRequest{
		Login:          "testuser",
		Password:       "password123",
		MasterPassword: "master123",
	}
	expectedErr := errors.New("connection error")

	mockConn.On("Invoke", mock.Anything, AuthService_Login_FullMethodName, req, mock.Anything).
		Return(expectedErr)

	resp, err := client.Login(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	mockConn.AssertExpectations(t)
}

func TestDataServiceClient_Upsert(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewDataServiceClient(mockConn)

	now := time.Now()
	updatedAt := timestamppb.New(now)
	deletedAt := timestamppb.New(now.Add(time.Hour))

	req := &UpsertRequest{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}
	expectedResp := &DataResponse{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
	}

	mockConn.On("Invoke", mock.Anything, DataService_Upsert_FullMethodName, req, mock.Anything).
		Run(func(args mock.Arguments) {
			resp := args.Get(3).(*DataResponse)
			resp.DataKey = expectedResp.DataKey
			resp.DataValue = expectedResp.DataValue
			resp.UpdatedAt = expectedResp.UpdatedAt
			resp.DeletedAt = expectedResp.DeletedAt
		}).
		Return(nil)

	resp, err := client.Upsert(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, expectedResp.DataKey, resp.DataKey)
	assert.Equal(t, expectedResp.DataValue, resp.DataValue)
	assert.Equal(t, expectedResp.UpdatedAt, resp.UpdatedAt)
	assert.Equal(t, expectedResp.DeletedAt, resp.DeletedAt)
	mockConn.AssertExpectations(t)
}

func TestDataServiceClient_Upsert_Error(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewDataServiceClient(mockConn)

	req := &UpsertRequest{
		DataKey:   "test-key",
		DataValue: []byte("test-value"),
	}
	expectedErr := errors.New("connection error")

	mockConn.On("Invoke", mock.Anything, DataService_Upsert_FullMethodName, req, mock.Anything).
		Return(expectedErr)

	resp, err := client.Upsert(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	mockConn.AssertExpectations(t)
}

func TestDataServiceClient_GetUpdates(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewDataServiceClient(mockConn)

	now := time.Now()
	updatedAfter := timestamppb.New(now.Add(-time.Hour))

	req := &GetUpdatesRequest{
		UpdatedAfter: updatedAfter,
	}

	item1 := &DataResponse{
		DataKey:   "key1",
		DataValue: []byte("value1"),
		UpdatedAt: timestamppb.New(now),
	}
	item2 := &DataResponse{
		DataKey:   "key2",
		DataValue: []byte("value2"),
		UpdatedAt: timestamppb.New(now),
	}
	expectedResp := &DataListResponse{
		Items: []*DataResponse{item1, item2},
	}

	mockConn.On("Invoke", mock.Anything, DataService_GetUpdates_FullMethodName, req, mock.Anything).
		Run(func(args mock.Arguments) {
			resp := args.Get(3).(*DataListResponse)
			resp.Items = expectedResp.Items
		}).
		Return(nil)

	resp, err := client.GetUpdates(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, "key1", resp.Items[0].DataKey)
	assert.Equal(t, "key2", resp.Items[1].DataKey)
	mockConn.AssertExpectations(t)
}

func TestDataServiceClient_GetUpdates_Error(t *testing.T) {
	mockConn := new(mockClientConn)
	client := NewDataServiceClient(mockConn)

	req := &GetUpdatesRequest{
		UpdatedAfter: timestamppb.Now(),
	}
	expectedErr := errors.New("connection error")

	mockConn.On("Invoke", mock.Anything, DataService_GetUpdates_FullMethodName, req, mock.Anything).
		Return(expectedErr)

	resp, err := client.GetUpdates(context.Background(), req)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	mockConn.AssertExpectations(t)
}

type mockAuthServer struct {
	mock.Mock
}

func (m *mockAuthServer) Register(ctx context.Context, req *RegisterRequest) (*TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*TokenResponse), args.Error(1)
}

func (m *mockAuthServer) Login(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*TokenResponse), args.Error(1)
}

func (m *mockAuthServer) mustEmbedUnimplementedAuthServiceServer() {}

func TestAuthService_RegisterHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		expectedResp := &TokenResponse{Token: "test-token"}
		req := &RegisterRequest{Login: "test", Password: "pass", MasterPassword: "master"}

		mockServer.On("Register", mock.Anything, req).Return(expectedResp, nil)

		resp, err := _AuthService_Register_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*RegisterRequest)
			if !ok {
				return errors.New("failed to cast to RegisterRequest")
			}
			req.Login = "test"
			req.Password = "pass"
			req.MasterPassword = "master"
			return nil
		}, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockServer.AssertExpectations(t)
	})

	t.Run("decode error", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		decodeErr := errors.New("failed to decode")

		resp, err := _AuthService_Register_Handler(mockServer, context.Background(), func(i interface{}) error {
			return decodeErr
		}, nil)

		assert.Error(t, err)
		assert.Equal(t, decodeErr, err)
		assert.Nil(t, resp)
	})

	t.Run("with interceptor error", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		expectedErr := errors.New("interceptor error")

		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, expectedErr
		}

		resp, err := _AuthService_Register_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*RegisterRequest)
			if !ok {
				return errors.New("failed to cast to RegisterRequest")
			}
			req.Login = "test"
			return nil
		}, interceptor)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("with interceptor success", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		expectedResp := &TokenResponse{Token: "intercepted-token"}

		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			// Verify the info data
			assert.Equal(t, AuthService_Register_FullMethodName, info.FullMethod)
			assert.Equal(t, mockServer, info.Server)

			// Call the handler
			_, err := handler(ctx, req)
			if err != nil {
				return nil, err
			}

			// Return intercepted response
			return expectedResp, nil
		}

		req := &RegisterRequest{Login: "test"}
		mockServer.On("Register", mock.Anything, req).Return(&TokenResponse{Token: "original-token"}, nil)

		resp, err := _AuthService_Register_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*RegisterRequest)
			if !ok {
				return errors.New("failed to cast to RegisterRequest")
			}
			req.Login = "test"
			return nil
		}, interceptor)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockServer.AssertExpectations(t)
	})
}

func TestAuthService_LoginHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		expectedResp := &TokenResponse{Token: "test-token"}
		req := &LoginRequest{Login: "test", Password: "pass", MasterPassword: "master"}

		mockServer.On("Login", mock.Anything, req).Return(expectedResp, nil)

		resp, err := _AuthService_Login_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*LoginRequest)
			if !ok {
				return errors.New("failed to cast to LoginRequest")
			}
			req.Login = "test"
			req.Password = "pass"
			req.MasterPassword = "master"
			return nil
		}, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockServer.AssertExpectations(t)
	})

	t.Run("decode error", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		decodeErr := errors.New("failed to decode")

		resp, err := _AuthService_Login_Handler(mockServer, context.Background(), func(i interface{}) error {
			return decodeErr
		}, nil)

		assert.Error(t, err)
		assert.Equal(t, decodeErr, err)
		assert.Nil(t, resp)
	})

	t.Run("server error", func(t *testing.T) {
		mockServer := new(mockAuthServer)
		expectedErr := errors.New("authentication failed")
		req := &LoginRequest{Login: "test"}

		mockServer.On("Login", mock.Anything, req).Return((*TokenResponse)(nil), expectedErr)

		resp, err := _AuthService_Login_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*LoginRequest)
			if !ok {
				return errors.New("failed to cast to LoginRequest")
			}
			req.Login = "test"
			return nil
		}, nil)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
		mockServer.AssertExpectations(t)
	})
}

type mockDataServer struct {
	mock.Mock
}

func (m *mockDataServer) Upsert(ctx context.Context, req *UpsertRequest) (*DataResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*DataResponse), args.Error(1)
}

func (m *mockDataServer) GetUpdates(ctx context.Context, req *GetUpdatesRequest) (*DataListResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*DataListResponse), args.Error(1)
}

func (m *mockDataServer) mustEmbedUnimplementedDataServiceServer() {}

func TestUnimplementedAuthServiceServer(t *testing.T) {
	server := UnimplementedAuthServiceServer{}

	// Test Register method
	resp, err := server.Register(context.Background(), &RegisterRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
	assert.Equal(t, "method Register not implemented", st.Message())

	// Test Login method
	resp2, err := server.Login(context.Background(), &LoginRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp2)

	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
	assert.Equal(t, "method Login not implemented", st.Message())

	// Test mustEmbedUnimplementedAuthServiceServer method
	server.mustEmbedUnimplementedAuthServiceServer()

	// Test testEmbeddedByValue method
	server.testEmbeddedByValue()
}

func TestUnimplementedDataServiceServer(t *testing.T) {
	server := UnimplementedDataServiceServer{}

	// Test Upsert method
	resp, err := server.Upsert(context.Background(), &UpsertRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
	assert.Equal(t, "method Upsert not implemented", st.Message())

	// Test GetUpdates method
	resp2, err := server.GetUpdates(context.Background(), &GetUpdatesRequest{})
	assert.Error(t, err)
	assert.Nil(t, resp2)

	st, ok = status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.Unimplemented, st.Code())
	assert.Equal(t, "method GetUpdates not implemented", st.Message())

	// Test mustEmbedUnimplementedDataServiceServer method
	server.mustEmbedUnimplementedDataServiceServer()

	// Test testEmbeddedByValue method
	server.testEmbeddedByValue()
}

func TestDataService_UpsertHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockServer := new(mockDataServer)
		now := time.Now()
		updatedAt := timestamppb.New(now)

		req := &UpsertRequest{DataKey: "key", DataValue: []byte("value"), UpdatedAt: updatedAt}
		expectedResp := &DataResponse{DataKey: "key", DataValue: []byte("value"), UpdatedAt: updatedAt}

		mockServer.On("Upsert", mock.Anything, req).Return(expectedResp, nil)

		resp, err := _DataService_Upsert_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*UpsertRequest)
			if !ok {
				return errors.New("failed to cast to UpsertRequest")
			}
			req.DataKey = "key"
			req.DataValue = []byte("value")
			req.UpdatedAt = updatedAt
			return nil
		}, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockServer.AssertExpectations(t)
	})

	t.Run("decode error", func(t *testing.T) {
		mockServer := new(mockDataServer)
		decodeErr := errors.New("failed to decode")

		resp, err := _DataService_Upsert_Handler(mockServer, context.Background(), func(i interface{}) error {
			return decodeErr
		}, nil)

		assert.Error(t, err)
		assert.Equal(t, decodeErr, err)
		assert.Nil(t, resp)
	})

	t.Run("with interceptor error", func(t *testing.T) {
		mockServer := new(mockDataServer)
		expectedErr := errors.New("interceptor error")

		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return nil, expectedErr
		}

		resp, err := _DataService_Upsert_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*UpsertRequest)
			if !ok {
				return errors.New("failed to cast to UpsertRequest")
			}
			req.DataKey = "test"
			return nil
		}, interceptor)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
	})

	t.Run("server error", func(t *testing.T) {
		mockServer := new(mockDataServer)
		expectedErr := errors.New("data storage error")
		req := &UpsertRequest{DataKey: "key"}

		mockServer.On("Upsert", mock.Anything, req).Return((*DataResponse)(nil), expectedErr)

		resp, err := _DataService_Upsert_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*UpsertRequest)
			if !ok {
				return errors.New("failed to cast to UpsertRequest")
			}
			req.DataKey = "key"
			return nil
		}, nil)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
		assert.Nil(t, resp)
		mockServer.AssertExpectations(t)
	})
}

func TestDataService_GetUpdatesHandler(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		mockServer := new(mockDataServer)
		updatedAfter := timestamppb.Now()

		req := &GetUpdatesRequest{UpdatedAfter: updatedAfter}
		expectedResp := &DataListResponse{
			Items: []*DataResponse{
				{DataKey: "key1", DataValue: []byte("value1")},
				{DataKey: "key2", DataValue: []byte("value2")},
			},
		}

		mockServer.On("GetUpdates", mock.Anything, req).Return(expectedResp, nil)

		resp, err := _DataService_GetUpdates_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*GetUpdatesRequest)
			if !ok {
				return errors.New("failed to cast to GetUpdatesRequest")
			}
			req.UpdatedAfter = updatedAfter
			return nil
		}, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockServer.AssertExpectations(t)
	})

	t.Run("decode error", func(t *testing.T) {
		mockServer := new(mockDataServer)
		decodeErr := errors.New("failed to decode")

		resp, err := _DataService_GetUpdates_Handler(mockServer, context.Background(), func(i interface{}) error {
			return decodeErr
		}, nil)

		assert.Error(t, err)
		assert.Equal(t, decodeErr, err)
		assert.Nil(t, resp)
	})

	t.Run("with interceptor", func(t *testing.T) {
		mockServer := new(mockDataServer)
		updatedAfter := timestamppb.Now()
		req := &GetUpdatesRequest{UpdatedAfter: updatedAfter}

		expectedResp := &DataListResponse{
			Items: []*DataResponse{
				{DataKey: "intercepted-key", DataValue: []byte("intercepted-value")},
			},
		}

		interceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			// Verify the info data
			assert.Equal(t, DataService_GetUpdates_FullMethodName, info.FullMethod)
			assert.Equal(t, mockServer, info.Server)

			// Call the handler to make sure it works
			_, err := handler(ctx, req)
			assert.NoError(t, err)

			// Return intercepted response
			return expectedResp, nil
		}

		mockServer.On("GetUpdates", mock.Anything, req).Return(&DataListResponse{Items: []*DataResponse{}}, nil)

		resp, err := _DataService_GetUpdates_Handler(mockServer, context.Background(), func(i interface{}) error {
			req, ok := i.(*GetUpdatesRequest)
			if !ok {
				return errors.New("failed to cast to GetUpdatesRequest")
			}
			req.UpdatedAfter = updatedAfter
			return nil
		}, interceptor)

		assert.NoError(t, err)
		assert.Equal(t, expectedResp, resp)
		mockServer.AssertExpectations(t)
	})
}

func TestMessageTypes(t *testing.T) {
	tests := []struct {
		name     string
		message  interface{}
		validate func(t *testing.T, msg interface{})
	}{
		{
			name: "RegisterRequest",
			message: &RegisterRequest{
				Login:          "testuser",
				Password:       "password123",
				MasterPassword: "master123",
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*RegisterRequest)
				assert.Equal(t, "testuser", m.Login)
				assert.Equal(t, "password123", m.Password)
				assert.Equal(t, "master123", m.MasterPassword)

				// Test getter methods
				assert.Equal(t, "testuser", m.GetLogin())
				assert.Equal(t, "password123", m.GetPassword())
				assert.Equal(t, "master123", m.GetMasterPassword())

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Empty(t, m.Login)
				assert.Empty(t, m.Password)
				assert.Empty(t, m.MasterPassword)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
		{
			name: "LoginRequest",
			message: &LoginRequest{
				Login:          "testuser",
				Password:       "password123",
				MasterPassword: "master123",
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*LoginRequest)
				assert.Equal(t, "testuser", m.Login)
				assert.Equal(t, "password123", m.Password)
				assert.Equal(t, "master123", m.MasterPassword)

				// Test getter methods
				assert.Equal(t, "testuser", m.GetLogin())
				assert.Equal(t, "password123", m.GetPassword())
				assert.Equal(t, "master123", m.GetMasterPassword())

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Empty(t, m.Login)
				assert.Empty(t, m.Password)
				assert.Empty(t, m.MasterPassword)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
		{
			name: "TokenResponse",
			message: &TokenResponse{
				Token: "jwt-token-123",
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*TokenResponse)
				assert.Equal(t, "jwt-token-123", m.Token)

				// Test getter methods
				assert.Equal(t, "jwt-token-123", m.GetToken())

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Empty(t, m.Token)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
		{
			name: "UpsertRequest",
			message: &UpsertRequest{
				DataKey:   "test-key",
				DataValue: []byte("test-value"),
				UpdatedAt: timestamppb.Now(),
				DeletedAt: timestamppb.Now(),
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*UpsertRequest)
				assert.Equal(t, "test-key", m.DataKey)
				assert.Equal(t, []byte("test-value"), m.DataValue)
				assert.NotNil(t, m.UpdatedAt)
				assert.NotNil(t, m.DeletedAt)

				// Test getter methods
				assert.Equal(t, "test-key", m.GetDataKey())
				assert.Equal(t, []byte("test-value"), m.GetDataValue())
				assert.NotNil(t, m.GetUpdatedAt())
				assert.NotNil(t, m.GetDeletedAt())

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Empty(t, m.DataKey)
				assert.Empty(t, m.DataValue)
				assert.Nil(t, m.UpdatedAt)
				assert.Nil(t, m.DeletedAt)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
		{
			name: "GetUpdatesRequest",
			message: &GetUpdatesRequest{
				UpdatedAfter: timestamppb.Now(),
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*GetUpdatesRequest)
				assert.NotNil(t, m.UpdatedAfter)

				// Test getter methods
				assert.NotNil(t, m.GetUpdatedAfter())

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Nil(t, m.UpdatedAfter)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
		{
			name: "DataResponse",
			message: &DataResponse{
				DataKey:   "test-key",
				DataValue: []byte("test-value"),
				UpdatedAt: timestamppb.Now(),
				DeletedAt: timestamppb.Now(),
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*DataResponse)
				assert.Equal(t, "test-key", m.DataKey)
				assert.Equal(t, []byte("test-value"), m.DataValue)
				assert.NotNil(t, m.UpdatedAt)
				assert.NotNil(t, m.DeletedAt)

				// Test getter methods
				assert.Equal(t, "test-key", m.GetDataKey())
				assert.Equal(t, []byte("test-value"), m.GetDataValue())
				assert.NotNil(t, m.GetUpdatedAt())
				assert.NotNil(t, m.GetDeletedAt())

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Empty(t, m.DataKey)
				assert.Empty(t, m.DataValue)
				assert.Nil(t, m.UpdatedAt)
				assert.Nil(t, m.DeletedAt)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
		{
			name: "DataListResponse",
			message: &DataListResponse{
				Items: []*DataResponse{
					{
						DataKey:   "key1",
						DataValue: []byte("value1"),
					},
					{
						DataKey:   "key2",
						DataValue: []byte("value2"),
					},
				},
			},
			validate: func(t *testing.T, msg interface{}) {
				m := msg.(*DataListResponse)
				assert.Len(t, m.Items, 2)
				assert.Equal(t, "key1", m.Items[0].DataKey)
				assert.Equal(t, "key2", m.Items[1].DataKey)

				// Test getter methods
				items := m.GetItems()
				assert.Len(t, items, 2)
				assert.Equal(t, "key1", items[0].DataKey)
				assert.Equal(t, "key2", items[1].DataKey)

				// Test ProtoMessage and ProtoReflect
				assert.NotNil(t, m.ProtoReflect())
				m.ProtoMessage()

				// Test Reset
				m.Reset()
				assert.Empty(t, m.Items)

				// Test descriptor function exists (we don't test the actual value)
				_ = m.Descriptor
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.message)
		})
	}
}

func TestRawDescGZIP(t *testing.T) {
	data := file_gophkeeper_proto_rawDescGZIP()
	assert.NotNil(t, data)
	assert.NotEmpty(t, data)
}
