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

func (t *RTx) QueryFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Rows], stmt string, args ...any) error {
	res, err := t.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (t *RTx) Query(ctx context.Context, stmt string, args ...any) (cuttle.Rows, error) {
	res, err := t.tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{res: res}, nil
}

func (t *RTx) QueryRowFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Row], stmt string, args ...any) error {
	res, err := t.QueryRow(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (t *RTx) QueryRow(ctx context.Context, stmt string, args ...any) (cuttle.Row, error) {
	res, err := t.tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	// Next loads the row into memory, making it safe to read after closing the reader
	if !res.Next() {
		if res.Err() == nil {
			return nil, cuttle.ErrNoRows
		}

		return nil, res.Err()
	}

	res.Close()

	if res.Err() != nil {
		return nil, res.Err()
	}

	return &Row{res: res}, nil
}

func (t *RTx) DispatchBatchR(ctx context.Context, b *cuttle.BatchR) error {
	return t.dispatchBatch(ctx, b.Entries)
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
		} else if entry.QueryRowHandler != nil {
			r, err := res.Query()
			if err == nil {
				err = r.Err()
			}

			if err == nil && !r.Next() && r.Err() == nil {
				err = cuttle.ErrNoRows
			}

			if r != nil {
				r.Close()
			}

			if err == nil {
				err = r.Err()
			}

			var row *Row

			if err == nil {
				row = &Row{res: r}
			}

			if err := entry.QueryRowHandler(ctx, row, err); err != nil {
				return err
			}
		} else {
			panic("unknown entry type")
		}
	}

	return res.Close()
}

type WTx struct {
	RTx
}

func (t *WTx) ExecFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Exec], stmt string, args ...any) error {
	res, err := t.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
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
