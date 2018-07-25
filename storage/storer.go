package storage

import (
	"sync"
	"reflect"
	"errors"
)

var ErrNotFound = errors.New("not found in cache")
var ErrSubSeqType = errors.New("subsequence must be defined by string or positive integer")
var ErrNotSequence = errors.New("returned value is not subsequence. Use Get method instead.")
var ErrUnknownDataType = errors.New("only strings, maps and slices are supported.")
var ErrNegativeTTL = errors.New("ttl must be positive integer")

type InputType int

const (
	STR InputType = iota
	ARRAY
	MAPPING
)

type Storer interface {
	Get(string) (*Value, error)
	GetContent(string, interface{}) ([]byte, error)
	Set(string, interface{}) error
	SetWithTTL(string, interface{}, uint64) error
	Keys(string) *sync.Map
	Remove(string) error
	Run()
	Close()
}

type Value struct {
	body interface{}
	ttl  uint64
	dataType InputType
}

func newValue(data interface{}, ttl uint64) (*Value, error) {
	var dataType InputType
	switch reflect.ValueOf(data).Kind() {
	case reflect.String:
		dataType = STR
	case reflect.Slice:
		dataType = ARRAY
	case reflect.Map:
		keyType := reflect.TypeOf(data).Key().Kind()
		if keyType != reflect.String {
			return nil, ErrUnknownDataType
		}
		dataType = MAPPING
	default:
		return nil, ErrUnknownDataType
	}
	v := &Value{
		body: data,
		ttl:  ttl,
		dataType: dataType,
	}
	return v, nil
}

func (v *Value) decrTTL() {
	v.ttl -= 1
}

func (v *Value) TTL() uint64 {
	return v.ttl
}

func (v *Value) Body() interface{} {
	return v.body
}

func (v *Value) Type() InputType {
	return v.dataType
}