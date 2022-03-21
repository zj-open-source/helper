package s3

import (
	"github.com/go-courier/envconf"
)

type S3Endpoint struct {
	*ObjectDB `env:"-"`
	Endpoint  envconf.Endpoint `env:""`
}

func (s *S3Endpoint) Init() {
	secure := true

	if s.Endpoint.Extra.Get("secure") == "false" || s.Endpoint.Extra.Get("secure") == "FALSE" {
		secure = false
	}

	s.ObjectDB = &ObjectDB{
		Endpoint:        s.Endpoint.Host(),
		AccessKeyID:     s.Endpoint.Username,
		SecretAccessKey: envconf.Password(s.Endpoint.Password),
		BucketName:      s.Endpoint.Base,
		Secure:          secure,
	}
}
