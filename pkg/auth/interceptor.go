// Package auth provides authentication and authorization utilities for gRPC and HTTP services.
package auth

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// JWTClientInterceptor creates a gRPC client interceptor that automatically attaches JWT tokens to outgoing requests.
// extracts the JWT token from the context and adds it to the authorization metadata header.
func JWTClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if token := ctx.Value("jwt_token"); token != nil {
			ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token.(string))
			log.Printf("🔐 Auto-forwarding JWT token for gRPC call: %s", method)
		}

		// Call the original method with enhanced context
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// NewAuthenticatedGRPCConnection establishes a gRPC client connection with JWT authentication interceptor.
// creates a connection to the target server with automatic JWT token forwarding for all requests.
func NewAuthenticatedGRPCConnection(target string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(
		target,
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(JWTClientInterceptor()),
	)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
