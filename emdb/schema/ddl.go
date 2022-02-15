package schema

type Reader interface {
	Read(buffer []byte) (string, error)
}

type Sequence struct {
	Id uint16
	Name string
}

type PrimaryKey struct {
	Id uint16
	Name string
	Seq Sequence
}

type Index struct {
	Id uint16
	Name string
	Reader Reader
	Unique bool
}

type ForeignKey struct {
	Id uint16
	Name string
	Ref Collection
}

type Collection struct {
	Name string
	PrimaryKey PrimaryKey
	Indexes []Index
	ForeignKeys []ForeignKey
	Reader Reader
}


