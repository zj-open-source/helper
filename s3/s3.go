package s3

import (
	"context"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/go-courier/envconf"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zj-open-source/tools/datatypes"
)

var (
	ErrUnknownContentType = errors.New("unknown content type")
	ErrInvalidObject      = errors.New("invalid object key")
)

type ObjectDB struct {
	Endpoint        string                                                             `env:",upstream"`
	AccessKeyID     string                                                             `env:""`
	SecretAccessKey envconf.Password                                                   `env:""`
	BucketName      string                                                             `env:""`
	Secure          bool                                                               `env:""`
	PresignedValues func(db *ObjectDB, key string, expiresIn time.Duration) url.Values `env:"-"`
}

func (db *ObjectDB) LivenessCheck() map[string]string {
	key := db.BucketName + "." + db.Endpoint
	m := map[string]string{
		key: "ok",
	}

	c, err := db.Client()

	if err != nil {
		m[key] = err.Error()
	} else {
		if _, err := c.GetBucketLocation(context.Background(), db.BucketName); err != nil {
			m[key] = err.Error()
		}
	}

	return m
}

func (db *ObjectDB) Client() (*minio.Client, error) {
	client, err := minio.New(db.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(db.AccessKeyID, db.SecretAccessKey.String(), ""),
		Secure: db.Secure,
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (db *ObjectDB) PublicURL(meta *ObjectMeta) *url.URL {
	u := &url.URL{}
	u.Scheme = "http"
	if db.Secure {
		u.Scheme += "s"
	}

	u.Host = db.Endpoint
	u.Path = db.BucketName + "/" + meta.Key()
	return u
}

func (db *ObjectDB) ProtectURL(ctx context.Context, meta *ObjectMeta, expiresIn time.Duration) (*url.URL, error) {
	c, err := db.Client()
	if err != nil {
		return nil, err
	}
	values := url.Values{}
	if db.PresignedValues != nil {
		values = db.PresignedValues(db, meta.Key(), expiresIn)
	}

	u, err := c.PresignedGetObject(ctx, db.BucketName, meta.Key(), expiresIn, values)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (db *ObjectDB) PutObject(ctx context.Context, fileReader io.Reader, meta ObjectMeta) error {
	if ctx == nil {
		ctx = context.Background()
	}

	c, err := db.Client()
	if err != nil {
		return err
	}

	if meta.Size == 0 {
		if canLen, ok := fileReader.(interface{ Len() int }); ok {
			meta.Size = int64(canLen.Len())
		}
	}

	_, err = c.PutObject(ctx, db.BucketName, meta.Key(), fileReader, meta.Size, minio.PutObjectOptions{
		ContentType: meta.ContentType,
	})

	return err
}

func (db *ObjectDB) ReadObject(ctx context.Context, writer io.Writer, address datatypes.Address) error {
	c, err := db.Client()
	if err != nil {
		return err
	}

	object, err := c.GetObject(ctx, db.BucketName, (&ObjectMeta{
		Address: address,
	}).Key(), minio.GetObjectOptions{})
	if err != nil {
		return err
	}
	defer object.Close()

	_, err = io.Copy(writer, object)
	if err != nil {
		return err
	}

	return err
}

func (db *ObjectDB) PresignedPutObject(ctx context.Context, address datatypes.Address, expiresIn time.Duration) (string, error) {
	c, err := db.Client()
	if err != nil {
		return "", err
	}
	presignedURL, err := c.PresignedPutObject(ctx, db.BucketName, (&ObjectMeta{
		Address: address,
	}).Key(), expiresIn)
	if err != nil {
		return "", err
	}
	presignedURL.Scheme = "http"
	if db.Secure {
		presignedURL.Scheme += "s"
	}
	return presignedURL.String(), nil
}

func (db *ObjectDB) DeleteObject(ctx context.Context, address datatypes.Address) error {
	c, err := db.Client()
	if err != nil {
		return err
	}

	return c.RemoveObject(ctx, db.BucketName, (&ObjectMeta{
		Address: address,
	}).Key(), minio.RemoveObjectOptions{})
}

func (db *ObjectDB) StatsObject(ctx context.Context, address datatypes.Address) (*ObjectMeta, error) {
	c, err := db.Client()
	if err != nil {
		return nil, err
	}

	object, err := c.GetObject(ctx, db.BucketName, (&ObjectMeta{
		Address: address,
	}).Key(), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	info, err := object.Stat()
	if err != nil {
		return nil, err
	}

	om := &ObjectMeta{
		Address: address,
	}

	om.ContentType = info.ContentType
	om.ETag = info.ETag
	om.Size = info.Size

	return om, err
}

func (db *ObjectDB) ListObjectByGroup(ctx context.Context, group string) ([]*ObjectMeta, error) {
	c, err := db.Client()
	if err != nil {
		return nil, err
	}

	metas := make([]*ObjectMeta, 0)

	for obj := range c.ListObjects(ctx, db.BucketName, minio.ListObjectsOptions{
		Prefix:    group,
		Recursive: true,
	}) {
		om, err := ParseObjectMetaFromKey(obj.Key)
		if err != nil {
			continue
		}

		om.ContentType = obj.ContentType
		om.ETag = obj.ETag
		om.Size = obj.Size

		metas = append(metas, om)
	}

	return metas, nil
}

func (db *ObjectDB) CopyObject(ctx context.Context, dstAddress datatypes.Address, srcBucketName string, srcAddress datatypes.Address) (*ObjectMeta, error) {
	c, err := db.Client()
	if err != nil {
		return nil, err
	}

	// Source object
	srcOpts := minio.CopySrcOptions{
		Bucket: srcBucketName,
		Object: (&ObjectMeta{
			Address: srcAddress,
		}).Key(),
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket: db.BucketName,
		Object: (&ObjectMeta{
			Address: dstAddress,
		}).Key(),
	}

	// Copy object call
	uploadInfo, err := c.CopyObject(ctx, dstOpts, srcOpts)
	if err != nil {
		return nil, err
	}

	om := &ObjectMeta{
		Address: dstAddress,
	}

	om.ETag = uploadInfo.ETag
	om.Size = uploadInfo.Size
	return om, nil
}

func ParseObjectMetaFromKey(s3Key string) (*ObjectMeta, error) {
	var (
		group, key, ext string
	)
	parts := strings.SplitN(s3Key, "/", 2)
	if len(parts) != 2 {
		return nil, ErrInvalidObject
	}
	group = parts[0]
	key = parts[1]

	keyParts := strings.Split(key, ".")
	if len(keyParts) > 1 {
		ext = keyParts[len(keyParts)-1]
		key = key[:len(key)-len(ext)-1]
	}

	address := datatypes.Address{
		Group: group,
		Key:   key,
		Ext:   ext,
	}

	return &ObjectMeta{
		Address: address,
	}, nil
}

type ObjectMeta struct {
	Address     datatypes.Address `json:"address"`
	Size        int64             `json:"size"`
	ContentType string            `json:"contentType"`
	ETag        string            `json:"etag"`
}

func (meta ObjectMeta) Key() string {
	return meta.Address.Path()
}
