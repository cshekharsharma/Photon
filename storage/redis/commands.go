package redis

import "github.com/redis/go-redis/v9"

// Type aliases and helpers to avoid leaking go-redis into downstream packages.
type Client = redis.Client
type StringCmd = redis.StringCmd
type StatusCmd = redis.StatusCmd
type IntCmd = redis.IntCmd

var Nil = redis.Nil

func NewStringResult(val string, err error) *StringCmd {
	return redis.NewStringResult(val, err)
}

func NewStatusResult(val string, err error) *StatusCmd {
	return redis.NewStatusResult(val, err)
}

func NewIntResult(val int64, err error) *IntCmd {
	return redis.NewIntResult(val, err)
}
