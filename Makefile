.PHONY: all

# include .env-compose-make
# include .env
# export $(shell sed 's/=.*//' .env-compose-make)
# export $(shell sed 's/=.*//' .env)

all: build

generate:
	go generate ./...
build:
	go build -o dist/server cmd/server/main.go
test:
	go test -v ./...
run:
	go run cmd/server/main.go
# runapi:
# 	PORT=8081 go run cmd/apiserver/main.go
# watch:
# 	gow run cmd/server/main.go
# watchapi:
# 	PORT=8081 gow run cmd/apiserver/main.go

# devdatabrowser:
# 	(cd web/app/databrowser && npm run dev)

# docker:
# 	docker build . -t quickcategory:testimage

# docker-run:
# 	docker run -it --net=host --rm -p 3333:3333 -e PORT=3333 -e POSTGRES_CONNECTION_STRING=$$POSTGRES_CONNECTION_STRING quickcategory:testimage

# seed-test-data:
# 	go run cmd/test_seed/main.go 