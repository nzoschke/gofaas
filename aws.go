package gofaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-xray-sdk-go/xray"
)

// AWS Clients that can be mocked for testing
var (
	APIGateway = NewAPIGateway()
	DynamoDB = NewDynamoDB()
	KMS      = NewKMS()
	Lambda   = NewLambda()
	S3       = NewS3()
	SNS      = NewSNS()

	sess = session.Must(session.NewSession())
)

func init() {
	xray.Configure(xray.Config{
		LogLevel: "info",
	})
}

// DynamoDBAPI is a subset of dynamodbiface.DynamoDBAPI
type DynamoDBAPI interface {
	DeleteItemWithContext(ctx aws.Context, input *dynamodb.DeleteItemInput, opts ...request.Option) (*dynamodb.DeleteItemOutput, error)
	GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error)
	PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error)
}

// KMSAPI is a subset of kmsiface.KMSAPI
type KMSAPI interface {
	DecryptWithContext(ctx aws.Context, input *kms.DecryptInput, opts ...request.Option) (*kms.DecryptOutput, error)
	EncryptWithContext(ctx aws.Context, input *kms.EncryptInput, opts ...request.Option) (*kms.EncryptOutput, error)
}

// NewAPIGateway is an xray instrumented APIGateway client
func NewAPIGateway() *apigateway.APIGateway {
	c := apigateway.New(sess)
	xray.AWS(c.Client)
	return c
}

// NewDynamoDB is an xray instrumented DynamoDB client
func NewDynamoDB() DynamoDBAPI {
	c := dynamodb.New(sess)
	xray.AWS(c.Client)
	return c
}

// NewKMS is an xray instrumented KMS client
func NewKMS() KMSAPI {
	c := kms.New(sess)
	xray.AWS(c.Client)
	return c
}

// NewLambda is an xray instrumented Lambda client
func NewLambda() *lambda.Lambda {
	c := lambda.New(sess)
	xray.AWS(c.Client)
	return c
}

// NewS3 is an xray instrumented S3 client
func NewS3() *s3.S3 {
	c := s3.New(sess)
	xray.AWS(c.Client)
	return c
}

// NewSNS is an xray instrumented SNS client
func NewSNS() *sns.SNS {
	c := sns.New(sess)
	xray.AWS(c.Client)
	return c
}
