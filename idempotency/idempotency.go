package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/rs/zerolog"
	"time"
)

type RequestRecord struct {
	Id string
	Result []byte

	createdAt time.Time
	updatedAt time.Time
	version uint
	lockedUntil time.Time
}

type EmbeddedServiceConfig struct {
	TableName string
	RetentionPeriodSec uint
}

type StandaloneServiceConfig struct {
	EmbeddedServiceConfig
	LockWaitTimeout time.Duration
	LockTimeout time.Duration
}


type EmbeddedService interface {

	Get(ctx context.Context, tx *sql.Tx, id string) (*RequestRecord, error)

	Set(ctx context.Context, tx *sql.Tx, record *RequestRecord) error
}

type StandaloneService interface {

	Get(ctx context.Context, id string) (*RequestRecord, error)

	Set(ctx context.Context, record *RequestRecord) error

	Release(ctx context.Context, record *RequestRecord) error
}

func NewEmbeddedService(ctx context.Context, db *sql.DB, config EmbeddedServiceConfig) (EmbeddedService, error) {
	s := embeddedService {
		config: config,
		insertStmt: fmt.Sprintf(`insert into %s (id, created_at, updated_at) values($1, $2, $3)`, config.TableName),
		updateStmt: fmt.Sprintf(`update %s set result = $2, updated_at = $3 where id=$1`, config.TableName),
		selectStmt: fmt.Sprintf(`select id, result, created_at, updated_at from %s where id = $1`, config.TableName),
	}
	//TODO - check that the table exists, start retention timer, etc
	return &s, nil
}

func NewStandaloneService(ctx context.Context, db *sql.DB, config StandaloneServiceConfig) (StandaloneService, error) {
	if config.LockWaitTimeout <= 0 {
		return nil, errors.New("lock wait timeout must be positive")
	}
	if config.LockTimeout <= 0 {
		return nil, errors.New("lock timeout must be positive")
	}
	s := standaloneService{
		config: config,
		db: db,
		insertStmt: fmt.Sprintf("insert into %s (id, created_at, updated_at, version, locked_until) values($1, $2, $3, $4, $5)", config.TableName),
		selectStmt: fmt.Sprintf("select id, result, created_at, updated_at, version, locked_until from %s where id = $1 for update", config.TableName),
		// TODO - setStmt / unlockStmt
		//setStmt:    fmt.Sprintf("update %s set result = $2, updated_at = $3 where id=$1", config.TableName),
		//unlockStmt: fmt.Sprintf("update %s set result = $2, updated_at = $3 where id=$1", config.TableName),
	}
	//TODO - check that the table exists, clean up timer
	return &s, nil
}

type embeddedService struct {
	config EmbeddedServiceConfig

	insertStmt string
	selectStmt string
	updateStmt string
}

func (s *embeddedService) Get(ctx context.Context, tx *sql.Tx, id string) (*RequestRecord, error) {
	log := getLogger(ctx)
	// Create TX safe point
	_, e := tx.Exec("savepoint idempotency_safepoint")
	if e != nil {
		return nil, e
	}

	now := time.Now()
	record := RequestRecord {
		Id: id,
		createdAt: now,
		updatedAt: now,
		Result: nil,
	}

	_, e = tx.Exec(s.insertStmt, record.Id, record.createdAt, record.updatedAt)
	if e == nil {
		log.Debug().Msgf("New Record Inserted %s", record.Id)
	} else {
		if pqerror, ok := e.(*pq.Error); ok {
			log.Debug().Msgf("code: %s, table: %s", pqerror.Code, pqerror.Table)
			if pqerror.Code.Name() != "unique_violation" || pqerror.Table != s.config.TableName {
				return nil, e
			}
			log.Debug().Msgf("Record already exist %s", id)
			_, e := tx.Exec("rollback to idempotency_safepoint")
			if e != nil {
				return nil, e
			}

			// TODO - does not work with LevelSerializable, fails with "could not serialize access due to read/write dependencies among transactions"
			// Not sure if anyone is using anything but LevelReadCommitted. If so - a new transaction must be started to read the data from the db
			rows, e := tx.QueryContext(context.Background(), s.selectStmt, record.Id)
			defer func() {
				if rows != nil {
					e := rows.Close()
					if e != nil {
						log.Error().Msgf("Error on rows close, %s", e)
					}
				}
			}()
			if e != nil {
				return nil, fmt.Errorf("failed to select a record, %s", e)
			}

			if rows.Next() {
				e = rows.Scan(&record.Id, &record.Result, &record.createdAt, &record.updatedAt)
				if e != nil {
					return nil, fmt.Errorf("failed to process a record, %s", e)
				}
				log.Debug().Msgf("Found exist record, %s ->[%s], created at: %s, updated at: %s", record.Id, string(record.Result), record.createdAt, record.updatedAt)
				return &record, nil
			} else {
				return nil, fmt.Errorf("select yeld no result")
			}
		} else {
			return nil, fmt.Errorf("failed to insert new record, %s", e)
		}
	}
	return &record, nil
}

func (s *embeddedService) Set(ctx context.Context, tx *sql.Tx, record *RequestRecord) error {
	log := getLogger(ctx)
	record.updatedAt = time.Now()
	result, e := tx.Exec(s.updateStmt, record.Id, record.Result, record.updatedAt)
	if e == nil {
		rowsAffected, e := result.RowsAffected()
		if e != nil {
			log.Error().Msgf("Failed to get affected rows %s", e)
		} else if rowsAffected != 1 {
			log.Error().Msgf("record set failed, expected rowsAffected:1, actual rowsAffected: %d", rowsAffected)
		} else {
			log.Debug().Msgf("record updated")
		}
	}
	return e
}

type standaloneService struct {
	config StandaloneServiceConfig

	db *sql.DB

	insertStmt string
	selectStmt string
	setStmt    string
	unlockStmt string
}

func (s *standaloneService) Get(ctx context.Context, id string) (*RequestRecord, error) {
	log := getLogger(ctx)

	// Open new connection
	conn, e := s.db.Conn(context.Background())
	defer func() {
		e = conn.Close()
		if e != nil {
			log.Error().Msgf("failed to close a connection %s", e)
		}
	}()
	if e != nil {
		return nil, fmt.Errorf("can not open connection %s", e)
	}
	log.Debug().Msg("Idempotency connection opened")

	// Insert new record
	now := time.Now()
	record := RequestRecord {
		Id: id,
		Result: nil,

		createdAt: now,
		updatedAt: now,
		lockedUntil: now.Add(s.config.LockTimeout),
		version: 1,
	}
	e = doInTx(conn, func(tx *sql.Tx) error {
		log.Debug().Msg("Idempotency transaction Started")
		_, e = tx.Exec(s.insertStmt, record.Id, record.createdAt, record.updatedAt, record.version, record.lockedUntil)
		return e
	})

	// No error - we got the lock!
	if e == nil {
		log.Debug().Msgf("New Record Inserted %s", record.Id)
		return &record, nil
	}

	// Not a PQ error - return it
	pqerror, ok := e.(*pq.Error)
	if !ok {
		return nil, e
	}

	// PQ error - check it
	log.Debug().Msgf("code: %s, table: %s", pqerror.Code, pqerror.Table)
	if pqerror.Code.Name() != "unique_violation" || pqerror.Table != s.config.TableName {
		// Not a constraint violation - return it
		return nil, e
	}

	// Row exist - now let try to obtain an optimistic lock
	deadline := time.Now().Add(s.config.LockWaitTimeout)
	for deadline.Before(time.Now()) {
		e = doInTx(conn, func(tx *sql.Tx) error {
			log.Debug().Msg("Read from the DB")
			rows, e := tx.QueryContext(context.Background(), s.selectStmt, record.Id)
			defer func() {
				if rows != nil {
					e := rows.Close()
					if e != nil {
						log.Error().Msgf("Error on rows close, %s", e)
					}
				}
			}()
			if e != nil {
				return fmt.Errorf("failed to select a record, %s", e)
			}

			if rows.Next() {
				// select id, result, created_at, updated_at, version, locked_until from %s where id = $1
				e = rows.Scan(&record.Id, &record.Result, &record.createdAt, &record.updatedAt, &record.version, &record.lockedUntil)
				if e != nil {
					return nil, fmt.Errorf("failed to process a record, %s", e)
				}
				log.Debug().Msgf("Found exist record, %s ->[%s], created at: %s, updated at: %s", record.Id, string(record.Result), record.createdAt, record.updatedAt)
				return &record, nil
			} else {
				/*
					This is an odd situation. We just got a unique key constraint violation, now the record is not found
					The record was either deleted by the timer or manualy
				 */
				return fmt.Errorf("request record not found")
			}
		})
	}
}

func (s *standaloneService) Set(ctx context.Context, record *RequestRecord) error {
	return errors.New("not implemented")
}

func (s *standaloneService) Release(ctx context.Context, record *RequestRecord) error {
	return errors.New("not implemented")
}

//TODO - do not copy, use utility from gokit
func getLogger(ctx context.Context) zerolog.Logger {
	val := ctx.Value("log")
	if val == nil {
		panic("No logger in context")
	}

	if log, ok := val.(zerolog.Logger); ok {
		return log
	}
	panic("Invalid logger in context")
}

func doInTx(conn *sql.Conn, action func (tx *sql.Tx) error) error {
	// Start a transaction - note that we are using ReadCommitted. This is important!
	opts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly: false,
	}
	tx, e := conn.BeginTx(context.Background(), &opts)
	defer func() {

	}()
	if e != nil {
		return fmt.Errorf("can not begin a transaction %s", e)
	}

	e = action(tx)
	if e != nil {
		ex := tx.Rollback()
		if ex != nil {
			// Rollback failed - most likely some serious issue - DB crashed, network issue, etc
			return fmt.Errorf("tx rollbacl failed %s, operation failed %s", ex, e)
		}
		// Do not modify the error - return it as is
		return e
	}

	e = tx.Commit()
	if e != nil {
		return fmt.Errorf("tx commit failed %s", e)
	}

	return nil
}