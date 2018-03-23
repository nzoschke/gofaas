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

var sess = session.Must(session.NewSession())

func init() {
	xray.Configure(xray.Config{
		LogLevel: "info",
	})
}

// DynamoDB is an xray instrumented DynamoDB client
func DynamoDB() *dynamodb.DynamoDB {
	c := dynamodb.New(sess)
	xray.AWS(c.Client)
	return c
}

// KMS is an xray instrumented KMS client
func KMS() *kms.KMS {
	c := kms.New(sess)
	xray.AWS(c.Client)
	return c
}

// Lambda is an xray instrumented Lambda client
func Lambda() *lambda.Lambda {
	c := lambda.New(sess)
	xray.AWS(c.Client)
	return c
}

// S3 is an xray instrumented S3 client
func S3() *s3.S3 {
	c := s3.New(sess)
	xray.AWS(c.Client)
	return c
}

// SNS is an xray instrumented SNS client
func SNS() *sns.SNS {
	c := sns.New(sess)
	xray.AWS(c.Client)
	return c
}
