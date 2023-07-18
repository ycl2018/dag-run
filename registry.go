package dagRun

import (
	"fmt"
	"log"
)

type Named interface{ Name() string }

type Registry[T Named] struct {
	mm map[string]T
}

func NewRegistry[T Named]() Registry[T] {
	return Registry[T]{mm: make(map[string]T)}
}

func (l Registry[T]) Register(t T) {
	if _, ok := l.mm[t.Name()]; ok {
		log.Panicf("duplicate register name:%s", t.Name())
	}
	l.mm[t.Name()] = t
}

func (l Registry[T]) Get(name string) (T, error) {
	if v, ok := l.mm[name]; ok {
		return v, nil
	}
	var v T
	return v, fmt.Errorf("not find register object by name:%s", name)
}

func (l Registry[T]) Has(name string) bool {
	if _, ok := l.mm[name]; ok {
		return true
	}
	return false
}
