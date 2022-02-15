package main

import (
	"fmt"
	"github.com/dgraph-io/badger"
	"math/rand"
	"time"
)

const (
	dataDir = "data"

)

func main() {
	opts := badger.DefaultOptions("data")
	opts.SyncWrites = true

	db, err := badger.Open(opts)
	if err != nil {
		panic(fmt.Errorf("failed to open db, %w", err))
	}
	defer silentClose(db)

	//drop(db)
	//write(db)
	//readById(db, 3)
	readAll(db)
	//readByAge(db, 3)

	fmt.Println("DB Created")
}

func silentClose(db closable) {
	if err := db.Close(); err != nil {
		fmt.Printf("Error while closing db %s\n", err)
	} else {
		fmt.Println("Closed")
	}
}

type closable interface {
	Close() error
}



func drop(db *badger.DB) {
	if err := db.DropAll(); err != nil {
		panic(fmt.Errorf("db drop failed %w", err))
	}
	fmt.Println("All records dropped")
}

func write(db *badger.DB) {
	var idBuffer []byte
	var value Value
	var valueBuffer []byte
	var ageIdxBuffer []byte
	idxValue := []byte{}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	sequence, seqErr := db.GetSequence([]byte{ sequence }, 10)
	if seqErr != nil {
		panic(fmt.Errorf("get sequence failed %w", seqErr))
	}
	defer func() {
		if err := sequence.Release(); err != nil {
			fmt.Printf("error while releasing a sequence")
		}
	}()

	for i := 0; i < 100; i++ {
		err := db.Update(func(tx *badger.Txn) error {
			next, seqNextErr := sequence.Next()
			if seqNextErr != nil {
				panic(fmt.Errorf("sequence next failed %w", seqNextErr))
			}

			value.ID = next
			value.Name = fmt.Sprintf("John Smith %d", i)
			value.Age = uint8(random.Int31n(10))

			idBuffer = WritePk(idBuffer, &value)
			valueBuffer = WriteValue(valueBuffer, &value)

			if dbErr := tx.Set(idBuffer, valueBuffer); dbErr != nil {
				return fmt.Errorf("failed to write value to db %w", dbErr)
			}

			ageIdxBuffer = WriteAgeIdx(ageIdxBuffer, &value)

			if dbErr := tx.Set(ageIdxBuffer, idxValue); dbErr != nil {
				return fmt.Errorf("failed to write age index to db %w", dbErr)
			}

			return nil
		})
		if err == nil {
			fmt.Println("Write completed")
		} else {
			panic(fmt.Errorf("write failed %w", err))
		}
	}

}

func readById(db *badger.DB, id uint64) {
	var idBuffer []byte
	var value Value
	err := db.View(func(tx *badger.Txn) error {
		value.ID = id
		idBuffer = WritePk(idBuffer, &value)
		item, dbErr := tx.Get(idBuffer)
		if dbErr != nil {
			return fmt.Errorf("failed to read item from the db %w", dbErr)
		}

		fmt.Println("Version of the item in the DB: ", item.Version())

		valueBuffer := make([]byte, item.ValueSize())
		valueBuffer, dbErr = item.ValueCopy(valueBuffer)
		if dbErr != nil {
			return fmt.Errorf("failed to read buffer from the value %w", dbErr)
		}

		dbErr = ReadValue(valueBuffer, &value)
		if dbErr != nil {
			return fmt.Errorf("failed to read value from the buffer %w", dbErr)
		}

		fmt.Println("Value from the DB: ", value)

		return nil
	})

	if err == nil {
		fmt.Println("Read completed")
	} else {
		fmt.Printf("Read failed %s\n", err)
	}
}

func readAll(db *badger.DB) {
	var value Value
	var valueBuffer []byte

	err := db.View(func(tx *badger.Txn) error {
		i := tx.NewIterator(badger.DefaultIteratorOptions)
		defer i.Close()

		for i.Rewind(); i.Valid(); i.Next() {
			var errDb error
			item := i.Item()

			keyId := item.Key()[0]
			switch keyId {
			case pk : errDb = ReadPk(item.Key(), &value)
			case ageIdx : errDb = ReadAgeIdx(item.Key(), &value)
			case sequence : {
				errDb = nil
				fmt.Println("Sequence object - ignored")
				continue
			}
			default: {
				fmt.Println("Unknown key id ", keyId)
				continue
			}
			}

			if errDb != nil {
				panic(fmt.Errorf("failed to read key %w", errDb))
			}

			if item.Key()[0] == pk {

				if cap(valueBuffer) < int(item.ValueSize()) {
					valueBuffer = make([]byte, item.ValueSize() * 2)
					fmt.Println("New value buffer created of capacity ", cap(valueBuffer))
				}
				valueBuffer = valueBuffer[:item.ValueSize()]
				valueBuffer, errDb = item.ValueCopy(valueBuffer)
				if errDb != nil {
					panic(fmt.Errorf("failed to copy value buffer %w", errDb))
				}
				errDb = ReadValue(valueBuffer, &value)
				if errDb != nil {
					panic(fmt.Errorf("failed to read value %w", errDb))
				}
				fmt.Println("Value ", value)
			} else {
				fmt.Println("Age key ", value)
			}
		}


		return nil
	})

	if err == nil {
		fmt.Println("Iteration completed")
	} else {
		fmt.Printf("Iteration failed %s\n", err)
	}
}

func readByAge(db *badger.DB, age uint8) {
	err := db.View(func(tx *badger.Txn) error {
		i := tx.NewIterator(badger.DefaultIteratorOptions)
		defer i.Close()
		prefix := make([]byte, 2)
		prefix[0] = ageIdx
		prefix[1] = age

		value := Value{}

		for i.Seek(prefix); i.ValidForPrefix(prefix); i.Next() {
			var errDb error
			item := i.Item()

			errDb = ReadAgeIdx(item.Key(), &value)
			if errDb != nil {
				panic(fmt.Errorf("failed to read key %w", errDb))
			}
			fmt.Println("Age key ", value)
		}

		return nil
	})

	if err == nil {
		fmt.Println("Iteration completed")
	} else {
		fmt.Printf("Iteration failed %s\n", err)
	}
}