package memory

import (
	"context"
	"reflect"
	"sync"
	"time"

	"github.com/go-courier/metax"
	"github.com/go-courier/reflectx"
	"github.com/zj-open-source/helper/kvstorage"
)

var _ kvstorage.KVStorage = (*MemoryKVStorage)(nil)

func NewMemoryKVStorage() *MemoryKVStorage {
	return &MemoryKVStorage{
		m: &sync.Map{},
	}
}

type MemoryKVStorage struct {
	m *sync.Map
	metax.Ctx
}

func (s *MemoryKVStorage) WithContext(ctx context.Context) kvstorage.KVStorage {
	return &MemoryKVStorage{
		m:   s.m,
		Ctx: s.Ctx.WithContext(ctx),
	}
}

type ValueWithExpire struct {
	Value     interface{}
	Always    bool
	ExpiredAt time.Time
}

func (s *MemoryKVStorage) Del(key string) error {
	s.m.Delete(key)
	return nil
}

func (s *MemoryKVStorage) Store(key string, value interface{}, expiresIn time.Duration) error {
	if expiresIn > 0 {
		s.m.Store(key, ValueWithExpire{
			Value:     value,
			ExpiredAt: time.Now().Add(expiresIn),
		})
		return nil
	}

	s.m.Store(key, ValueWithExpire{
		Value:  value,
		Always: true,
	})

	return nil
}

func (s *MemoryKVStorage) LoadAndDel(key string, value interface{}) error {
	if err := s.Load(key, value); err != nil {
		return err
	}
	return s.Del(key)
}

func (s *MemoryKVStorage) Load(key string, value interface{}) error {
	if val, ok := s.m.Load(key); ok {
		v := val.(ValueWithExpire)
		if !v.Always {
			if time.Now().After(v.ExpiredAt) {
				s.Del(key)
				return nil
			}
		}
		rv := reflectx.Indirect(reflect.ValueOf(value))
		rv.Set(reflectx.Indirect(reflect.ValueOf(v.Value)))
		return nil
	}
	return nil
}
