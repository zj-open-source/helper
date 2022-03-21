package memory

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestMemoryKVStorage(t *testing.T) {
	c := NewMemoryKVStorage()

	key := "key"
	value := "value"

	t.Run("store always", func(t *testing.T) {
		NewWithT(t).Expect(c.Store(key, value, -1)).To(BeNil())

		v := ""
		NewWithT(t).Expect(c.Load(key, &v)).To(BeNil())
		NewWithT(t).Expect(v).To(Equal(value))
	})

	t.Run("store expired", func(t *testing.T) {
		NewWithT(t).Expect(c.Store(key, value, 1*time.Second)).To(BeNil())
		time.Sleep(2 * time.Second)

		v := ""
		NewWithT(t).Expect(c.Load(key, &v)).To(BeNil())
		NewWithT(t).Expect(v).To(BeEmpty())
	})

	t.Run("load and del", func(t *testing.T) {
		NewWithT(t).Expect(c.Store(key, value, -1)).To(BeNil())

		{
			v := ""
			NewWithT(t).Expect(c.LoadAndDel(key, &v)).To(BeNil())
			NewWithT(t).Expect(v).To(Equal(value))
		}

		{
			v := ""
			NewWithT(t).Expect(c.Load(key, &v)).To(BeNil())
			NewWithT(t).Expect(v).To(BeEmpty())
		}
	})
}
