package storage


type cacheOpt func(o *cacheOptions)

type cacheOptions struct {
	ItemsNum uint
	BucketsNum uint
	TTL int
}

func ShardsNum(i uint) cacheOpt {
	return func(o *cacheOptions) {
		o.ItemsNum = i
	}
}

func ItemsPerShard(i uint) cacheOpt {
	return func(o *cacheOptions) {
		o.BucketsNum = i
	}
}

func DefaultTTL(i int) cacheOpt {
	return func(o *cacheOptions) {
		o.TTL = i
	}
}

