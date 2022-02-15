package main

import (
	"context"
	"database/sql"
	"fmt"
	i "github.com/iyarkov2/chat/idempotency"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"os"
	"strconv"
	"time"
)

// Localhost
const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	dbname   = "postgres"
)

var db *sql.DB
var embeddedService i.EmbeddedService
var standaloneService i.StandaloneService
var log zerolog.Logger


func main() {
	setUp()
	defer tearDown()

	sequentialCalls()
	parallelCalls()
	multiParallelCalls()
	multiParallelCallsWithFail()
}

func setUp() {
	log = zerolog.New(zerolog.ConsoleWriter {Out: os.Stdout, TimeFormat: "2006-01-02T15:04:0543"}).With().Timestamp().Logger()
	log.Info().Msg("Init")
	// connection string
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	var err error
	// open database
	db, err = sql.Open("postgres", psqlconn)
	checkError(err)

	// check db
	err = db.Ping()
	checkError(err)
	log.Info().Msg("Connected")

	// Clean up tables
	conn, err := db.Conn(context.Background())
	cleanup("request_record", conn, log)
	cleanup("resource", conn, log)
	err = conn.Close()
	checkError(err)

	// Create EmbeddedService
	embeddedServiceConfig := i.EmbeddedServiceConfig {
		TableName: "request_record",
		RetentionPeriodSec: 3600,
	}
	embeddedService, err = i.NewEmbeddedService(context.Background(), db, embeddedServiceConfig)
	checkError(err)

	// Create StandaloneService
	standaloneServiceConfig := i.StandaloneServiceConfig {
		EmbeddedServiceConfig : i.EmbeddedServiceConfig {
			TableName: "request_record",
			RetentionPeriodSec: 3600,
		},
		LockWaitTimeout: 5 * time.Second,
		LockTimeout: 30 * time.Second,
	}
	standaloneService, err = i.NewStandaloneService(context.Background(), db, standaloneServiceConfig)
	checkError(err)

	log.Info().Msg("Init completed")
}

func cleanup(table string, conn *sql.Conn, log zerolog.Logger) {
	result, err := conn.ExecContext(context.Background(), "delete from " + table)
	checkError(err)
	rowsAffected, err := result.RowsAffected()
	checkError(err)
	log.Info().Msgf("Table %s cleaned, %d records deleted\n", table, rowsAffected)
}

func tearDown() {
	if db != nil {
		log.Info().Msg("Closing the DB")
		e := db.Close()
		if e != nil {
			log.Info().Msgf("Error while closing DB %s\n", e)
		}
	}
	log.Info().Msg("Resources released")
}

// A mock operation, does not do anything, it waits for a second to emulate a long operations and also inserts a record into the DB
func operation(requestId string, goroutineId string, withFail bool) {
	//operationEmbedded(requestId, goroutineId, withFail)
	operationStandalone(requestId, goroutineId, withFail)
}

func operationEmbedded(requestId string, goroutineId string, withFail bool) {
	requestLog := zerolog.New(zerolog.ConsoleWriter {Out: os.Stdout, TimeFormat: "2006-01-02T15:04:0543"}).With().Timestamp().Logger().With().Str("id", goroutineId).Logger()
	ctx := context.WithValue(context.Background(), "log", requestLog)

	// Use Service.
	// Open new connection and begin a transaction
	conn, err := db.Conn(context.Background())
	checkError(err)
	requestLog.Info().Msg("New connection opened")

	// Start a transaction - experimented with both ReadCommitted and Serializable, current impl does not work for Serializable
	opts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		// Isolation: sql.LevelSerializable,
		ReadOnly: false,
	}
	tx, err := conn.BeginTx(context.Background(), &opts)
	checkError(err)
	requestLog.Info().Msg("Transaction Started")

	// Get the record
	record, err := embeddedService.Get(ctx, tx, requestId)
	checkError(err)

	if record.Result == nil {
		requestLog.Info().Msg("New record, perform operation")
		// Simulate some work
		time.Sleep(1 * time.Second)

		// Insert a record
		something := fmt.Sprintf("Request %s, goroutine %s, withFail %t", requestId, goroutineId, withFail)
		_, err = tx.Exec("insert into resource(something, created_at) values($1, $2)", something, time.Now())
		checkError(err)

		result := fmt.Sprintf("Time: %s", time.Now())
		record.Result = []byte(result)
		err = embeddedService.Set(ctx, tx, record)
		checkError(err)
		requestLog.Info().Msgf("Done, result: %s", result)

		if withFail {
			err = tx.Rollback()
			checkError(err)
			requestLog.Info().Msg("Transaction Rolled Back")
		} else {
			err = tx.Commit()
			checkError(err)
			requestLog.Info().Msg("Transaction Committed")
		}
	} else {
		result := string(record.Result)
		requestLog.Info().Msgf("Already exists, use result: %s", result)
		err = tx.Commit()
		checkError(err)
		requestLog.Info().Msg("Transaction Committed")
	}
}

func operationStandalone(requestId string, goroutineId string, withFail bool) {
	requestLog := zerolog.New(zerolog.ConsoleWriter {Out: os.Stdout, TimeFormat: "2006-01-02T15:04:0543"}).With().Timestamp().Logger().With().Str("id", goroutineId).Logger()
	ctx := context.WithValue(context.Background(), "log", requestLog)

	// Get the record
	record, err := standaloneService.Get(ctx, requestId)
	checkError(err)

	if record.Result == nil {
		requestLog.Info().Msg("New record, perform operation")
		// Open new connection and begin a transaction
		conn, err := db.Conn(context.Background())
		checkError(err)
		requestLog.Info().Msg("New connection opened")

		// Start a transaction - experimented with both ReadCommitted and Serializable
		opts := sql.TxOptions{
			Isolation: sql.LevelReadCommitted,
			// Isolation: sql.LevelSerializable,
			ReadOnly: false,
			// Question - is there a way to specify a timeout???
		}
		tx, err := conn.BeginTx(context.Background(), &opts)
		checkError(err)
		requestLog.Info().Msg("Transaction Started")

		// Simulate some work
		time.Sleep(1 * time.Second)

		// Insert a record
		something := fmt.Sprintf("Request %s, goroutine %s, withFail %t", requestId, goroutineId, withFail)
		_, err = tx.Exec("insert into resource(something, created_at) values($1, $2)", something, time.Now())
		checkError(err)

		// Calculation result
		result := fmt.Sprintf("Time: %s", time.Now())

		if withFail {
			err = tx.Rollback()
			checkError(err)
			requestLog.Info().Msg("Transaction Rolled Back")

			// it failed - so release the lock without saving anything
			err = standaloneService.Release(ctx, record)
			checkError(err)
		} else {
			err = tx.Commit()
			checkError(err)
			requestLog.Info().Msgf("Transaction Committed, result: %s", result)

			// it succeeded - Set, it also will release the lock
			record.Result = []byte(result)
			err = standaloneService.Set(ctx, record)
			checkError(err)
		}
	} else {
		result := string(record.Result)
		requestLog.Info().Msgf("Already exists, use result: %s", result)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}


func sequentialCalls() {
	log.Info().Msgf("-------------------- sequentialCalls ---------------------")
	operation("0001", "T1", false)
	log.Info().Msgf("")
	operation("0001", "T1", false)
	log.Info().Msgf("-------------------------- done --------------------------")
}

func parallelCalls() {
	log.Info().Msg("-------------------- parallelCalls ---------------------")
	ch := make(chan int)
	go func() {
		operation("0002", "T1", false)
		ch <- 1
	}()
	go func() {
		operation("0002", "T2", false)
		ch <- 2
	}()

	log.Info().Msgf("Goroutine T%d Done", <- ch)
	log.Info().Msgf("Goroutine T%d Done", <- ch)
	log.Info().Msg("-------------------------- done --------------------------")
}

func multiParallelCalls() {
	log.Info().Msg("-------------------- parallelCalls ---------------------")
	ch := make(chan int)
	const count = 6
	for i := 0; i < count; i++ {
		tid := i
		go func() {
			if tid % 2 == 0 {
				operation("0003", "T" + strconv.Itoa(tid), false)
			} else {
				operation("0004", "T" + strconv.Itoa(tid), false)
			}
			ch <- tid
		}()
	}

	for i := 0; i < count; i++ {
		log.Info().Msgf("Goroutine T%d Done", <- ch)
	}
	log.Info().Msg("-------------------------- done --------------------------")
}

func multiParallelCallsWithFail() {
	log.Info().Msg("-------------------- multiParallelCallsWithFail ---------------------")
	ch := make(chan int)
	const count = 6
	for i := 0; i < count; i++ {
		tid := i
		go func() {
			if tid % 2 == 0 {
				operation("0005", "T" + strconv.Itoa(tid), true)
			} else {
				operation("0006", "T" + strconv.Itoa(tid), false)
			}
			ch <- tid
		}()
	}

	for i := 0; i < count; i++ {
		log.Info().Msgf("Goroutine T%d Done", <- ch)
	}
	log.Info().Msg("-------------------------- done --------------------------")
}


