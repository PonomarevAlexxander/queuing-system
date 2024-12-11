package grpccontroller

import (
	"context"
	"net"

	"google.golang.org/grpc"
)

type grpcController struct {
	server     *grpc.Server
	connection net.Listener
}

func NewGrpcController(server *grpc.Server, connection net.Listener) *grpcController {
	return &grpcController{
		server:     server,
		connection: connection,
	}
}

func (gc *grpcController) Run(ctx context.Context) error {
	go gc.server.Serve(gc.connection)
	<-ctx.Done()
	return nil
}

func (gc *grpcController) Stop() {
	gc.server.Stop()
}
