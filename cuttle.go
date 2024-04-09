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
}

type WTx interface {
	RTx

	Exec(ctx context.Context, stmt string, args ...any) (Exec, error)
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

type BatchEntry struct {
	Stmt         string
	Args         []any
	ExecHandler  AsyncHandler[Exec]
	QueryHandler AsyncHandler[Rows]
}

type BatchRW struct {
	Entries []*BatchEntry
}

func NewBatchRW() *BatchRW {
	return &BatchRW{}
}

func (b *BatchRW) Exec(handler AsyncHandler[Exec], stmt string, args ...any) {
	b.Entries = append(b.Entries, &BatchEntry{
		Stmt:        stmt,
		Args:        args,
		ExecHandler: handler,
	})
}

func (b *BatchRW) Query(handler AsyncHandler[Rows], stmt string, args ...any) {
	b.Entries = append(b.Entries, &BatchEntry{
		Stmt:         stmt,
		Args:         args,
		QueryHandler: handler,
	})
}

func (b *BatchRW) QueryRow(handler AsyncHandler[Row], stmt string, args ...any) { //nolint:revive
	panic("implement me")
}

type BatchR struct {
	Entries []*BatchEntry
}

func NewBatchR() *BatchR {
	return &BatchR{}
}

func (b *BatchR) Query(handler AsyncHandler[Rows], stmt string, args ...any) {
	b.Entries = append(b.Entries, &BatchEntry{
		Stmt:         stmt,
		Args:         args,
		QueryHandler: handler,
	})
}

func (b *BatchR) QueryRow(handler AsyncHandler[Row], stmt string, args ...any) { //nolint:revive
	panic("implement me")
}

var (
	_ AsyncRTx = (*BatchR)(nil)
	_ AsyncRTx = (*BatchRW)(nil)
	_ AsyncWTx = (*BatchRW)(nil)
)
