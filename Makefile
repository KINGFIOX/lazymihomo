BIN := bin/lazymihomo

.PHONY: build test vet run clean

build:
	go build -trimpath -ldflags "-s -w" -o $(BIN) .

test:
	go test ./...

vet:
	go vet ./...

run:
	go run .

clean:
	rm -rf bin
