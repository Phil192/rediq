package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

type shard struct {
	shMux sync.RWMutex
	items map[string]*Value
}

func (s *shard) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")
	length := len(s.items)
	count := 0
	for key, value := range s.items {
		jsonValue, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		buffer.WriteString(fmt.Sprintf("\"%s\":%s", key, string(jsonValue)))
		count++
		if count < length {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
	return buffer.Bytes(), nil
}

func (s *shard) UnmarshalJSON(b []byte) error {
	s.shMux = sync.RWMutex{}
	s.items = make(map[string]*Value, 0) // idk for now how to detect shard len from damp
	err := json.Unmarshal(b, &s.items)
	if err != nil {
		return err
	}
	for key, value := range s.items {
		s.items[key] = value
	}
	return nil
}
