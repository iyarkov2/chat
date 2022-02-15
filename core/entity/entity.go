package entity

import (
	"github.com/google/uuid"
	"time"
)

type Entity struct {
	id uuid.UUID
	createdAt time.Time
	updatedAt time.Time
}

type Versioned struct {
	Version int
}
