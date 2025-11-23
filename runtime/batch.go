package runtime

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Batch represents a batch of operations
type Batch struct {
	pool  *pgxpool.Pool
	batch *pgx.Batch
}

// NewBatch creates a new batch
func NewBatch(pool *pgxpool.Pool) *Batch {
	return &Batch{
		pool:  pool,
		batch: &pgx.Batch{},
	}
}

// Queue adds a query to the batch
func (b *Batch) Queue(sql string, args ...interface{}) {
	b.batch.Queue(sql, args...)
}

// Send sends the batch and returns results
func (b *Batch) Send(ctx context.Context) (BatchResults, error) {
	br := b.pool.SendBatch(ctx, b.batch)
	return &batchResults{br: br}, nil
}

// BatchResults represents batch results
type BatchResults interface {
	// Exec reads the results from the next query in the batch
	Exec() (pgx.CommandTag, error)

	// Query reads the results from the next query in the batch
	Query() (pgx.Rows, error)

	// QueryRow reads the results from the next query in the batch
	QueryRow() pgx.Row

	// Close closes the batch results
	Close() error
}

type batchResults struct {
	br pgx.BatchResults
}

func (r *batchResults) Exec() (pgx.CommandTag, error) {
	return r.br.Exec()
}

func (r *batchResults) Query() (pgx.Rows, error) {
	return r.br.Query()
}

func (r *batchResults) QueryRow() pgx.Row {
	return r.br.QueryRow()
}

func (r *batchResults) Close() error {
	return r.br.Close()
}

// Transaction support

// Tx wraps a pgx transaction
type Tx struct {
	tx pgx.Tx
}

// BeginTx starts a new transaction
func BeginTx(ctx context.Context, pool *pgxpool.Pool) (*Tx, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return &Tx{tx: tx}, nil
}

// Commit commits the transaction
func (t *Tx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction
func (t *Tx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// Exec executes a query in the transaction
func (t *Tx) Exec(ctx context.Context, sql string, args ...interface{}) (pgx.CommandTag, error) {
	return t.tx.Exec(ctx, sql, args...)
}

// Query executes a query in the transaction
func (t *Tx) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

// QueryRow executes a query in the transaction
func (t *Tx) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}
