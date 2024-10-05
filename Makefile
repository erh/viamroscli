
module: module.tar.gz

bin/viamrosclimodule: go.mod *.go cmd/module/*.go
	go build -o bin/viamrosclimodule cmd/module/cmd.go

bin/static/viamrosclimodule: go.mod *.go cmd/module/*.go
	mkdir -p bin/static
	go build -tags osusergo,netgo -ldflags="-extldflags=-static -s -w" -o $@ cmd/module/cmd.go

lint:
	gofmt -s -w .

updaterdk:
	go get go.viam.com/rdk@latest
	go mod tidy

test:
	go test ./...


module.tar.gz: bin/viamrosclimodule
	tar czf $@ $^

all: test bin/viamroscli module 


