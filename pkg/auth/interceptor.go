package auth

import (
	"context"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// JWTClientInterceptor automatically adds JWT token to all gRPC requests
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
