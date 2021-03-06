package storage

import (
	"errors"
	"reflect"
	"time"
)

var ErrNotFound = errors.New("not found in cache")
var ErrSubSeqType = errors.New("subsequence must be defined by string or positive integer")
var ErrNotSequence = errors.New("returned value is not subsequence. Use Get method instead.")
var ErrUnknownDataType = errors.New("only strings, maps and slices are supported.")
var ErrNegativeTTL = errors.New("ttl must be positive integer")
var ErrDumpFail = errors.New("fail to dump data")

type InputType int

const (
	STR InputType = iota
	ARRAY
	MAPPING
)

type Storer interface {
	Get(string) (*Value, error)
	GetBy(string, interface{}) (interface{}, error)
	Set(string, interface{}, time.Duration) error
	Keys(string) []string
	Remove(string) error
	Run()
	Close()
}

type Value struct {
	Body     interface{}   `json:"body"`
	TTL      time.Duration `json:"ttl"`
	DataType InputType     `json:"-"`
}

func newValue(data interface{}, ttl time.Duration) (*Value, error) {
	if ttl < 0 {
		return nil, ErrNegativeTTL
	}
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
		Body:     data,
		TTL:      ttl,
		DataType: dataType,
	}
	return v, nil
}

func (v *Value) decrTTL() bool {
	v.TTL -= 1 * time.Second
	return v.TTL > 0
}
