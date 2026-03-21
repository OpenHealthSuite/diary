.PHONY: all

all: build

generate:
	go generate ./...
build:
	go build -o dist/server cmd/server/main.go
test:
	go test -v ./...
run:
	OPENFOODDIARY_USERID="f1750ac3-d6cc-4981-9466-f1de2ebbad33" go run cmd/server/main.go
watch:
	OPENFOODDIARY_USERID="f1750ac3-d6cc-4981-9466-f1de2ebbad33" gow run cmd/server/main.go
