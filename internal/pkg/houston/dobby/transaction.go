package dobby

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Serializable    TxIsoLevel = "serializable"
	RepeatableRead  TxIsoLevel = "repeatable read"
	ReadCommitted   TxIsoLevel = "read committed"
	ReadUncommitted TxIsoLevel = "read uncommitted"
)

type Executor interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
	SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type TxIsoLevel string

type TxOptions struct {
	IsoLevel TxIsoLevel
}

type txKey struct{}

type PGXTransactor struct {
	db *pgxpool.Pool
}

func NewPGXTransactor(db *pgxpool.Pool) *PGXTransactor {
	return &PGXTransactor{db: db}
}

func (t *PGXTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error, opts TxOptions) error {
	tx, err := t.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.TxIsoLevel(opts.IsoLevel),
	})
	if err != nil {
		return fmt.Errorf("can't begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	err = fn(injectPGXTx(ctx, tx))
	if err != nil {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
			return fmt.Errorf("can't rollback transaction: %w, initial error: %v", rollbackErr, err)
		}

		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("can't commit transaction: %w", err)
	}

	return nil
}

func injectPGXTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func ExtractPGXTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return nil
}
