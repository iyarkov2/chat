package util

import (
	"context"
	"database/sql"
)

func WithConnection(ctx context.Context, db *sql.DB, function func(conn *sql.Conn) error) error {
	conn, err := db.Conn(ctx)
	defer CloseQuiet(ctx, "connection", conn)
	if err != nil {
		return err
	}
	return function(conn)
}

func WithTx(ctx context.Context, db *sql.DB, function func(conn *sql.Tx) error) error {
	return WithConnection(ctx, db, func(conn *sql.Conn) error {
		// Start new transaction. TODO - add options to allow read only transactions
		opts := sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			ReadOnly: false,
		}
		tx, err := conn.BeginTx(context.Background(), &opts)

		// Oops, begin TX failed
		if err != nil {
			return err
		}

		// Execute the operation
		err = function(tx)

		// Oops, the operation has failed
		if err != nil {
			// Trying to rollback
			rbError := tx.Rollback()

			// Oops, rollback has failed. Just log it
			if rbError != nil {
				logger := GetLogger(ctx)
				logger.Error().Msgf("Error while rollin back transaction %w", rbError)
			}

			// Return original issue produced by the operation
			return err

		}

		// Trying to commit
		return tx.Commit()
	})
}