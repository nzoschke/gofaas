package gofaas

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func init() {
	xray.Configure(xray.Config{
		LogLevel: "info",
	})
}

// S3 is an xray instrumented S3 client
func S3() *s3.S3 {
	c := s3.New(session.Must(session.NewSession()))
	xray.AWS(c.Client)
	return c
}

// SNS is an xray instrumented SNS client
func SNS() *sns.SNS {
	c := sns.New(session.Must(session.NewSession()))
	xray.AWS(c.Client)
	return c
}
