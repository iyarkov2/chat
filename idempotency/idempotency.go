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
	LockWaitTimeout uint
	LockTimeout uint
}


type EmbeddedService interface {

	Get(ctx context.Context, tx *sql.Tx, id string) (*RequestRecord, error)

	Set(ctx context.Context, tx *sql.Tx, record *RequestRecord) error
}

type StandaloneService interface {

	Get(ctx context.Context, id string) (RequestRecord, error)

	Set(ctx context.Context, record RequestRecord) error

	Release(ctx context.Context, record RequestRecord) error
}

func NewEmbeddedService(ctx context.Context, conn *sql.DB, config EmbeddedServiceConfig) (EmbeddedService, error) {
	s := embeddedService {
		config: config,
		insertStmt: fmt.Sprintf(`insert into %s (id, created_at, updated_at) values($1, $2, $3)`, config.TableName),
		updateStmt: fmt.Sprintf(`update %s set result = $2, updated_at = $3 where id=$1`, config.TableName),
		selectStmt: fmt.Sprintf(`select id, result, created_at, updated_at from %s where id = $1`, config.TableName),
	}
	//TODO - init logs, check that the table exists
	return &s, nil
}

func NewStandaloneService(ctx context.Context, conn *sql.DB, config StandaloneServiceConfig) (StandaloneService, error) {
	return nil, errors.New("not implemented yet")
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