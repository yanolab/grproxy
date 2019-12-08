package grproxy

import (
	"net"

	grpc "google.golang.org/grpc"
)

type ProxyServer struct {
	service *ProxyServerService
	grpcsrv *grpc.Server
}

func NewProxyServer(grpcsrv *grpc.Server, service *ProxyServerService) *ProxyServer {
	RegisterProxyServiceServer(grpcsrv, service)

	return &ProxyServer{
		service: service,
		grpcsrv: grpcsrv,
	}
}

func (srv *ProxyServer) Serve(lis net.Listener) error {
	return srv.grpcsrv.Serve(lis)
}
