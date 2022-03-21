package kvstorage

import (
	"context"
	"time"
)

type KVStorage interface {
	Store(key string, value interface{}, expiresIn time.Duration) error
	Load(key string, value interface{}) error
	LoadAndDel(key string, value interface{}) error
	Del(key string) error

	Context() context.Context
	WithContext(ctx context.Context) KVStorage
}
