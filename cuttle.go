package cuttle

import (
	"context"
	"errors"
)

var ErrNoRows = errors.New("no rows")

type RTxFunc = func(ctx context.Context, tx RTx) error

type WTxFunc = func(ctx context.Context, tx WTx) error

type DB interface {
	WTxFuncer

	RTx(ctx context.Context, f RTxFunc) error

	WTx(ctx context.Context, f WTxFunc) error

	Dialect() Dialect
}

type TxFunc[T any] func(ctx context.Context, result T) error

type RTxFuncer interface {
	QueryFunc(ctx context.Context, handler TxFunc[Rows], stmt string, args ...any) error

	QueryRowFunc(ctx context.Context, handler TxFunc[Row], stmt string, args ...any) error

	DispatchBatchR(ctx context.Context, b *BatchR) error
}

type RTx interface {
	RTxFuncer

	Query(ctx context.Context, stmt string, args ...any) (Rows, error)

	QueryRow(ctx context.Context, stmt string, args ...any) (Row, error)
}

type WTxFuncer interface {
	RTxFuncer

	ExecFunc(ctx context.Context, handler TxFunc[Exec], stmt string, args ...any) error

	DispatchBatchRW(ctx context.Context, b *BatchRW) error
}

type WTx interface {
	RTx
	WTxFuncer

	Exec(ctx context.Context, stmt string, args ...any) (Exec, error)
}

type AsyncHandler[T any] func(ctx context.Context, result T, err error) error

type Exec interface {
	RowsAffected() int64
}

type Row interface {
	Scan(dest ...any) error
}

type Rows interface{}

type AsyncRTx interface {
	Query(handler AsyncHandler[Rows], stmt string, args ...any)

	QueryRow(handler AsyncHandler[Row], stmt string, args ...any)
}

type AsyncWTx interface {
	AsyncRTx

	Exec(handler AsyncHandler[Exec], stmt string, args ...any)
}
