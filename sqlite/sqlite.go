package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/csnewman/cuttle"
	"github.com/tailscale/sqlite/sqliteh"
	"github.com/tailscale/sqlite/sqlitepool"
)

var _ cuttle.DB = (*DB)(nil)

type DB struct {
	pool *sqlitepool.Pool
}

func Open(filename string, poolSize int) (*DB, error) {
	pool, err := sqlitepool.NewPool(filename, poolSize, func(db sqliteh.DB) error {
		return nil
	}, nil)
	if err != nil {
		return nil, err
	}

	return &DB{pool: pool}, nil
}

func (d *DB) ExecFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Exec], stmt string, args ...any) error {
	return d.WTx(ctx, func(ctx context.Context, tx cuttle.WTx) error {
		res, err := tx.Exec(ctx, stmt, args...)
		if err != nil {
			return err
		}

		return handler(ctx, res)
	})
}

func (d *DB) QueryFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Rows], stmt string, args ...any) error {
	return d.RTx(ctx, func(ctx context.Context, tx cuttle.RTx) error {
		res, err := tx.Query(ctx, stmt, args...)
		if err != nil {
			return err
		}

		return handler(ctx, res)
	})
}

func (d *DB) QueryRowFunc(ctx context.Context, handler cuttle.TxFunc[cuttle.Row], stmt string, args ...any) error {
	return d.RTx(ctx, func(ctx context.Context, tx cuttle.RTx) error {
		res, err := tx.QueryRow(ctx, stmt, args...)
		if err != nil {
			return err
		}

		return handler(ctx, res)
	})
}

func (d *DB) DispatchBatchR(ctx context.Context, b *cuttle.BatchR) error {
	return d.RTx(ctx, func(ctx context.Context, tx cuttle.RTx) error {
		return tx.DispatchBatchR(ctx, b)
	})
}

func (d *DB) DispatchBatchRW(ctx context.Context, b *cuttle.BatchRW) error {
	return d.WTx(ctx, func(ctx context.Context, tx cuttle.WTx) error {
		return tx.DispatchBatchRW(ctx, b)
	})
}

func (d *DB) RTx(ctx context.Context, f cuttle.RTxFunc) error {
	tx, err := d.pool.BeginRx(ctx, "rtx")
	if err != nil {
		return err
	}

	defer tx.Rollback()

	return f(ctx, &RTx{tx: tx})
}

func (d *DB) WTx(ctx context.Context, f cuttle.WTxFunc) error {
	tx, err := d.pool.BeginTx(ctx, "wtx")
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := f(ctx, &WTx{tx: tx}); err != nil {
		return fmt.Errorf("error during tx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("error during commit: %w", err)
	}

	return nil
}

func (d *DB) Dialect() cuttle.Dialect {
	return cuttle.DialectSQLite
}

type Rows struct {
	res *sqlitepool.Rows
}

func (r *Rows) Close() error {
	err := r.res.Close()
	if err != nil {
		return err
	}

	return errors.Join(err, r.res.Err())
}

func (r *Rows) Next(dest ...any) (bool, error) {
	if r.res.Err() != nil {
		return false, r.Close()
	}

	if !r.res.Next() {
		return false, r.Close()
	}

	if err := r.res.Scan(dest...); err != nil {
		_ = r.Close()

		return false, err
	}

	return true, nil
}

type Row struct {
	row *sqlitepool.Row
}

func (r *Row) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}

type Exec struct {
	rowsAffected int64
}

func (e *Exec) RowsAffected() int64 {
	return e.rowsAffected
}
