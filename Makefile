.PHONY: all

include .env-compose-make
export $(shell sed 's/=.*//' .env-compose-make)
export $(shell sed 's/=.*//' .env)

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

docker:
	docker build . -t openhealthsuite/diary:testimage

docker-run:
	docker run -it --net=host --rm -p 3333:3333 -e OPENFOODDIARY_USERID="f1750ac3-d6cc-4981-9466-f1de2ebbad33" -e PORT=3333 -e OPENFOODDIARY_POSTGRES_CONNECTION_STRING=$$OPENFOODDIARY_POSTGRES_CONNECTION_STRING openhealthsuite/diary:testimage
