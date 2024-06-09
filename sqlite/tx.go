package sqlite

import (
	"context"

	"github.com/csnewman/cuttle"
	"github.com/tailscale/sqlite/sqlitepool"
)

var (
	_ cuttle.RTx = (*RTx)(nil)
	_ cuttle.WTx = (*WTx)(nil)
)

type RTx struct {
	tx *sqlitepool.Rx
}

func (r *RTx) QueryFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Rows], stmt string, args ...any) error {
	res, err := r.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (r *RTx) Query(_ context.Context, stmt string, args ...any) (cuttle.Rows, error) {
	res, err := r.tx.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{res: res}, nil
}

func (r *RTx) QueryRowFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Row], stmt string, args ...any) error {
	res, err := r.QueryRow(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (r *RTx) QueryRow(_ context.Context, stmt string, args ...any) (cuttle.Row, error) {
	row := r.tx.QueryRow(stmt, args...)
	if err := row.Err(); err != nil {
		return nil, err
	}

	return &Row{row: row}, nil
}

func (r *RTx) DispatchBatchR(ctx context.Context, b *cuttle.BatchR) error {
	for _, e := range b.Entries {
		if e.QueryRowHandler != nil {
			res, err := r.QueryRow(ctx, e.Stmt, e.Args...)

			if err := e.QueryRowHandler(ctx, res, err); err != nil {
				return err
			}
		} else if e.QueryHandler != nil {
			res, err := r.Query(ctx, e.Stmt, e.Args...)

			if err := e.QueryHandler(ctx, res, err); err != nil {
				return err
			}
		} else {
			panic("unknown entry type")
		}
	}

	return nil
}

type WTx struct {
	tx *sqlitepool.Tx
}

func (w *WTx) QueryFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Rows], stmt string, args ...any) error {
	res, err := w.Query(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (w *WTx) Query(_ context.Context, stmt string, args ...any) (cuttle.Rows, error) {
	res, err := w.tx.Query(stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{res: res}, nil
}

func (w *WTx) QueryRowFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Row], stmt string, args ...any) error {
	res, err := w.QueryRow(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (w *WTx) QueryRow(_ context.Context, stmt string, args ...any) (cuttle.Row, error) {
	row := w.tx.QueryRow(stmt, args...)
	if err := row.Err(); err != nil {
		return nil, err
	}

	return &Row{row: row}, nil
}

func (w *WTx) ExecFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Exec], stmt string, args ...any) error {
	res, err := w.Exec(ctx, stmt, args...)
	if err != nil {
		return err
	}

	return handler(ctx, res)
}

func (w *WTx) Exec(_ context.Context, stmt string, args ...any) (cuttle.Exec, error) {
	res, err := w.tx.ExecRes(stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Exec{rowsAffected: res}, nil
}

func (w *WTx) DispatchBatchR(ctx context.Context, b *cuttle.BatchR) error {
	for _, e := range b.Entries {
		if e.ExecHandler != nil {
			res, err := w.Exec(ctx, e.Stmt, e.Args...)

			if err := e.ExecHandler(ctx, res, err); err != nil {
				return err
			}
		} else if e.QueryRowHandler != nil {
			res, err := w.QueryRow(ctx, e.Stmt, e.Args...)

			if err := e.QueryRowHandler(ctx, res, err); err != nil {
				return err
			}
		} else if e.QueryHandler != nil {
			res, err := w.Query(ctx, e.Stmt, e.Args...)

			if err := e.QueryHandler(ctx, res, err); err != nil {
				return err
			}
		} else {
			panic("unknown entry type")
		}
	}

	return nil
}

func (w *WTx) DispatchBatchRW(ctx context.Context, b *cuttle.BatchRW) error {
	for _, e := range b.Entries {
		if e.ExecHandler != nil {
			res, err := w.Exec(ctx, e.Stmt, e.Args...)

			if err := e.ExecHandler(ctx, res, err); err != nil {
				return err
			}
		} else if e.QueryRowHandler != nil {
			res, err := w.QueryRow(ctx, e.Stmt, e.Args...)

			if err := e.QueryRowHandler(ctx, res, err); err != nil {
				return err
			}
		} else if e.QueryHandler != nil {
			res, err := w.Query(ctx, e.Stmt, e.Args...)

			if err := e.QueryHandler(ctx, res, err); err != nil {
				return err
			}
		} else {
			panic("unknown entry type")
		}
	}

	return nil
}
