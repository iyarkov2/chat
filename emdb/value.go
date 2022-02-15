package main

import (
	"encoding/binary"
	"fmt"
)

const (
	pk byte = 1
	ageIdx byte = 2
	sequence byte = 3
)

type Value struct {
	ID uint64
	Name string
	Age uint8
}

func WritePk(buffer []byte, value *Value) []byte {
	size := binary.Size(value.ID) + 1
	if len(buffer) != size {
		fmt.Println("Creating new ID buffer with size ", size)
		buffer = make([]byte, size)
	}
	buffer[0] = pk
	binary.BigEndian.PutUint64(buffer[1:], value.ID)
	return buffer
}

func ReadPk(buffer []byte, value *Value) error {
	size := binary.Size(value.ID) + 1
	if len(buffer) != size {
		return fmt.Errorf("invalid ID buffer size. expecting %d, actual %d", size, len(buffer))
	}
	if buffer[0] != pk {
		return fmt.Errorf("invalid key type, expecting %d, actual %d", pk, buffer[0])
	}
	value.ID = binary.BigEndian.Uint64(buffer[1:])
	return nil
}

func WriteValue(buffer []byte, v *Value) []byte {
	size := len(v.Name) + 1
	if cap(buffer) < size {
		fmt.Println("Creating new value buffer with size ", size)
		buffer = make([]byte, size)
	} else {
		buffer = buffer[:size]
	}
	copy(buffer, v.Name)
	buffer[len(v.Name)] = v.Age
	return buffer
}

func ReadValue(buffer []byte, v *Value) error {
	if len(buffer) < 1 {
		return fmt.Errorf("invalid value size, must be at least 1 byte")
	} else if len(buffer) == 1 {
		v.Name = ""
		v.Age = buffer[0]
	} else {
		v.Name = string(buffer[:len(buffer) - 1])
		v.Age = buffer[len(buffer) - 1]
	}
	return nil
}

func WriteAgeIdx(buffer []byte, value *Value) []byte {
	size := binary.Size(value.ID) + 2
	if len(buffer) != size {
		fmt.Println("Creating new ID buffer with size ", size)
		buffer = make([]byte, size)
	}
	buffer[0] = ageIdx
	buffer[1] = value.Age
	binary.BigEndian.PutUint64(buffer[2:], value.ID)
	return buffer
}

func ReadAgeIdx(buffer []byte, value *Value) error {
	size := binary.Size(value.ID) + 2
	if len(buffer) != size {
		return fmt.Errorf("invalid ID buffer size. expecting %d, actual %d", size, len(buffer))
	}
	if buffer[0] != ageIdx {
		return fmt.Errorf("invalid key type, expecting %d, actual %d", ageIdx, buffer[0])
	}
	value.Name = ""
	value.Age = buffer[1]
	value.ID = binary.BigEndian.Uint64(buffer[2:])
	return nil

}