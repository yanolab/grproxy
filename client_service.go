package grproxy

import (
	"context"
	"net"
	"sync"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ProxyClientService interface {
	Dial(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)
	Bind(ctx context.Context, proxycli ProxyServiceClient, conn net.Conn) error
}

type proxyClientService struct {
	dialer func(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)
}

func NewProxyClientService(dialer func(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error)) ProxyClientService {
	return &proxyClientService{
		dialer: dialer,
	}
}

func (svc *proxyClientService) Dial(ctx context.Context, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return svc.dialer(ctx, opts...)
}

func (svc *proxyClientService) Bind(ctx context.Context, proxycli ProxyServiceClient, conn net.Conn) error {
	grpccli, err := proxycli.Connect(ctx)
	if err != nil {
		return err
	}

	var once sync.Once
	eg, ctx := errgroup.WithContext(ctx)
	close := func() { grpccli.CloseSend() }
	eg.Go(func() error {
		defer once.Do(close)
		return proxy(ctx, conn, newReceiver(grpccli.Recv), make([]byte, 4096))
	})
	eg.Go(func() error {
		defer once.Do(close)
		return proxy(ctx, newSender(grpccli.Send), conn, make([]byte, 4096))
	})

	return eg.Wait()
}
