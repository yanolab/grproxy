package main

import (
	"context"
	"database/sql"
	"log"
	"net"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/yanolab/grproxy"
	"google.golang.org/grpc"
)

type wrapper struct {
	grpc.ClientStream
	method string
}

func (w *wrapper) CloseSend() error {
	err := w.ClientStream.CloseSend()
	//log.Printf("finished method:%s error:%v", w.method, err)
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
	//log.Printf("start method:%s ctx:%v", method, ctx)
	cs, err := streamer(ctx, desc, conn, method, opts...)
	return &wrapper{ClientStream: cs, method: method}, err
}

func logServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	//log.Printf("start method:%s, ctx:%v", info.FullMethod, ss.Context())
	err := handler(srv, ss)
	//log.Printf("finish method:%s error:%v", info.FullMethod, err)
	return err
}

func setupClient() {
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

func setupServer() {
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
		grpc.NewServer(grpc.StreamInterceptor(logServerInterceptor)),
		grproxy.NewProxyServerService(dialer),
	)
	log.Println(srv.Serve(lis))
}

var db *sql.DB

func init() {
	go func() { setupClient() }()
	go func() { setupServer() }()

	time.Sleep(10)

	_db, err := sql.Open("mysql", "root:pass@tcp(127.0.0.1:3333)/testdb")
	if err != nil {
		panic(err)
	}
	db = _db

	_, err = db.ExecContext(context.TODO(), "DELETE FROM tests WHERE true")
	if err != nil {
		panic(err)
	}

	ret, err := db.ExecContext(context.TODO(), "INSERT INTO tests(id) VALUE(?)", "testid")
	if err != nil {
		panic(err)
	}
	if n, err := ret.RowsAffected(); err != nil || n != 1 {
		panic(err)
	}
}

func BenchmarkAppendSelect(b *testing.B) {
	q := "SELECT * FROM tests WHERE id = 'testid'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		row := db.QueryRowContext(context.TODO(), q)
		if row == nil {
			b.Fatalf("failed to read data")
		}
	}
}

func BenchmarkAppendSelectWithTx(b *testing.B) {
	q := "SELECT * FROM tests WHERE id = 'testid'"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.TODO()
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			b.Fatalf("failed to begin tx: %v", err)
		}
		_, err = tx.ExecContext(ctx, q)
		if err != nil {
			b.Fatalf("failed to read data: %v", err)
		}
		if err := tx.Commit(); err != nil {
			b.Fatalf("failed to commit: %v", err)
		}
	}
}

func BenchmarkAppendWriteWithTx(b *testing.B) {
	q := "INSERT INTO tests(id) VALUE(?)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.TODO()
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: false})
		if err != nil {
			b.Fatalf("failed to begin tx: %v", err)
		}
		id, err := uuid.NewUUID()
		if err != nil {
			b.Fatalf("failed to create uuid: %v", err)
		}
		_, err = tx.ExecContext(ctx, q, id.String())
		if err != nil {
			b.Fatalf("failed to read data: %v", err)
		}
		if err := tx.Commit(); err != nil {
			b.Fatalf("failed to commit: %v", err)
		}
	}
}
