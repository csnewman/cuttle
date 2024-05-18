package postgres

import (
	"context"
	"fmt"

	"github.com/csnewman/cuttle"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var _ cuttle.DB = (*DB)(nil)

type DB struct {
	pool *pgxpool.Pool
}

func FromPool(pool *pgxpool.Pool) *DB {
	return &DB{
		pool: pool,
	}
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
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx) //nolint:errcheck

	if err := f(ctx, &RTx{tx: tx}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (d *DB) WTx(ctx context.Context, f cuttle.WTxFunc) error {
	tx, err := d.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx) //nolint:errcheck

	if err := f(ctx, &WTx{RTx{tx: tx}}); err != nil {
		return fmt.Errorf("error during tx: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error during commit: %w", err)
	}

	return nil
}

func (d *DB) Dialect() cuttle.Dialect {
	return cuttle.DialectPostgres
}

type Rows struct {
	res pgx.Rows
}

type Row struct {
	res pgx.Rows
}

func (r *Row) Scan(dest ...any) error {
	return r.res.Scan(dest...)
}

type Exec struct {
	res pgconn.CommandTag
}

func (e *Exec) RowsAffected() int64 {
	return e.res.RowsAffected()
}
