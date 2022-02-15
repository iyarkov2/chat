package lock

import (
	"database/sql"
	"fmt"
	"github.com/iyarkov2/chat/core/util"
	"time"
)

/*
	Experimenting with DB locks
 */
import (
	"context"
	"errors"
)

const (
	NotImplemented = "not implemented"
	//X = util.Error{}
)

type Lock struct {
	name string
	expiresAt time.Time
	version int
	obtained bool
}

type Configuration struct {
	LockTimeout time.Duration
	LockAttempts uint
	LockDuration time.Duration
	TableName string
}

func (configuration Configuration) validate() error {
	validation := make([]string, 0)
	if configuration.LockTimeout <= 0 {
		validation = append(validation, "lock timeouts must be positive")
	}
	if configuration.LockAttempts <= 0 {
		validation = append(validation, "lock attempts must be positive")
	}
	if configuration.LockDuration <= 0 {
		validation = append(validation, "lock duration must be positive")
	}
	if configuration.TableName == "" {
		validation = append(validation, "table name required")
	}
	if len(validation) > 0 {
		return fmt.Errorf("invalid configuration %v", validation)
	}
	return nil
}

type Service struct {
	config Configuration

	statementSelect string
	statementInsert string
	statementUpdate string
	statementDelete string

	db *sql.DB
}

func New(ctx context.Context, db *sql.DB, configuration Configuration) (Service, error) {
	if err := configuration.validate(); err != nil {
		return Service{}, err
	}
	if db == nil {
		return Service{}, errors.New("DB must not be nil")
	}

	if err := util.WithConnection(ctx, db, func(conn *sql.Conn) error {
		// TODO - very simple check that the table exists and is valid
		// Need an improvement. Read from the DB schema, check column types, indexes
		statementSelect := fmt.Sprintf("SELECT name, version, expires_at FROM %s LIMIT 1", configuration.TableName)
		if rows, e := conn.QueryContext(context.Background(), statementSelect); e != nil {
			return fmt.Errorf("DB check failed %w", e)
		} else {
			util.CloseQuiet(ctx, "rows", rows)
		}

		// All checks passed
		return nil
	}); err != nil {
		//Check failed - return error
		return Service{}, err
	}

	logger := util.GetLogger(ctx)
	logger.Info().Msg("Lock service started")
	return Service {
		config: configuration,
		statementInsert: fmt.Sprintf("INSERT INTO %s(name, version, expires_at) values (?, ?, ?)", configuration.TableName),
		db: db,
	}, nil
}

func (s Service) Get(ctx context.Context, name string) (Lock, error) {
	return Lock{}, errors.New(NotImplemented)
	//result := Lock {
	//	name: name,
	//}
	//err := util.WithTx(ctx, s.db, func (tx *sql.Tx) error {
	//	// Initialize new lock record
	//	result.version = 1
	//	result.expiresAt = time.Now().Add(s.config.LockDuration)
	//
	//	// Try to insert
	//	ok, e := s.insert(ctx, tx, result)
	//	if e != nil {
	//		return fmt.Errorf("insert failed %w", e)
	//	}
	//	result.obtained = ok
	//	return nil
	//})
	//
	//return result, err
	//if err != nil {
	//	err = fmt.Errorf("failed to obtain lock, %w", err)
	//}
	//return result, err
	//
	//if
	//
	//now := time.Now()
	//record := RequestRecord {
	//	Id: id,
	//	createdAt: now,
	//	updatedAt: now,
	//	Result: nil,
	//}
	//
	//_, e = tx.Exec(s.insertStmt, record.Id, record.createdAt, record.updatedAt)
	//if e == nil {
	//	log.Debug().Msgf("New Record Inserted %s", record.Id)
	//} else {
	//	if pqerror, ok := e.(*pq.Error); ok {
	//		log.Debug().Msgf("code: %s, table: %s", pqerror.Code, pqerror.Table)
	//		if pqerror.Code.Name() != "unique_violation" || pqerror.Table != s.config.TableName {
	//			return nil, e
	//		}
	//		log.Debug().Msgf("Record already exist %s", id)
	//		_, e := tx.Exec("rollback to idempotency_safepoint")
	//		if e != nil {
	//			return nil, e
	//		}
	//
	//		// TODO - does not work with LevelSerializable, fails with "could not serialize access due to read/write dependencies among transactions"
	//		// Not sure if anyone is using anything but LevelReadCommitted. If so - a new transaction must be started to read the data from the db
	//		rows, e := tx.QueryContext(context.Background(), s.selectStmt, record.Id)
	//		defer func() {
	//			if rows != nil {
	//				e := rows.Close()
	//				if e != nil {
	//					log.Error().Msgf("Error on rows close, %s", e)
	//				}
	//			}
	//		}()
	//		if e != nil {
	//			return nil, fmt.Errorf("failed to select a record, %s", e)
	//		}
	//
	//		if rows.Next() {
	//			e = rows.Scan(&record.Id, &record.Result, &record.createdAt, &record.updatedAt)
	//			if e != nil {
	//				return nil, fmt.Errorf("failed to process a record, %s", e)
	//			}
	//			log.Debug().Msgf("Found exist record, %s ->[%s], created at: %s, updated at: %s", record.Id, string(record.Result), record.createdAt, record.updatedAt)
	//			return &record, nil
	//		} else {
	//			return nil, fmt.Errorf("select yeld no result")
	//		}
	//	} else {
	//		return nil, fmt.Errorf("failed to insert new record, %s", e)
	//	}
	//}
	//return &record, nil
}

func (s Service) Release(ctx context.Context, lock Lock) error {
	return errors.New(NotImplemented)
}

//func (s Service) insert(ctx context.Context, tx *sql.Tx, lock Lock) (bool, error) {
//	// Create TX safe point
//	if _, err := tx.Exec("savepoint p1"); err != nil {
//		return false, fmt.Errorf("failed to create safepoint, %w", err)
//	}
//
//	_, insertError := tx.Exec(s.statementInsert, lock.name, lock.version, lock.expiresAt)
//	if insertError == nil {
//		return true, nil
//	}
//
//
//}