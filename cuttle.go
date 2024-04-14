package cuttle

import (
	"context"
)

type RTxFunc = func(ctx context.Context, tx RTx) error

type WTxFunc = func(ctx context.Context, tx WTx) error

type DB interface {
	WTx

	RTx(ctx context.Context, f RTxFunc) error

	WTx(ctx context.Context, f WTxFunc) error

	Dialect() Dialect
}

type RTx interface {
	Query(ctx context.Context, stmt string, args ...any) (Rows, error)

	QueryRow(ctx context.Context, stmt string, args ...any) (Row, error)

	DispatchBatchR(ctx context.Context, b *BatchR) error
}

type WTx interface {
	RTx

	Exec(ctx context.Context, stmt string, args ...any) (Exec, error)

	DispatchBatchRW(ctx context.Context, b *BatchRW) error
}

type AsyncHandler[T any] func(ctx context.Context, result T, err error) error

type Exec interface {
	RowsAffected() int64
}

type Row interface{}

type Rows interface{}

type AsyncRTx interface {
	Query(handler AsyncHandler[Rows], stmt string, args ...any)

	QueryRow(handler AsyncHandler[Row], stmt string, args ...any)
}

type AsyncWTx interface {
	AsyncRTx

	Exec(handler AsyncHandler[Exec], stmt string, args ...any)
}
