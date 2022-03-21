package s3

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-courier/envconf"
	"github.com/stretchr/testify/require"
	"github.com/zj-open-source/helper/datatypes"
)

func TestS3Endpoint(t *testing.T) {
	endpoint, err := envconf.ParseEndpoint("s3://key:secret@s3.com/base")
	require.NoError(t, err)
	for _, db := range []*S3Endpoint{
		{
			Endpoint: *endpoint,
		},
	} {
		db.Init()
		t.Run("put object & get", func(t *testing.T) {
			ctx := context.Background()

			id, _ := idGen.ID()

			meta := ObjectMeta{
				Address: datatypes.Address{
					Group: "test",
					Key:   fmt.Sprint(id),
				},
				ContentType: "text/plain",
			}

			bf := bytes.NewBuffer([]byte(meta.Key()))
			err := db.PutObject(ctx, bf, meta)
			require.NoError(t, err)

			obj, err := db.StatsObject(ctx, meta.Address)
			require.NoError(t, err)

			publicURL := db.PublicURL(obj)
			t.Log(publicURL)

			protectURL, err := db.ProtectURL(ctx, obj, 1+time.Hour)
			require.NoError(t, err)
			t.Log(protectURL)

			buf := bytes.NewBuffer(nil)
			err = db.ReadObject(ctx, buf, meta.Address)
			require.NoError(t, err)

			t.Log(buf.String())
		})

		t.Run("presigned", func(t *testing.T) {
			ctx := context.Background()

			id, _ := idGen.ID()

			meta := ObjectMeta{
				Address: datatypes.Address{
					Group: "test",
					Key:   fmt.Sprint(id),
				},
				ContentType: "text/plain",
			}

			presignedURL, err := db.PresignedPutObject(ctx, meta.Address, 2*time.Hour)
			require.NoError(t, err)

			buf := bytes.NewBufferString(meta.Key())
			req, err := http.NewRequest("PUT", presignedURL, buf)
			require.NoError(t, err)
			req.Header.Set("Content-Type", "text/plain")

			resp, err := (&http.Client{}).Do(req)
			require.NoError(t, err)
			data, _ := httputil.DumpResponse(resp, true)
			fmt.Println(string(data))

			obj, err := db.StatsObject(ctx, meta.Address)
			require.NoError(t, err)

			fmt.Println(obj)

			b := bytes.NewBuffer(nil)
			err = db.ReadObject(ctx, b, meta.Address)
			require.NoError(t, err)

			t.Log(meta.Key())
			t.Log(buf.String())
		})

		t.Run("list object", func(t *testing.T) {
			ctx := context.Background()

			list, err := db.ListObjectByGroup(ctx, "app")
			require.NoError(t, err)
			spew.Dump(list)
		})
	}
}
