package grpc

import (
	"context"
	"fmt"
	"os"

	console_api "github.com/nextsurfer/oracle/api/console"
	"google.golang.org/grpc"
)

func DialConnectorGrpc(opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return grpc.Dial(fmt.Sprintf("%s:%s", os.Getenv("CONSOLE_HOST"), os.Getenv("CONSOLE_GRPC_PORT")), grpc.WithInsecure())
}

func RegisterGatewayNode(ctx context.Context, req *console_api.RegisterGatewayNodeRequest) (*console_api.RegisterGatewayNodeResponse, error) {
	conn, err := DialConnectorGrpc()
	if err != nil {
		return nil, err
	}
	client := console_api.NewConsoleServiceClient(conn)
	return client.RegisterGatewayNode(ctx, req)
}

func UpsertService(ctx context.Context, req *console_api.UpsertServiceRequest) (*console_api.UpsertServiceResponse, error) {
	conn, err := DialConnectorGrpc()
	if err != nil {
		return nil, err
	}
	client := console_api.NewConsoleServiceClient(conn)
	return client.UpsertService(ctx, req)
}
