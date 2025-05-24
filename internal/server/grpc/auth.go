package grpc

import (
	"context"
	"strings"

	"github.com/m1khal3v/gophkeeper/internal/server/jwt"
	"github.com/m1khal3v/gophkeeper/internal/server/manager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type AuthInterceptor struct {
	userManager UserManagerInterface
}

func NewAuthInterceptor(um *manager.UserManager) *AuthInterceptor {
	return &AuthInterceptor{
		userManager: um,
	}
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is missing")
		}

		token := strings.TrimPrefix(authHeader[0], "Bearer ")
		claims, err := i.userManager.DecodeToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, userClaimsKey{}, claims)
		return handler(ctx, req)
	}
}

type userClaimsKey struct{}

func GetClaimsFromContext(ctx context.Context) (*jwt.Claims, error) {
	val := ctx.Value(userClaimsKey{})
	if val == nil {
		return nil, status.Error(codes.Unauthenticated, "no auth info in context")
	}
	return val.(*jwt.Claims), nil
}
