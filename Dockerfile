FROM --platform=$BUILDPLATFORM docker.io/golang:1.25.1 AS server-builder
ARG TARGETPLATFORM
WORKDIR /usr/src/app

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=$(echo $TARGETPLATFORM | sed 's/linux\///') \
  go build -o dist/server cmd/server/main.go

FROM docker.io/debian:stable-slim AS runner

WORKDIR /app
COPY api api
COPY --from=server-builder /usr/src/app/dist/server /app

# Copy templates and static assets
COPY web/template /app/web/template
COPY web/static /app/web/static

EXPOSE 8080
ENV GIN_MODE=release
CMD ["./server"]
