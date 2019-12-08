package grproxy

import (
	"bytes"
	"context"
	"errors"
	"reflect"
	"testing"

	"google.golang.org/grpc"
)

type mockProxyClient struct {
	ProxyServiceClient

	mockConnect func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error)

	mockConnectClient *mockConnectClient
}

func (m *mockProxyClient) Connect(ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
	return m.mockConnect(m, ctx, opts...)
}

type mockConnectClient struct {
	ProxyService_ConnectClient

	mockSend      func(*mockConnectClient, *ReadWrite) error
	mockRecv      func(*mockConnectClient) (*ReadWrite, error)
	mockCloseSend func() error

	wb *bytes.Buffer
	rb *bytes.Buffer
}

func (m *mockConnectClient) Send(rw *ReadWrite) error {
	return m.mockSend(m, rw)
}

func (m *mockConnectClient) Recv() (*ReadWrite, error) {
	return m.mockRecv(m)
}

func (m *mockConnectClient) CloseSend() error {
	return m.mockCloseSend()
}

func Test_ProxyClientService_Bind(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx             context.Context
		mockProxyClient *mockProxyClient
		mockConn        *mockConn
	}
	tests := map[string]struct {
		args       args
		clientWant []byte
		connWant   []byte
		wantErr    bool
	}{
		"success": {
			args: args{
				ctx: context.TODO(),
				mockProxyClient: &mockProxyClient{
					mockConnect: func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
						return m.mockConnectClient, nil
					},
					mockConnectClient: &mockConnectClient{
						mockSend: func(m *mockConnectClient, rw *ReadWrite) error {
							_, err := m.wb.Write(rw.Buf)
							return err
						},
						mockRecv: func(m *mockConnectClient) (*ReadWrite, error) {
							b := make([]byte, 4096)
							n, err := m.rb.Read(b)
							if err != nil {
								return nil, err
							}
							return &ReadWrite{
								Buf: b[:n],
								Len: int32(n),
							}, nil
						},
						mockCloseSend: func() error {
							return nil
						},
						wb: &bytes.Buffer{},
						rb: bytes.NewBuffer(bytes.Repeat([]byte("a"), 10)),
					},
				},
				mockConn: &mockConn{
					mockRead: func(m *mockConn, b []byte) (int, error) {
						return m.rb.Read(b)
					},
					mockWrite: func(conn *mockConn, b []byte) (int, error) {
						return conn.wb.Write(b)
					},
					mockClose: func() error {
						return nil
					},
					wb: &bytes.Buffer{},
					rb: bytes.NewBuffer(bytes.Repeat([]byte("b"), 10)),
				},
			},
			clientWant: bytes.Repeat([]byte("b"), 10),
			connWant:   bytes.Repeat([]byte("a"), 10),
		},
		"connect error": {
			args: args{
				ctx: context.TODO(),
				mockProxyClient: &mockProxyClient{
					mockConnect: func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
						return nil, errors.New("error")
					},
				},
			},
			wantErr: true,
		},
		"send error": {
			args: args{
				ctx: context.TODO(),
				mockProxyClient: &mockProxyClient{
					mockConnect: func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
						return m.mockConnectClient, nil
					},
					mockConnectClient: &mockConnectClient{
						mockSend: func(m *mockConnectClient, rw *ReadWrite) error {
							return errors.New("error")
						},
						mockRecv: func(m *mockConnectClient) (*ReadWrite, error) {
							b := make([]byte, 4096)
							n, err := m.rb.Read(b)
							if err != nil {
								return nil, err
							}
							return &ReadWrite{
								Buf: b[:n],
								Len: int32(n),
							}, nil
						},
						mockCloseSend: func() error {
							return nil
						},
						wb: &bytes.Buffer{},
						rb: bytes.NewBuffer(bytes.Repeat([]byte("a"), 10)),
					},
				},
				mockConn: &mockConn{
					mockRead: func(m *mockConn, b []byte) (int, error) {
						return m.rb.Read(b)
					},
					mockWrite: func(conn *mockConn, b []byte) (int, error) {
						return conn.wb.Write(b)
					},
					mockClose: func() error {
						return nil
					},
					wb: &bytes.Buffer{},
					rb: bytes.NewBuffer(bytes.Repeat([]byte("b"), 10)),
				},
			},
			wantErr: true,
		},
		"recv error": {
			args: args{
				ctx: context.TODO(),
				mockProxyClient: &mockProxyClient{
					mockConnect: func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
						return m.mockConnectClient, nil
					},
					mockConnectClient: &mockConnectClient{
						mockSend: func(m *mockConnectClient, rw *ReadWrite) error {
							_, err := m.wb.Write(rw.Buf)
							return err
						},
						mockRecv: func(m *mockConnectClient) (*ReadWrite, error) {
							return nil, errors.New("error")
						},
						mockCloseSend: func() error {
							return nil
						},
						wb: &bytes.Buffer{},
						rb: bytes.NewBuffer(bytes.Repeat([]byte("a"), 10)),
					},
				},
				mockConn: &mockConn{
					mockRead: func(m *mockConn, b []byte) (int, error) {
						return m.rb.Read(b)
					},
					mockWrite: func(conn *mockConn, b []byte) (int, error) {
						return conn.wb.Write(b)
					},
					mockClose: func() error {
						return nil
					},
					wb: &bytes.Buffer{},
					rb: bytes.NewBuffer(bytes.Repeat([]byte("b"), 10)),
				},
			},
			wantErr: true,
		},
		"read error": {
			args: args{
				ctx: context.TODO(),
				mockProxyClient: &mockProxyClient{
					mockConnect: func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
						return m.mockConnectClient, nil
					},
					mockConnectClient: &mockConnectClient{
						mockSend: func(m *mockConnectClient, rw *ReadWrite) error {
							_, err := m.wb.Write(rw.Buf)
							return err
						},
						mockRecv: func(m *mockConnectClient) (*ReadWrite, error) {
							b := make([]byte, 4096)
							n, err := m.rb.Read(b)
							if err != nil {
								return nil, err
							}
							return &ReadWrite{
								Buf: b[:n],
								Len: int32(n),
							}, nil
						},
						mockCloseSend: func() error {
							return nil
						},
						wb: &bytes.Buffer{},
						rb: bytes.NewBuffer(bytes.Repeat([]byte("a"), 10)),
					},
				},
				mockConn: &mockConn{
					mockRead: func(m *mockConn, b []byte) (int, error) {
						return 0, errors.New("error")
					},
					mockWrite: func(conn *mockConn, b []byte) (int, error) {
						return conn.wb.Write(b)
					},
					mockClose: func() error {
						return nil
					},
					wb: &bytes.Buffer{},
					rb: bytes.NewBuffer(bytes.Repeat([]byte("b"), 10)),
				},
			},
			wantErr: true,
		},
		"write error": {
			args: args{
				ctx: context.TODO(),
				mockProxyClient: &mockProxyClient{
					mockConnect: func(m *mockProxyClient, ctx context.Context, opts ...grpc.CallOption) (ProxyService_ConnectClient, error) {
						return m.mockConnectClient, nil
					},
					mockConnectClient: &mockConnectClient{
						mockSend: func(m *mockConnectClient, rw *ReadWrite) error {
							_, err := m.wb.Write(rw.Buf)
							return err
						},
						mockRecv: func(m *mockConnectClient) (*ReadWrite, error) {
							b := make([]byte, 4096)
							n, err := m.rb.Read(b)
							if err != nil {
								return nil, err
							}
							return &ReadWrite{
								Buf: b[:n],
								Len: int32(n),
							}, nil
						},
						mockCloseSend: func() error {
							return nil
						},
						wb: &bytes.Buffer{},
						rb: bytes.NewBuffer(bytes.Repeat([]byte("a"), 10)),
					},
				},
				mockConn: &mockConn{
					mockRead: func(m *mockConn, b []byte) (int, error) {
						return m.rb.Read(b)
					},
					mockWrite: func(conn *mockConn, b []byte) (int, error) {
						return 0, errors.New("error")
					},
					mockClose: func() error {
						return nil
					},
					wb: &bytes.Buffer{},
					rb: bytes.NewBuffer(bytes.Repeat([]byte("b"), 10)),
				},
			},
			wantErr: true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			svc := NewProxyClientService(nil)
			err := svc.Bind(tc.args.ctx, tc.args.mockProxyClient, tc.args.mockConn)
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			} else if err != nil {
				return
			}

			cases := []struct {
				got  []byte
				want []byte
			}{
				{got: tc.args.mockProxyClient.mockConnectClient.wb.Bytes(), want: tc.clientWant},
				{got: tc.args.mockConn.wb.Bytes(), want: tc.connWant},
			}
			for _, v := range cases {
				if !reflect.DeepEqual(v.got, v.want) {
					t.Errorf("unexpected result got:%s want:%s", v.got, v.want)
				}
			}
		})
	}
}
