build:
	go build -o gogo

deploy: setup test
	GOOS=linux GOARCH=amd64 go build
	scripts/deploy
	rm gogo gogo.tar.gz

run: setup build
	scripts/run

setup:
	brew services start redis
	go get

test:
	go fmt ./...
	go vet ./...
	go test ./...
