PATH_MAIN=./cmd/main.go

run:
	go run ${PATH_MAIN} -port=3000 -logging
test:
	go test ./...