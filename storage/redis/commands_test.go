package redis

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStringResult(t *testing.T) {
	cmd := NewStringResult("val", nil)
	val, err := cmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, "val", val)

	expErr := errors.New("bad")
	cmd = NewStringResult("", expErr)
	_, err = cmd.Result()
	assert.ErrorIs(t, err, expErr)
}

func TestNewStatusResult(t *testing.T) {
	cmd := NewStatusResult("OK", nil)
	val, err := cmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, "OK", val)

	expErr := errors.New("bad")
	cmd = NewStatusResult("", expErr)
	_, err = cmd.Result()
	assert.ErrorIs(t, err, expErr)
}

func TestNewIntResult(t *testing.T) {
	cmd := NewIntResult(42, nil)
	val, err := cmd.Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(42), val)

	expErr := errors.New("bad")
	cmd = NewIntResult(0, expErr)
	_, err = cmd.Result()
	assert.ErrorIs(t, err, expErr)
}
