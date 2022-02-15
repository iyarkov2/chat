package outbox

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/iyarkov2/chat/core/util"
	"sync"
	"time"
)

/*
	Implementation of transactional outbox pattern
 */

type Status int8

const (
	Pending Status = iota
	Completed
	Failed
)

type Task struct {
	ID uuid.UUID
	Version int
	Type string
	Data []byte
	ExecCounter int
	Status Status
	CreateAt time.Time
	UpdatedAt time.Time

	Hydrated bool
	Metadata interface{}
}

/*
	Takes the stored data and creates a context by loading all necessarily data from the DB or other services
 */
type Hydrator interface {
	Hydrate(ctx context.Context, Data []byte) (interface{}, error)
	HydrateBulk(ctx context.Context, Data []byte) ([]interface{}, error)
}

/*
	Generates the store data from a context
*/
type Dehydrator interface {
	Dehydrate(ctx context.Context, metadata interface{}) ([]byte,error)
}

type Worker interface {
	Do(ctx context.Context, task Task) error
}

type Config struct {
	Table string
}

type Service struct {
	config Config
	registry map[string]registrationRecord
	mtx *sync.Mutex

	insertStmt string
}

type registrationRecord struct {
	hydrator Hydrator
	dehydrator Dehydrator
	worker Worker
}

func NewService(ctx context.Context, db *sql.DB, config Config) (Service, error) {
	// TODO - validate

	result := Service {
		registry: make(map[string]registrationRecord),
		mtx: new(sync.Mutex),

		insertStmt: fmt.Sprintf("insert into "),
	}

	return result, nil
}

func (s Service) Register(ctx context.Context, name string, hydrator Hydrator, dehydrator Dehydrator, worker Worker) bool {
	log := util.GetLogger(ctx)
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if _, ok := s.registry[name]; ok {
		log.Error().Msgf("Task type %s already registered", name)
		return false
	} else {
		s.registry[name] = registrationRecord {
			hydrator: hydrator,
			dehydrator: dehydrator,
			worker: worker,
		}
		log.Info().Msgf("Task type %s registered", name)
		return true
	}
}

func (s Service) get(name string) *registrationRecord {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	result, ok := s.registry[name]
	if ok {
		return &result
	} else {
		return nil
	}
}


//func (s Service) Create(ctx context.Context, tx *sql.Tx, taskType string, Metadata interface{}) (Task, error) {
//	registration := s.get(taskType)
//	if registration == nil {
//		return Task{}, fmt.Errorf("unknown task type %s", taskType)
//	}
//
//	task := Task {
//		ID: uuid.New(),
//		Version: 1,
//		Type: taskType,
//		ExecCounter: 0,
//		Status: Pending,
//		CreateAt: time.Now(),
//		UpdatedAt: time.Now(),
//	}
//	//_, err := tx.Exec(s.insertStmt, task.ID, record.createdAt, record.updatedAt)
//}

func (s Service) Execute(task Task) error {
	return errors.New("not implemented")
}