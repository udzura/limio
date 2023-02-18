package limio

import (
	"context"
	"io"

	"golang.org/x/time/rate"
)

var BurstLimit = 1024 * 1024 * 1024

type Pool struct {
	limiter *rate.Limiter
	ctx     context.Context
}

func NewPool(bps float64) *Pool {
	return &Pool{
		limiter: rate.NewLimiter(rate.Limit(bps), BurstLimit),
		ctx:     context.Background(),
	}
}

func NewPoolWithContext(ctx context.Context, bps float64) *Pool {
	return &Pool{
		limiter: rate.NewLimiter(rate.Limit(bps), BurstLimit),
		ctx:     ctx,
	}
}

func (p *Pool) GetWriteCloser(w io.WriteCloser) *WriteCloser {
	return &WriteCloser{
		w:       w,
		limiter: p.limiter,
		ctx:     p.ctx,
	}
}

type WriteCloser struct {
	w       io.WriteCloser
	limiter *rate.Limiter
	ctx     context.Context
}

func (w *WriteCloser) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil {
		return n, err
	}
	if err := w.limiter.WaitN(w.ctx, n); err != nil {
		return n, err
	}
	return n, err
}

func (w *WriteCloser) Close() error {
	return w.w.Close()
}
