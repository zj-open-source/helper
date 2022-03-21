package redis

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-courier/envconf"
	"github.com/gomodule/redigo/redis"
)

var env = strings.ToLower(os.Getenv("GOENV"))
var projectName = strings.ToLower(os.Getenv("PROJECT_NAME"))
var prefix = fmt.Sprintf("%s:%s", env, projectName)

type RedisEndpoint struct {
	Endpoint envconf.Endpoint `env:""`
	Wait     bool
	pool     *redis.Pool
}

func (r *RedisEndpoint) Get() Conn {
	c, _ := r.GetContext(context.Background())
	return c
}

func (r *RedisEndpoint) GetContext(ctx context.Context) (Conn, error) {
	if r.pool != nil {
		return r.pool.GetContext(ctx)
	}
	return nil, nil
}

func (r *RedisEndpoint) Exec(cmd *CMD, others ...*CMD) (interface{}, error) {
	return r.ExecContext(context.Background(), cmd, others...)
}

func (r *RedisEndpoint) ExecContext(ctx context.Context, cmd *CMD, others ...*CMD) (interface{}, error) {
	c, err := r.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	if (len(others)) == 0 {
		return c.Do(cmd.name, cmd.args...)
	}

	err = c.Send("MULTI")
	if err != nil {
		return nil, err
	}

	err = c.Send(cmd.name, cmd.args...)
	if err != nil {
		return nil, err
	}

	for i := range others {
		o := others[i]
		if o == nil {
			continue
		}
		err := c.Send(o.name, o.args...)
		if err != nil {
			return nil, err
		}
	}

	return c.Do("EXEC")
}

func (r *RedisEndpoint) Prefix(key string) string {
	return fmt.Sprintf("%s:%s", prefix, key)
}

func (r *RedisEndpoint) LivenessCheck() map[string]string {
	m := map[string]string{}

	conn := r.Get()
	defer conn.Close()
	_, err := conn.Do("PING")
	if err != nil {
		m[r.Endpoint.Host()] = err.Error()
	} else {
		m[r.Endpoint.Host()] = "ok"
	}

	return m
}

func (r *RedisEndpoint) Init() {
	if r.pool == nil {
		r.initial()
	}
}

func (r *RedisEndpoint) initial() {
	opt := struct {
		ConnectTimeout envconf.Duration `name:"connectTimeout" default:"10s"`
		ReadTimeout    envconf.Duration `name:"readTimeout" default:"10s"`
		WriteTimeout   envconf.Duration `name:"writeTimeout" default:"10s"`
		IdleTimeout    envconf.Duration `name:"idleTimeout" default:"240s"`
		MaxActive      int              `name:"maxActive" default:"5"`
		MaxIdle        int              `name:"maxIdle" default:"3"`
		DB             int              `name:"db" default:"10"`
	}{}

	err := envconf.UnmarshalExtra(r.Endpoint.Extra, &opt)
	if err != nil {
		panic(err)
	}

	dialFunc := func() (c redis.Conn, err error) {
		options := []redis.DialOption{
			redis.DialDatabase(opt.DB),
			redis.DialConnectTimeout(time.Duration(opt.ConnectTimeout)),
			redis.DialWriteTimeout(time.Duration(opt.WriteTimeout)),
			redis.DialReadTimeout(time.Duration(opt.ReadTimeout)),
		}

		if r.Endpoint.Password != "" {
			options = append(options, redis.DialPassword(r.Endpoint.Password))
		}

		return redis.Dial(
			"tcp",
			r.Endpoint.Host(),
			options...,
		)
	}

	r.pool = &redis.Pool{
		Dial:        dialFunc,
		MaxIdle:     opt.MaxIdle,
		MaxActive:   opt.MaxActive,
		IdleTimeout: time.Duration(opt.IdleTimeout),
		Wait:        true,
	}
}
