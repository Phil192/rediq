package storage


import "time"

type cacheOpt func(o *cacheOptions)

type cacheOptions struct {
	ItemsNum uint
	BucketsNum uint
	TTL uint64
	TimeUnit time.Duration
}

func BucketsNum(i uint) cacheOpt {
	return func(o *cacheOptions) {
		o.ItemsNum = i
	}
}

func ItemsPerBucket(i uint) cacheOpt {
	return func(o *cacheOptions) {
		o.BucketsNum = i
	}
}

func DefaultTTL(i uint64) cacheOpt {
	return func(o *cacheOptions) {
		o.TTL = i
	}
}

func TimeUnit(unit time.Duration) cacheOpt {
	return func(o *cacheOptions) {
		o.TimeUnit = unit
	}
}
