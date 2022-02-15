/*
	Purpose:

	1. Configuration validation
	2. Building tools

 */
package schema

import (
	"fmt"
	"sync"
)

type Type uint8

const (
	typeSize = 2

	ValueId Type = iota
	PrimaryKeyId
	IndexId
	ForeignKeyId
	SequenceId
)

func (t Type ) String() string {
	switch t {
	//case Value: return "Value"
	//case PrimaryKey: return "Primary Key"
	//case Index: return "Index"
	//case ForeignKey: return "Foreign Key"
	//case Sequence: return "Sequence"
	default: return "Unknown"
	}
}


type record struct {
	Id uint16
	GetType func() Type
	Name string // Must be unique
	Reader func(data []byte) (string, error)
}

//
//
//func (pk PrimaryKey) Read(data []byte) (string, error) {
//	var id uint16
//	size := binary.Size(id) + typeSize
//	if len(data) != size {
//		return "", fmt.Errorf("invalid ID buffer size. expecting %d, actual %d", size, len(data))
//	}
//	if data[0] != PrimaryKeyId {
//		return "", fmt.Errorf("invalid key type, expecting %d, actual %d", pk, data[0])
//	}
//	return strconv.FormatUint(binary.BigEndian.Uint64(data[typeSize:]), 10), nil
//}
//
//func (r record) String() string {
//	return fmt.Sprintf("Id: %d, Type: %s, Name: %s", r.Id, r.GetType().String(), r.Name)
//}
//
//var (
//	ErrorIdRequired = fmt.Errorf("id required")
//	ErrorNameRequired = fmt.Errorf("name required")
//	ErrorInvalidType = fmt.Errorf("invalid type")
//	ErrorReaderRequired = fmt.Errorf("reader required")
//	ErrorIdAlreadyRegistered = fmt.Errorf("record with this ID already registered")
//	ErrorNameAlreadyRegistered = fmt.Errorf("record with this name already registered")
//)
//
type Schema struct {
	mtx sync.Mutex
	byId map[uint16]record
	byName map[string]*record
}

func NewSchema(collections []Collection, sequences []Sequence) Schema {
	fmt.Printf("Users sequence 3 %s\n", collections[0].PrimaryKey.Seq.Name)
	fmt.Printf("Users seq 3 %s\n", sequences[0].Name)
	sequences[0].Name = "ABC"

	fmt.Printf("Users sequence 4 %s\n", collections[0].PrimaryKey.Seq.Name)
	fmt.Printf("Users seq 4 %s\n", sequences[0].Name)

	return Schema {
		mtx: sync.Mutex{},
		byId: make(map[uint16]record),
		byName: make(map[string]*record),
	}
}
//
//func (r *Schema) Register(record record) error {
//	if record.Id == 0 {
//		return ErrorIdRequired
//	}
//
//	if record.Name == "" {
//		return ErrorNameRequired
//	}
//
//	//switch record.Type {
//	//case Value:
//	//case PrimaryKey:
//	//case Index:
//	//case ForeignKey:
//	//case Sequence:
//	//default: return ErrorInvalidType
//	//}
//
//	if record.Reader == nil {
//		return ErrorReaderRequired
//	}
//
//	r.mtx.Lock()
//	defer r.mtx.Unlock()
//
//	if _, ok := r.byId[record.Id]; ok {
//		return ErrorIdAlreadyRegistered
//	}
//
//	if _, ok := r.byName[record.Name]; ok {
//		return ErrorNameAlreadyRegistered
//	}
//
//	r.byId[record.Id] = record
//	r.byName[record.Name] = &record
//
//	return nil
//}

//func (r *Schema) FindById(id uint16) (record Record, ok bool) {
//	r.mtx.Lock()
//	defer r.mtx.Unlock()
//	record, ok = r.byId[id]
//	return
//}
//
//func (r *Schema) FindByName(name string) (Record, bool) {
//	r.mtx.Lock()
//	defer r.mtx.Unlock()
//	recordRef, ok := r.byName[name]
//	return *recordRef, ok
//}
//
//func (r *Schema) Dump() {
//	for _, v := range r.byId {
//		fmt.Println(v)
//	}
//}
