package gofaas

import (
	"encoding/base64"

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
type MockKMS struct{}

func (m *MockKMS) DecryptWithContext(ctx aws.Context, input *kms.DecryptInput, opts ...request.Option) (*kms.DecryptOutput, error) {
	s, _ := base64.StdEncoding.DecodeString(string(input.CiphertextBlob))
	return &kms.DecryptOutput{
		Plaintext: s,
	}, nil
}

func (m *MockKMS) EncryptWithContext(ctx aws.Context, input *kms.EncryptInput, opts ...request.Option) (*kms.EncryptOutput, error) {
	return &kms.EncryptOutput{
		CiphertextBlob: []byte(base64.StdEncoding.EncodeToString(input.Plaintext)),
	}, nil
}
