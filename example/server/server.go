package main

import (
	"context"
	"log"
	"net"

	"github.com/yanolab/grproxy"
	"google.golang.org/grpc"
)

func logInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Printf("start method:%s, ctx:%v", info.FullMethod, ss.Context())
	err := handler(srv, ss)
	log.Printf("finish method:%s error:%v", info.FullMethod, err)
	return err
}

func main() {
	lis, err := net.Listen("tcp", ":3000")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("listen", lis.Addr())

	dialer := func(ctx context.Context) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "tcp", ":3306")
	}

	srv := grproxy.NewProxyServer(
		grpc.NewServer(grpc.StreamInterceptor(logInterceptor)),
		grproxy.NewProxyServerService(dialer),
	)
	log.Println(srv.Serve(lis))
}
