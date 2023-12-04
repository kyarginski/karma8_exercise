package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTestRedis(t *testing.T) {
	got, err := NewTestRedis(t)
	assert.NoError(t, err)
	defer got.Close(t)
}
