package postgres

import (
	"context"

	"github.com/csnewman/cuttle"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	_ cuttle.DB  = (*DB)(nil)
	_ cuttle.RTx = (*RTx)(nil)
	_ cuttle.WTx = (*WTx)(nil)
)

type DB struct {
	pool *pgxpool.Pool
}

func FromPool(pool *pgxpool.Pool) *DB {
	return &DB{
		pool: pool,
	}
}

func (d *DB) Query(ctx context.Context, stmt string, args ...any) (cuttle.Rows, error) {
	var res cuttle.Rows

	err := d.WTx(ctx, func(ctx context.Context, tx cuttle.WTx) error {
		innerRes, err := tx.Query(ctx, stmt, args...)

		res = innerRes

		return err
	})

	return res, err
}

func (d *DB) QueryRow(ctx context.Context, stmt string, args ...any) (cuttle.Row, error) {
	var res cuttle.Row

	err := d.WTx(ctx, func(ctx context.Context, tx cuttle.WTx) error {
		innerRes, err := tx.QueryRow(ctx, stmt, args...)

		res = innerRes

		return err
	})

	return res, err
}

func (d *DB) Exec(ctx context.Context, stmt string, args ...any) (cuttle.Exec, error) {
	var res cuttle.Exec

	err := d.WTx(ctx, func(ctx context.Context, tx cuttle.WTx) error {
		innerRes, err := tx.Exec(ctx, stmt, args...)

		res = innerRes

		return err
	})

	return res, err
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
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (d *DB) Dialect() cuttle.Dialect {
	return cuttle.DialectPostgres
}

type RTx struct {
	tx pgx.Tx
}

type Rows struct {
	res pgx.Rows
}

func (t *RTx) Query(ctx context.Context, stmt string, args ...any) (cuttle.Rows, error) {
	res, err := t.tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Rows{res: res}, nil
}

type Row struct {
	res pgx.Rows
}

func (t *RTx) QueryRow(ctx context.Context, stmt string, args ...any) (cuttle.Row, error) {
	res, err := t.tx.Query(ctx, stmt, args...)
	if err != nil {
		return nil, err
	}

	return &Row{res: res}, nil
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

type Exec struct {
	res pgconn.CommandTag
}

func (e *Exec) RowsAffected() int64 {
	return e.res.RowsAffected()
}
