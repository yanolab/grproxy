package grproxy

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
)

func Test_receiver(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		receiver func() (*ReadWrite, error)
		want     []byte
		wantErr  bool
	}{
		"success": {
			receiver: func() (*ReadWrite, error) {
				return &ReadWrite{
					Buf: bytes.Repeat([]byte("a"), 10),
					Len: 10,
				}, nil
			},
			want: bytes.Repeat([]byte("a"), 10),
		},
		"error": {
			receiver: func() (*ReadWrite, error) {
				return nil, errors.New("error")
			},
			wantErr: true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			r := newReceiver(tc.receiver)
			got := make([]byte, 1024)
			n, err := r.Read(got)
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			} else if err != nil {
				return
			}
			got = got[0:n]
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("unexpected value: %v", got)
			}
		})
	}
}

func Test_sender(t *testing.T) {
	t.Parallel()

	type args struct {
		b []byte
	}
	tests := map[string]struct {
		sender  func(*ReadWrite) error
		args    args
		want    int
		wantErr bool
	}{
		"success": {
			sender: func(*ReadWrite) error {
				return nil
			},
			args: args{
				b: bytes.Repeat([]byte("a"), 10),
			},
			want: 10,
		},
		"error": {
			sender: func(*ReadWrite) error {
				return errors.New("error")
			},
			args: args{
				b: bytes.Repeat([]byte("a"), 10),
			},
			wantErr: true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			w := newSender(tc.sender)
			got, err := w.Write(tc.args.b)
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			} else if err != nil {
				return
			}
			if got != tc.want {
				t.Errorf("unexpected value: %v", got)
			}
		})
	}
}

func Test_proxy(t *testing.T) {
	t.Parallel()

	type args struct {
		ctx context.Context
		w   io.Writer
		r   io.Reader
	}
	tests := map[string]struct {
		args    args
		want    []byte
		wantErr bool
	}{
		"success": {
			args: args{
				ctx: context.TODO(),
				w:   &bytes.Buffer{},
				r:   bytes.NewReader([]byte("abcde")),
			},
			want: []byte("abcde"),
		},
		"write error": {
			args: args{
				ctx: context.TODO(),
				w: writer(func(b []byte) (int, error) {
					return 0, errors.New("error")
				}),
				r: bytes.NewReader([]byte("abcde")),
			},
			wantErr: true,
		},
		"read error": {
			args: args{
				ctx: context.TODO(),
				w:   &bytes.Buffer{},
				r: reader(func(b []byte) (int, error) {
					return 0, errors.New("error")
				}),
			},
			wantErr: true,
		},
		"context cancel": {
			args: args{
				ctx: func() context.Context {
					ctx := context.TODO()
					ctx, cancel := context.WithCancel(ctx)
					cancel()
					return ctx
				}(),
				w: &bytes.Buffer{},
				r: bytes.NewReader([]byte("abcde")),
			},
			wantErr: true,
		},
	}

	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			err := proxy(tc.args.ctx, tc.args.w, tc.args.r, make([]byte, 1024))
			if (err != nil) != tc.wantErr {
				t.Fatal(err)
			} else if err != nil {
				return
			}
			if w, ok := tc.args.w.(*bytes.Buffer); ok {
				got := w.Bytes()
				if !reflect.DeepEqual(got, tc.want) {
					t.Errorf("unexpected value: %v", got)
				}
			}
		})
	}
}
