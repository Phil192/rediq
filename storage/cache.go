package storage

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/gobwas/glob"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
	"time"
)

type cache struct {
	shards map[string]*shard
	GCChan chan itemOnDelete
	stopGC chan struct{}

	opt *cacheOptions
}

func NewCache(opts ...cacheOpt) Storer {
	c := cache{
		GCChan: make(chan itemOnDelete, 256),
		stopGC: make(chan struct{}),
		opt: &cacheOptions{
			2048,
			256,
			".",
		},
	}
	for _, o := range opts {
		if o != nil {
			o(c.opt)
		}
	}
	c.shards = make(map[string]*shard, c.opt.BucketsNum)
	return &c
}

func (c *cache) Get(key string) (*Value, error) {
	return c.get(key)
}

func (c *cache) GetBy(key string, subSeq interface{}) (interface{}, error) {
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
		return body.Index(int(subKey)).Interface(), nil
	case reflect.String:
		return body.MapIndex(reflect.ValueOf(subSeq)).Interface(), nil
	default:
		return nil, ErrSubSeqType
	}
}

func (c *cache) get(key string) (*Value, error) {
	shard, _, err := c.getOrCreateShard(key)
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

func (c *cache) Set(key string, data interface{}, ttl time.Duration) error {
	shard, _, err := c.getOrCreateShard(key)
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
	log.Debugln("set key:", key, "with value:", b.items[key])
}

func (c *cache) Remove(key string) error {
	shard, shardKey, err := c.getOrCreateShard(key)
	if err != nil {
		return err
	}
	shard.shMux.Lock()

	_, ok := shard.items[key]
	if ok {
		delete(shard.items, key)
		log.Debugln("deleted:", key)
	}
	shard.shMux.Unlock()
	if len(shard.items) == 0 {
		delete(c.shards, shardKey)
	}
	return nil
}

// https://github.com/gobwas/glob/blob/master/readme.md
func (c *cache) Keys(mask string) []string {
	var g glob.Glob
	matchings := make([]string, 0)
	var wg sync.WaitGroup
	mx := new(sync.Mutex)

	g = glob.MustCompile(mask)
	for _, sh := range c.shards {
		wg.Add(1)
		go func(sh *shard) {
			defer wg.Done()
			for k := range sh.items {
				if g.Match(k) {
					mx.Lock()
					defer mx.Unlock()
					matchings = append(matchings, k)
				}
			}
		}(sh)
	}
	wg.Wait()
	log.Debugln("keys found:", len(matchings))
	return matchings
}

func (c *cache) Run() {
	go c.doExpiration()
	c.readDump()
}

func (c *cache) readDump() {
	defer os.Remove(c.opt.DumpPath)
	data, err := ioutil.ReadFile(c.opt.DumpPath)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return
		}
		log.Warningln("fail to read dumped data:", err)
		return
	}
	if err := json.Unmarshal(data, &c.shards); err != nil {
		log.Warningln("fail to unmarshal dumped data:", err)
		return
	}
	var wg sync.WaitGroup
	for _, sh := range c.shards {
		wg.Add(1)
		go func(sh *shard) {
			defer wg.Done()
			for k, v := range sh.items {
				c.GCChan <- itemOnDelete{key: k, val: v}
			}
		}(sh)
	}
	wg.Wait()
}

func (c *cache) dumpData() error {
	if len(c.shards) == 0 {
		return nil
	}
	for tryOut := 3; tryOut > 0; tryOut-- {
		data, err := json.Marshal(c.shards)
		if err != nil {
			log.Warningf("fail to dump data: %d tryout", tryOut)
			continue
		}
		if err := ioutil.WriteFile(c.opt.DumpPath, data, 0755); err != nil {
			log.Warningf("fail to dump data: %d tryout", tryOut)
			continue
		}
		return nil
	}
	return ErrDumpFail
}

func (c *cache) doExpiration() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)
	for {
		select {
		case item := <-c.GCChan:
			log.Debugln("item to purge", item.key, "time", item.val.TTL)
			if item.val.TTL == 0 {
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
		case <-interrupt:
			c.Close()
			log.Warningln("interrupted")
			os.Exit(1)
		default:
			break
		}
	}
}

func (c *cache) Close() {
	if err := c.dumpData(); err != nil {
		log.Warningln(err)
		return
	}
	close(c.GCChan)
	log.Debugln("cache closed")
}

func (c *cache) getOrCreateShard(key string) (*shard, string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(key))
	if err != nil {
		return nil, "", err
	}
	shardKey := fmt.Sprintf("%x", hasher.Sum(nil))[0:2]
	_, ok := c.shards[shardKey]
	if !ok {
		c.newShard(shardKey)
	}
	return c.shards[shardKey], shardKey, nil
}

func (c *cache) newShard(shardKey string) {
	c.shards[shardKey] = &shard{
		shMux: new(sync.RWMutex),
		items: make(map[string]*Value, c.opt.ItemsNum),
	}
}
