package redis

import (
	"testing"

	"github.com/gomodule/redigo/redis"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	r := &Redis{
		Host: "127.0.0.1",
		Port: 6379,
	}

	r.SetDefaults()
	r.Init()

	t.Run("Set", func(t *testing.T) {
		_, err := r.Exec(Command("SET", r.Prefix("KEY"), "1"))
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("Set", func(t *testing.T) {
		result, err := redis.String(r.Exec(Command("GET", r.Prefix("KEY"))))
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(result).To(Equal("1"))
	})

	t.Run("Del", func(t *testing.T) {
		_, err := r.Exec(Command("Del", r.Prefix("KEY")))
		NewWithT(t).Expect(err).To(BeNil())
	})

	t.Run("Exists", func(t *testing.T) {
		exists, err := redis.Bool(r.Exec(Command("EXISTS", r.Prefix("KEY"))))
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(exists).To(BeFalse())
	})

	t.Run("Multi", func(t *testing.T) {
		values, err := redis.Values(
			r.Exec(
				Command("SET", r.Prefix("KEY"), "1"),
				Command("GET", r.Prefix("KEY")),
				Command("DEL", r.Prefix("KEY")),
			),
		)
		NewWithT(t).Expect(err).To(BeNil())
		NewWithT(t).Expect(values[0]).To(Equal("OK"))
	})
}
