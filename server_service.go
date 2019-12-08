package grproxy

import (
	"context"
	"net"

	"golang.org/x/sync/errgroup"
)

type ProxyServerService struct {
	dialer func(ctx context.Context) (net.Conn, error)
}

func NewProxyServerService(dialer func(ctx context.Context) (net.Conn, error)) *ProxyServerService {
	return &ProxyServerService{
		dialer: dialer,
	}
}

func (svc *ProxyServerService) Connect(srv ProxyService_ConnectServer) error {
	ctx := srv.Context()
	conn, err := svc.dialer(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return proxy(ctx, conn, newReceiver(srv.Recv), make([]byte, 4096))
	})
	eg.Go(func() error {
		return proxy(ctx, newSender(srv.Send), conn, make([]byte, 4096))
	})

	return eg.Wait()
}
