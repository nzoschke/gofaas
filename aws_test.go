package gofaas

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
)

// MockDynamoDB is a mock DynamoDBAPI implementation
type MockDynamoDB struct {
	DeleteItemOutput *dynamodb.DeleteItemOutput
	GetItemOutput    *dynamodb.GetItemOutput
	PutItemOutput    *dynamodb.PutItemOutput
}

func (m *MockDynamoDB) DeleteItemWithContext(ctx aws.Context, input *dynamodb.DeleteItemInput, opts ...request.Option) (*dynamodb.DeleteItemOutput, error) {
	return m.DeleteItemOutput, nil
}

func (m *MockDynamoDB) GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error) {
	return m.GetItemOutput, nil
}

func (m *MockDynamoDB) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	return m.PutItemOutput, nil
}

// MockKMS is a mock KMSAPI implementation
type MockKMS struct {
	DecryptOutput *kms.DecryptOutput
	EncryptOutput *kms.EncryptOutput
}

func (m *MockKMS) DecryptWithContext(ctx aws.Context, input *kms.DecryptInput, opts ...request.Option) (*kms.DecryptOutput, error) {
	return m.DecryptOutput, nil
}

func (m *MockKMS) EncryptWithContext(ctx aws.Context, input *kms.EncryptInput, opts ...request.Option) (*kms.EncryptOutput, error) {
	return m.EncryptOutput, nil
}
