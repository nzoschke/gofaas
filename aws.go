package gofaas

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-xray-sdk-go/xray"
)

func init() {
	xray.Configure(xray.Config{
		LogLevel: "info",
	})
}

// DynamoDB is an xray instrumented DynamoDB client
func DynamoDB() *dynamodb.DynamoDB {
	c := dynamodb.New(session.Must(session.NewSession()))
	xray.AWS(c.Client)
	return c
}

// KMS is an xray instrumented KMS client
func KMS() *kms.KMS {
	c := kms.New(session.Must(session.NewSession()))
	xray.AWS(c.Client)
	return c
}

// Lambda is an xray instrumented Lambda client
func Lambda() *lambda.Lambda {
	c := lambda.New(session.Must(session.NewSession()))
	xray.AWS(c.Client)
	return c
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
