package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"sync"
	"time"
	"encoding/json"
	"github.com/gobwas/glob"
)

const offset = "%02x"

var ErrNotFound = errors.New("not found in cache")

type CacheResolver interface {
	Get(string) (ValueResolver, error)
	GetByIndex(key string, i uint) (string, error)
	GetByKey(key string, innerKey string) (string, error)
	Set(string, []byte, int64) error
	Keys(string) map[string][]byte
	Remove(string) error
	Run()
	Close()

}

type cache struct {
	buckets map[string]*bucket
	GCChan  chan itemOnDelete
	stopGC  chan struct{}
}

type bucket struct {
	buckMux *sync.RWMutex
	items   map[string]ValueResolver
}

func NewCache() CacheResolver {
	c := cache{
		buckets: make(map[string]*bucket, 256),
		GCChan:  make(chan itemOnDelete),
		stopGC:  make(chan struct{}),
	}
	for i := 0; i < 256; i++ {
		c.buckets[fmt.Sprintf(offset, i)] = &bucket{
			buckMux: new(sync.RWMutex),
			items:   make(map[string]ValueResolver, 2048),
		}
	}
	return &c
}

func (c *cache) Get(key string) (ValueResolver, error) {
	return c.get(key)
}

func (c *cache) GetByIndex(key string, i uint) (string, error) {
	list := make([]string, 0)
	item, err := c.get(key)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(item.Body(), &list)
	if err != nil {
		return "", err
	}
	return list[i], nil
}

func (c *cache) GetByKey(key string, innerKey string) (string, error) {
	var dict map[string]string
	item, err := c.get(key)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(item.Body(), &dict)
	if err != nil {
		return "", err
	}
	return dict[innerKey], nil
}

func (c *cache) get(key string) (ValueResolver, error) {
	bucket, err := c.useBucket(key)
	if err != nil {
		return nil, err
	}
	bucket.buckMux.RLock()
	defer bucket.buckMux.RUnlock()
	item, ok := bucket.items[key]
	if !ok {
		return nil, ErrNotFound
	}
	return item, nil
}

type itemOnDelete struct {
	key string
	val ValueResolver
}

func (c *cache) Set(key string, data []byte, ttl int64) error {
	bucket, err := c.useBucket(key)
	if err != nil {
		return err
	}
	bucket.buckMux.Lock()
	v := newValue(data, ttl)
	bucket.items[key] = v
	bucket.buckMux.Unlock()

	c.GCChan <- itemOnDelete{key:key, val:v}
	fmt.Println("sended key", key)
	return nil
}

func (c *cache) Remove(key string) error {
	bucket, err := c.useBucket(key)
	if err != nil {
		return err
	}
	bucket.buckMux.Lock()
	defer bucket.buckMux.Unlock()
	delete(bucket.items, key)
	fmt.Println("deleted:", key)
	return nil
}

// https://github.com/gobwas/glob/blob/master/readme.md
func (c *cache) Keys(pattern string) (matchings map[string][]byte) {
	var g glob.Glob
	matchings = make(map[string][]byte)
	var wg sync.WaitGroup

	g = glob.MustCompile(pattern)
	for _, b := range c.buckets {
		wg.Add(1)
		go func(b *bucket) {
			defer wg.Done()
			for k, v := range b.items {
				if g.Match(k) {
					b.buckMux.Lock()
					matchings[k] = v.Body()
					b.buckMux.Unlock()
				}
			}
		}(b)
	}
	wg.Wait()
	return
}

func (c *cache) Run() {
	for {
		select {
		case item := <-c.GCChan:
			go func() {
				for range time.Tick(time.Second) {
					item.val.decrTTL()
					if item.val.TTL() == 0 {
						c.Remove(item.key)
						return
					}
				}
			}()
		case <- c.stopGC:
			return
		default:
			break
		}
	}
}

func (c *cache) useBucket(key string) (*bucket, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		return nil, err
	}
	bucketKey := fmt.Sprintf("%x", hasher.Sum(nil))[0:2]
	return c.buckets[bucketKey], nil
}

func (c *cache) Close() {
	close(c.stopGC)
	close(c.GCChan)
}

type ValueResolver interface {
	TTL() int64
	Body() []byte
	decrTTL()
}

type value struct {
	body      []byte
	ttl int64
}

func newValue(data []byte, ttl int64) ValueResolver {
	v := &value{
		body:      data,
		ttl: ttl,
	}
	return v
}

func (v *value) decrTTL() {
	v.ttl -= 1
}

func (v *value) TTL() int64 {
	return v.ttl
}

func (v *value) Body() []byte {
	return v.body
}



