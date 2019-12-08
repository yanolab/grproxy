package grproxy

import (
	"bytes"
	"context"
	"errors"
	"net"
	"reflect"
	"testing"
)

type mockServer struct {
	ProxyService_ConnectServer

	mockSend    func(*mockServer, *ReadWrite) error
	mockRecv    func(*mockServer) (*ReadWrite, error)
	mockContext func() context.Context

	wb *bytes.Buffer
	rb *bytes.Buffer
}

func (m *mockServer) Send(rw *ReadWrite) error {
	return m.mockSend(m, rw)
}

func (m *mockServer) Recv() (*ReadWrite, error) {
	return m.mockRecv(m)
}

func (m *mockServer) Context() context.Context {
	return m.mockContext()
}

type mockConn struct {
	net.Conn

	mockRead  func(*mockConn, []byte) (int, error)
	mockWrite func(*mockConn, []byte) (int, error)
	mockClose func() error

	wb *bytes.Buffer
	rb *bytes.Buffer
}

func (m *mockConn) Read(b []byte) (int, error) {
	return m.mockRead(m, b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	return m.mockWrite(m, b)
}

func (m *mockConn) Close() error {
	return m.mockClose()
}

func Test_ProxyService(t *testing.T) {
	t.Parallel()

	type args struct {
		mockServer *mockServer
	}
	tests := map[string]struct {
		args       args
		mockConn   *mockConn
		serverWant []byte
		connWant   []byte
		wantErr    bool
	}{
		"success": {
			args: args{
				mockServer: &mockServer{
					mockSend: func(srv *mockServer, rw *ReadWrite) error {
						_, err := srv.wb.Write(rw.Buf)
						return err
					},
					mockRecv: func(m *mockServer) (*ReadWrite, error) {
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
					mockContext: func() context.Context {
						return context.TODO()
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
			serverWant: bytes.Repeat([]byte("b"), 10),
			connWant:   bytes.Repeat([]byte("a"), 10),
		},
		"recv error": {
			args: args{
				mockServer: &mockServer{
					mockSend: func(srv *mockServer, rw *ReadWrite) error {
						_, err := srv.wb.Write(rw.Buf)
						return err
					},
					mockRecv: func(m *mockServer) (*ReadWrite, error) {
						return nil, errors.New("error")
					},
					mockContext: func() context.Context {
						return context.TODO()
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
			wantErr: true,
		},
		"send error": {
			args: args{
				mockServer: &mockServer{
					mockSend: func(srv *mockServer, rw *ReadWrite) error {
						return errors.New("error")
					},
					mockRecv: func(m *mockServer) (*ReadWrite, error) {
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
					mockContext: func() context.Context {
						return context.TODO()
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
			wantErr: true,
		},
		"write error": {
			args: args{
				mockServer: &mockServer{
					mockSend: func(srv *mockServer, rw *ReadWrite) error {
						_, err := srv.wb.Write(rw.Buf)
						return err
					},
					mockRecv: func(m *mockServer) (*ReadWrite, error) {
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
					mockContext: func() context.Context {
						return context.TODO()
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
			wantErr: true,
		},
		"read error": {
			args: args{
				mockServer: &mockServer{
					mockSend: func(srv *mockServer, rw *ReadWrite) error {
						_, err := srv.wb.Write(rw.Buf)
						return err
					},
					mockRecv: func(m *mockServer) (*ReadWrite, error) {
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
					mockContext: func() context.Context {
						return context.TODO()
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
			wantErr: true,
		},
		"context cancel": {
			args: args{
				mockServer: &mockServer{
					mockSend: func(srv *mockServer, rw *ReadWrite) error {
						_, err := srv.wb.Write(rw.Buf)
						return err
					},
					mockRecv: func(m *mockServer) (*ReadWrite, error) {
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
					mockContext: func() context.Context {
						ctx := context.TODO()
						ctx, cancel := context.WithCancel(ctx)
						cancel()
						return ctx
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
			serverWant: bytes.Repeat([]byte("b"), 10),
			connWant:   bytes.Repeat([]byte("a"), 10),
			wantErr:    true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			svc := NewProxyServerService(
				func(ctx context.Context) (net.Conn, error) {
					return tc.mockConn, nil
				},
			)
			err := svc.Connect(tc.args.mockServer)
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			} else if err != nil {
				return
			}

			cases := []struct {
				got  []byte
				want []byte
			}{
				{got: tc.args.mockServer.wb.Bytes(), want: tc.serverWant},
				{got: tc.mockConn.wb.Bytes(), want: tc.connWant},
			}
			for _, v := range cases {
				if !reflect.DeepEqual(v.got, v.want) {
					t.Errorf("unexpected result got:%s want:%s", v.got, v.want)
				}
			}
		})
	}
}
