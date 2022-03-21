package datatypes

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestSFID(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		var id = SFID(12312312313123)

		NewWithT(t).Expect(id.String()).To(Equal("12312312313123"))
	})

	t.Run("MarshalText", func(t *testing.T) {
		id, _ := SFID(12312312313123).MarshalText()

		NewWithT(t).Expect(string(id)).To(Equal("12312312313123"))
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		var id SFID

		err := id.UnmarshalText([]byte("12312312313123"))
		NewWithT(t).Expect(err).To(BeNil())

		NewWithT(t).Expect(id).To(Equal(SFID(12312312313123)))
	})
}
