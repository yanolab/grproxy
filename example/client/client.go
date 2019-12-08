package main

import (
	"context"
	"log"
	"net"

	"github.com/yanolab/grproxy"
	"google.golang.org/grpc"
)

type wrapper struct {
	grpc.ClientStream
	method string
}

func (w *wrapper) CloseSend() error {
	err := w.ClientStream.CloseSend()
	log.Printf("finished method:%s error:%v", w.method, err)
	return err
}

func logInterceptor(
	ctx context.Context,
	desc *grpc.StreamDesc,
	conn *grpc.ClientConn,
	method string,
	streamer grpc.Streamer,
	opts ...grpc.CallOption,
) (grpc.ClientStream, error) {
	log.Printf("start method:%s ctx:%v", method, ctx)
	cs, err := streamer(ctx, desc, conn, method, opts...)
	return &wrapper{ClientStream: cs, method: method}, err
}

func main() {
	lis, err := net.Listen("tcp", ":3333")
	if err != nil {
		log.Fatal(err)
	}
	log.Print("listen", lis.Addr())

	dialer := func(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
		return grpc.Dial(
			":3000",
			grpc.WithInsecure(),
			grpc.WithStreamInterceptor(logInterceptor),
		)
	}

	srv := grproxy.NewProxyClientServer(
		grproxy.NewProxyClientService(dialer),
	)
	log.Println(srv.Serve(lis))
}
