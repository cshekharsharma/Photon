package redis

import (
	"errors"
	"testing"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestIsNil(t *testing.T) {
	assert.True(t, IsNil(goredis.Nil))
	assert.True(t, IsNil(Nil))
	assert.False(t, IsNil(errors.New("boom")))
	assert.False(t, IsNil(nil))
}
