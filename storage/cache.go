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
	buckets map[string]*shard
	GCChan  chan itemOnDelete
	stopGC  chan struct{}

	opt *cacheOptions
}

type shard struct {
	shMux *sync.RWMutex
	items   map[string]*Value
}

func NewCache(opts ...cacheOpt) Storer {
	c := cache{
		GCChan:  make(chan itemOnDelete),
		stopGC:  make(chan struct{}),
		opt: &cacheOptions{
			2048,
			256,
			-1,
		},
	}
	for _, o := range opts {
		if o != nil {
			o(c.opt)
		}
	}
	c.buckets = make(map[string]*shard, c.opt.BucketsNum)
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
	switch reflect.TypeOf(item.Body).Kind() {
	case reflect.Slice:
		body = reflect.ValueOf(item.Body)
	case reflect.Map:
		body = reflect.ValueOf(item.Body)
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
	shard, err := c.useOrCreateShard(key)
	if err != nil {
		return nil, err
	}
	shard.shMux.RLock()
	defer shard.shMux.RUnlock()
	item, ok := shard.items[key]
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
	shard, err := c.useOrCreateShard(key)
	if err != nil {
		return err
	}
	v, err := newValue(data, c.opt.TTL)
	if err != nil {
		return err
	}
	c.set(shard, key, v)
	c.GCChan <- itemOnDelete{key: key, val: v}

	return nil
}

func (c *cache) SetWithTTL(key string, data interface{}, ttl int) error {
	shard, err := c.useOrCreateShard(key)
	if err != nil {
		return err
	}

	v, err := newValue(data, ttl)
	if err != nil {
		return err
	}
	c.set(shard, key, v)
	c.GCChan <- itemOnDelete{key: key, val: v}

	return nil
}

func (c *cache) set(b *shard, key string, v *Value) {
	b.shMux.Lock()
	defer b.shMux.Unlock()
	b.items[key] = v
	log.Debugln("set key:", key)
}

func (c *cache) Remove(key string) error {
	bucket, err := c.useOrCreateShard(key)
	if err != nil {
		return err
	}
	bucket.shMux.Lock()
	defer bucket.shMux.Unlock()
	_, ok := bucket.items[key]
	if ok {
		delete(bucket.items, key)
		log.Debugln("deleted:", key)
	}
	return nil
}

// https://github.com/gobwas/glob/blob/master/readme.md
// syncmap is slow but for this method it is ok
func (c *cache) Keys(pattern string) *[]string {
	log.Debugln("keys pattern", pattern)
	var g glob.Glob
	matchings := make([]string, 0)
	var wg sync.WaitGroup

	g = glob.MustCompile(pattern)
	for _, b := range c.buckets {
		wg.Add(1)
		go func(b *shard) {
			defer wg.Done()
			for k := range b.items {
				if g.Match(k) {
					b.shMux.Lock()
					defer b.shMux.Unlock()
					matchings = append(matchings, k)
				}
			}
		}(b)
	}
	wg.Wait()
	log.Debugln("keys done")
	return &matchings
}

func (c *cache) Run() {
	for {
		select {
		case item := <-c.GCChan:
			log.Debugln("item to purge", item.val.TTL)
			if item.val.TTL < 0 {
				break
			}
			go func() {
				for range time.Tick(time.Second) {
					if ok := item.val.decrTTL(); !ok {
						c.Remove(item.key)
						return
					}

				}
			}()
		case <-c.stopGC:
			close(c.GCChan)
			return
		default:
			break
		}
	}
}

func (c *cache) Close() {
	close(c.stopGC)
	log.Debugln("cache closed")
}

func (c *cache) useOrCreateShard(key string) (*shard, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		return nil, err
	}
	bucketKey := fmt.Sprintf("%x", hasher.Sum(nil))[0:2]
	_, ok := c.buckets[bucketKey]
	if !ok {
		c.newShard(bucketKey)
	}
	return c.buckets[bucketKey], nil
}

func (c *cache) newShard(bucketKey string) {
	c.buckets[bucketKey] = &shard{
		shMux: new(sync.RWMutex),
		items:   make(map[string]*Value, c.opt.ItemsNum),
	}
}

