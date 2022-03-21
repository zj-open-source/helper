package redis

import (
	"context"
	"github.com/gomodule/redigo/redis"
)

type Conn = redis.Conn

func Command(name string, args ...interface{}) *CMD {
	return &CMD{
		name: name,
		args: args,
	}
}

type CMD struct {
	name string
	args []interface{}
}

type RedisOperator interface {
	Prefix(key string) string
	Get() Conn
	GetContext(ctx context.Context) (Conn, error)
	Exec(cmd *CMD, others ...*CMD) (interface{}, error)
	ExecContext(ctx context.Context, cmd *CMD, others ...*CMD) (interface{}, error)
}
