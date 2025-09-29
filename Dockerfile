FROM --platform=$BUILDPLATFORM docker.io/golang:1.25.1 AS server-builder
ARG TARGETPLATFORM
WORKDIR /usr/src/app

COPY go.mod go.mod
COPY go.sum go.sum

RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=$(echo $TARGETPLATFORM | sed 's/linux\///') \
  go build -o dist/server cmd/server/main.go

FROM --platform=$BUILDPLATFORM docker.io/node:18.7.0 AS databrowser-builder
ARG TARGETPLATFORM
WORKDIR /usr/src/app

COPY web/app .
RUN npm ci
RUN npm run build

FROM docker.io/debian:stable-slim AS runner

RUN apt update -y && apt install ffmpeg -y
WORKDIR /app
COPY api api
COPY --from=server-builder /usr/src/app/dist/server /app
COPY --from=databrowser-builder /usr/src/app/dist /app/web/dist

EXPOSE 8080
ENV GIN_MODE=release
CMD ["./server"]