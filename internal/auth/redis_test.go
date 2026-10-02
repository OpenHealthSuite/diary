package auth

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const testRedisImage = "docker.io/redis:8"

var testRedisUrl string

func TestMain(m *testing.M) {
	code, err := runWithRedis(m)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not start redis for the auth tests: %v\n", err)
		os.Exit(1)
	}
	os.Exit(code)
}

func runWithRedis(m *testing.M) (int, error) {
	ctx := context.Background()

	container, err := testcontainers.Run(ctx, testRedisImage,
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("6379/tcp").WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return 0, err
	}
	defer container.Terminate(context.Background())

	host, err := container.Host(ctx)
	if err != nil {
		return 0, err
	}
	port, err := container.MappedPort(ctx, nat.Port("6379/tcp"))
	if err != nil {
		return 0, err
	}
	testRedisUrl = fmt.Sprintf("redis://%s:%s", host, port.Port())

	return m.Run(), nil
}
