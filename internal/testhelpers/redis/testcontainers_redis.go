package redis

/*
	usage:
	testRedis := testhelpers.NewTestContainerRedis(t)
	defer testRedis.Close(t)
	println(testRedis.ConnectionString(t))
*/

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestContainerRedis struct {
	instance testcontainers.Container
}

func newTestContainerRedis(t *testing.T) *TestContainerRedis {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)
	return &TestContainerRedis{
		instance: redisC,
	}
}

func (r *TestContainerRedis) Port(t *testing.T) int {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := r.instance.MappedPort(ctx, "6379")
	require.NoError(t, err)
	return p.Int()
}

func (r *TestContainerRedis) ConnectionString(t *testing.T) string {
	return fmt.Sprintf("redis://localhost:%d?protocol=3", r.Port(t))
}

func (r *TestContainerRedis) Close(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	require.NoError(t, r.instance.Terminate(ctx))
}
