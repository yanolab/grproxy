package grproxy

import (
	"context"
	"io"
)

type reader func(b []byte) (int, error)

func (r reader) Read(b []byte) (int, error) {
	return r(b)
}

type writer func(b []byte) (int, error)

func (w writer) Write(b []byte) (int, error) {
	return w(b)
}

func newReceiver(recv func() (*ReadWrite, error)) io.Reader {
	return reader(func(b []byte) (int, error) {
		req, err := recv()
		if err != nil {
			return 0, err
		}

		n := copy(b, req.Buf[0:req.Len])
		return n, nil
	})
}

func newSender(send func(*ReadWrite) error) io.Writer {
	return writer(func(b []byte) (int, error) {
		n := len(b)
		if err := send(&ReadWrite{
			Buf: b,
			Len: int32(n),
		}); err != nil {
			return 0, err
		}

		return n, nil
	})
}

func proxy(ctx context.Context, w io.Writer, r io.Reader, b []byte) error {
	ch := make(chan error)

	go func() {
		defer close(ch)
		_, err := io.CopyBuffer(w, r, b)
		ch <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-ch:
		return err
	}
}
