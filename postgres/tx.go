package postgres

import (
	"context"

	"github.com/csnewman/cuttle"
	"github.com/jackc/pgx/v5"
)

var (
	_ cuttle.RTx = (*RTx)(nil)
	_ cuttle.WTx = (*WTx)(nil)
)

type RTx struct {
	tx pgx.Tx
}

func (t *RTx) Query(ctx context.Context, stmt string, args ...any) (cuttle.Rows, error) {
	res, err := t.tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{res: res}, nil
}

func (t *RTx) QueryRow(ctx context.Context, stmt string, args ...any) (cuttle.Row, error) {
	res, err := t.tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Row{res: res}, nil
}

func (t *RTx) dispatchBatch(ctx context.Context, entries []*cuttle.BatchEntry) error {
	pb := &pgx.Batch{}

	for _, entry := range entries {
		pb.Queue(entry.Stmt, entry.Args...)
	}

	res := t.tx.SendBatch(ctx, pb)
	defer res.Close()

	for _, entry := range entries {
		if entry.ExecHandler != nil {
			ct, err := res.Exec()
			if err := entry.ExecHandler(ctx, &Exec{res: ct}, err); err != nil {
				return err
			}
		} else if entry.QueryHandler != nil {
			r, err := res.Query()
			if err := entry.QueryHandler(ctx, &Rows{res: r}, err); err != nil {
				return err
			}
		} else {
			panic("unknown entry type")
		}
	}

	return nil
}

func (t *RTx) DispatchBatchR(ctx context.Context, b *cuttle.BatchR) error {
	return t.dispatchBatch(ctx, b.Entries)
}

type WTx struct {
	RTx
}

func (t *WTx) Exec(ctx context.Context, stmt string, args ...any) (cuttle.Exec, error) {
	res, err := t.tx.Exec(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Exec{res: res}, nil
}

func (t *WTx) DispatchBatchRW(ctx context.Context, b *cuttle.BatchRW) error {
	return t.dispatchBatch(ctx, b.Entries)
}