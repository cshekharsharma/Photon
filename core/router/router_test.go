package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChiRouter(t *testing.T) {
	router := NewChiRouter()
	assert.NotNil(t, router, "NewChiRouter should create a new router instance")
}

func TestNewRouter(t *testing.T) {
	router := NewRouter(RouterCHI)
	assert.NotNil(t, router, "NewChiRouter should create a new router instance")

	router = NewRouter("random-router")
	assert.Nil(t, router)
}
