run-server:
	go run cmd/server/main.go

run-client:
	go run cmd/client/main.go

test-cache:
	go test storage/cache_test.go storage/cache.go storage/storer.go storage/shard.go storage/options.go storage/elastic.go storage/cirq.go

test-rest:
	go test rest/listener_test.go rest/options.go rest/middleware.go rest/listener.go