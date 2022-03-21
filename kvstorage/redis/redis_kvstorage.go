package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-courier/metax"
	"github.com/gomodule/redigo/redis"
	"github.com/zj-open-source/tools/kvstorage"
	redis1 "github.com/zj-open-source/tools/redis"
)

var _ kvstorage.KVStorage = (*RedisKVStorage)(nil)

func NewRedisKVStorage(op redis1.RedisOperator) *RedisKVStorage {
	return &RedisKVStorage{
		op: op,
	}
}

type RedisKVStorage struct {
	op redis1.RedisOperator
	metax.Ctx
}

func (s *RedisKVStorage) WithContext(ctx context.Context) kvstorage.KVStorage {
	return &RedisKVStorage{
		op:  s.op,
		Ctx: s.Ctx.WithContext(ctx),
	}
}

func (s *RedisKVStorage) Del(key string) error {
	_, err := s.op.Exec(redis1.Command("DEL", s.op.Prefix(key)))
	return err
}

type data struct {
	Value interface{} `json:"value"`
}

func (s *RedisKVStorage) Store(key string, value interface{}, expiresIn time.Duration) error {
	bytes, errMarshal := json.Marshal(data{Value: value})
	if errMarshal != nil {
		return errMarshal
	}

	if expiresIn > 0 {
		_, err := s.op.Exec(
			redis1.Command("SET", s.op.Prefix(key), bytes),
			redis1.Command("EXPIRE", s.op.Prefix(key), transToSecond(expiresIn)),
		)
		return err
	}

	_, err := s.op.Exec(
		redis1.Command("SET", s.op.Prefix(key), bytes),
	)
	return err
}

func (s *RedisKVStorage) LoadAndDel(key string, value interface{}) error {
	values, err := redis.Values(s.op.Exec(
		redis1.Command("GET", s.op.Prefix(key)),
		redis1.Command("DEL", s.op.Prefix(key)),
	))
	if err != nil {
		if err == redis.ErrNil {
			return nil
		}
		return err
	}
	if bytes, ok := values[0].([]byte); ok {
		return json.Unmarshal(bytes, &data{Value: value})
	}
	return nil
}

func (s *RedisKVStorage) Load(key string, value interface{}) error {
	bytes, err := redis.Bytes(s.op.Exec(redis1.Command("GET", s.op.Prefix(key))))
	if err != nil {
		if err == redis.ErrNil {
			return nil
		}
		return err
	}

	return json.Unmarshal(bytes, &data{Value: value})
}

func transToSecond(dur time.Duration) int64 {
	if dur > 0 && dur < time.Second {
		return 0
	}
	return int64(dur / time.Second)
}
