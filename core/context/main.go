package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/iyarkov2/chat/core/util"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	dbname   = "experiment"
)

func main() {
	ctx := util.WithLogger(context.Background(), map[string]string {})
	log := util.GetLogger(ctx)

	// Open DB connection
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)
	var err error
	db, err := sql.Open("postgres", psqlconn)
	defer util.CloseQuiet(ctx, "DB", db)
	util.PanicIfError(err)
	util.PanicIfError(db.Ping())
	log.Info().Msg("DB Connected")

	// Open new connection
	conn, err := db.Conn(context.Background())
	util.PanicIfError(err)
	log.Info().Msg("New connection opened")

	// Start transaction
	opts := sql.TxOptions {
		Isolation: sql.LevelReadCommitted,
		// Isolation: sql.LevelSerializable,
		ReadOnly: false,
	}
	tx, err := conn.BeginTx(ctx, &opts)
	util.PanicIfError(err)
	log.Info().Msg("Transaction Started")

	// Insert first record
	id := uuid.New()
	_, err = tx.Exec("insert into outbox(id, message) values ($1, $2)", id, "First")
	util.PanicIfError(err)
	log.Info().Msg("First Record Inserted")


	// Insert second record
	_, err = tx.Exec("insert into outbox(id, message) values ($1, $2)", id, "Second")
	util.PanicIfError(err)
	log.Info().Msg("First Record Inserted")
	util.PanicIfError(tx.Commit())

	//id         | uuid                        |           | not null |
	//	version    | integer                     |           |          |
	//	created_at | timestamp without time zone |           |          |
	//	updated_at | timestamp without time zone |           |          |
	//	message
}
