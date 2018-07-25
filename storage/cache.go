package storage

import (
	"crypto/sha1"
	"fmt"
	"github.com/gobwas/glob"
	"sync"
	"time"
	"reflect"
	log "github.com/sirupsen/logrus"
)


type cache struct {
	buckets map[string]*bucket
	GCChan  chan itemOnDelete
	stopGC  chan struct{}

	opt *cacheOptions
}

type bucket struct {
	buckMux *sync.RWMutex
	items   map[string]*Value
}

func NewCache(opts ...cacheOpt) Storer {
	c := cache{
		GCChan:  make(chan itemOnDelete),
		stopGC:  make(chan struct{}),
		opt: &cacheOptions{
			2048,
			256,
			0,
			time.Second,
		},
	}
	for _, o := range opts {
		if o != nil {
			o(c.opt)
		}
	}
	c.buckets = make(map[string]*bucket, c.opt.BucketsNum)
	return &c
}

func (c *cache) Get(key string) (*Value, error) {
	return c.get(key)
}

func (c *cache) GetContent(key string, subSeq interface{}) ([]byte, error) {
	item, err := c.get(key)
	if err != nil {
		return nil, err
	}
	var body reflect.Value
	switch reflect.TypeOf(item.Body()).Kind() {
	case reflect.Slice:
		body = reflect.ValueOf(item.Body())
	case reflect.Map:
		body = reflect.ValueOf(item.Body())
	default:
		return nil, ErrNotSequence
	}
	switch reflect.TypeOf(subSeq).Kind() {
	case reflect.Uint, reflect.Int:
		ss := reflect.ValueOf(subSeq)
		subKey := ss.Int()
		return body.Index(int(subKey)).Bytes(), nil
	case reflect.String:
		return body.MapIndex(reflect.ValueOf(subSeq)).Bytes(), nil
	default:
		return nil, ErrSubSeqType
	}
}

func (c *cache) get(key string) (*Value, error) {
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
	val *Value
}

func (c *cache) Set(key string, data interface{}) error {
	bucket, err := c.useBucket(key)
	if err != nil {
		return err
	}
	v, err := newValue(data, c.opt.TTL)
	if err != nil {
		return err
	}
	c.set(bucket, key, v)
	if c.opt.TTL != 0 {
		c.GCChan <- itemOnDelete{key: key, val: v}
	}
	return nil
}

func (c *cache) SetWithTTL(key string, data interface{}, ttl uint64) error {
	bucket, err := c.useBucket(key)
	if err != nil {
		return err
	}

	v, err := newValue(data, ttl)
	if err != nil {
		return err
	}
	c.set(bucket, key, v)
	if ttl != 0 {
		c.GCChan <- itemOnDelete{key: key, val: v}
	}
	return nil
}

func (c *cache) set(b *bucket, key string, v *Value) {
	b.buckMux.Lock()
	defer b.buckMux.Unlock()
	b.items[key] = v
	log.Debugln("set key:", key)
}

func (c *cache) Remove(key string) error {
	bucket, err := c.useBucket(key)
	if err != nil {
		return err
	}
	bucket.buckMux.Lock()
	defer bucket.buckMux.Unlock()
	delete(bucket.items, key)
	log.Debugln("deleted:", key, "from bucket", bucket)
	return nil
}

// https://github.com/gobwas/glob/blob/master/readme.md
// syncmap is slow but for this method it is ok
func (c *cache) Keys(pattern string) (*sync.Map) {
	var g glob.Glob
	matchings := new(sync.Map)
	var wg sync.WaitGroup

	g = glob.MustCompile(pattern)
	for _, b := range c.buckets {
		wg.Add(1)
		go func(b *bucket) {
			defer wg.Done()
			for k, v := range b.items {
				if g.Match(k) {
					matchings.Store(k, v.Body())
				}
			}
		}(b)
	}
	wg.Wait()
	return matchings
}

func (c *cache) Run() {
	for {
		select {
		case item := <-c.GCChan:
			go func() {
				for range time.Tick(c.opt.TimeUnit) {
					item.val.decrTTL()
					if item.val.TTL() == 0 {
						c.Remove(item.key)
						return
					}
				}
			}()
		case <-c.stopGC:
			return
		default:
			break
		}
	}
}

func (c *cache) Close() {
	close(c.stopGC)
	close(c.GCChan)
	log.Debugln("cache closed")
}

func (c *cache) useBucket(key string) (*bucket, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		return nil, err
	}
	bucketKey := fmt.Sprintf("%x", hasher.Sum(nil))[0:2]
	_, ok := c.buckets[bucketKey]
	if !ok {
		c.newBucket(bucketKey)
	}
	return c.buckets[bucketKey], nil
}

func (c *cache) newBucket(bucketKey string) {
	c.buckets[bucketKey] = &bucket{
		buckMux: new(sync.RWMutex),
		items:   make(map[string]*Value, c.opt.ItemsNum),
	}
}

