package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/iyarkov2/chat/core/lock"
	"github.com/iyarkov2/chat/core/util"
	_ "github.com/lib/pq"
	"time"
)

// Localhost
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	dbname   = "postgres"
)

var (
	lockService lock.Service
	db *sql.DB
)

func main() {
	setUp()
	defer tearDown()
	smoke()

}

func setUp() {
	ctx := util.WithLogger(context.Background(), map[string]string {"step" : "setUp"})
	log := util.GetLogger(ctx)

	// Open DB connection
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	var err error
	db, err = sql.Open("postgres", psqlconn)
	util.PanicIfError(err)
	util.PanicIfError(db.Ping())
	log.Info().Msg("DB Connected")

	// Clean up tables
	err = util.WithConnection(ctx, db, func(conn *sql.Conn) error {
		cleanupTable(ctx,"request_record", conn)
		cleanupTable(ctx,"resource", conn)
		return nil
	})
	util.PanicIfError(err)

	configuration := lock.Configuration {
		LockTimeout : 3 * time.Second,
		LockAttempts : 3,
		LockDuration : 30 * time.Second,
		TableName : "lock",

	}
	lockService, err = lock.New(context.Background(), db, configuration)
	util.PanicIfError(err)

	log.Info().Msg("Completed")
}

func cleanupTable(ctx context.Context, table string, conn *sql.Conn) {
	log := util.GetLogger(ctx)
	result, err := conn.ExecContext(context.Background(), "delete from " + table)
	util.PanicIfError(err)
	rowsAffected, err := result.RowsAffected()
	util.PanicIfError(err)
	log.Info().Msgf("Table %s cleaned, %d records deleted", table, rowsAffected)
}

func tearDown() {
	ctx := util.WithLogger(context.Background(), map[string]string {"step" : "tear down"})
	log := util.GetLogger(ctx)

	util.CloseQuiet(ctx,  "database", db)

	log.Info().Msg("Completed")

}

func smoke() {
	ctx := util.WithLogger(context.Background(), map[string]string {"step" : "smoke test"})
	log := util.GetLogger(ctx)

	lock, err := lockService.Get(ctx,"Smoke")
	log.Info().Msg("Lock obtained")
	util.PanicIfError(err)

	err = lockService.Release(ctx, lock)
	util.PanicIfError(err)
	log.Info().Msg("Lock released")
}