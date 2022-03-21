package datatypes

import (
	"testing"
	
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/require"
)

func TestImageAddress(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		t.Run("asset", func(t *testing.T) {
			var address = Address{
				Group: "avatar",
				Key:   "abc",
				Ext:   "png",
			}
			NewWithT(t).Expect(address.String()).To(Equal("asset://avatar/abc.png"))
		})
		t.Run("url", func(t *testing.T) {
			var img = Address{
				URL: "http://img.baidu.com/avatar/abc.png",
			}
			NewWithT(t).Expect(img.String()).To(Equal("http://img.baidu.com/avatar/abc.png"))
		})
		t.Run("asset", func(t *testing.T) {
			var address = Address{
				Group: "avatar",
				Key:   "abc",
			}
			NewWithT(t).Expect(address.String()).To(Equal("asset://avatar/abc"))
		})

	})
	t.Run("ParseAddress", func(t *testing.T) {
		t.Run("url", func(t *testing.T) {
			a := "http://img.baidu.com/avatar/abc.png"
			asset, err := ParseAddress(a)
			require.NoError(t, err)
			NewWithT(t).Expect(asset.String()).To(Equal(a))
		})
		t.Run("asset", func(t *testing.T) {
			a := "asset://avatar/abc/abc.png"
			asset, err := ParseAddress(a)
			require.NoError(t, err)
			NewWithT(t).Expect(asset.String()).To(Equal(a))
		})
		t.Run("asset, full", func(t *testing.T) {
			a := "asset://avatar/abc/2020/05/20/abc.png"
			asset, err := ParseAddress(a)
			require.NoError(t, err)
			NewWithT(t).Expect(asset.String()).To(Equal(a))
		})
		t.Run("asset no ext", func(t *testing.T) {
			a := "asset://image/1261327743736360967"
			asset, err := ParseAddress(a)
			require.NoError(t, err)
			NewWithT(t).Expect(asset.String()).To(Equal(a))
		})
		t.Run("asset no ext, full", func(t *testing.T) {
			a := "asset://image/1261327743736360967"
			asset, err := ParseAddress(a)
			require.NoError(t, err)
			NewWithT(t).Expect(asset.String()).To(Equal(a))
		})
	})
}
