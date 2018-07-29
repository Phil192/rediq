run-server:
	go run main.go

run-client:
	go run cmd/client/main.go

test-cache:
	go test ./storage

test-rest:
	go test ./rest