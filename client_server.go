package grproxy

import (
	"context"
	"net"
)

type ProxyClientServer struct {
	service ProxyClientService
}

func NewProxyClientServer(service ProxyClientService) *ProxyClientServer {
	return &ProxyClientServer{service: service}
}

func (srv *ProxyClientServer) Serve(lis net.Listener) error {
	for {
		conn, err := lis.Accept()
		if err != nil {
			continue
		}

		go func() {
			defer conn.Close()

			ctx := context.Background()
			grpcconn, err := srv.service.Dial(ctx)
			if err != nil {
				return
			}
			defer grpcconn.Close()

			if err := srv.service.Bind(ctx, NewProxyServiceClient(grpcconn), conn); err != nil {
				return
			}
		}()
	}
}
