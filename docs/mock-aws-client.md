# Testing AWS Code
### With Go, Interfaces and Mock Clients

We would like a strategy for unit testing our functions that does not require using the network to make API calls to AWS resources.

The [AWS SDK for Go](https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/welcome.html) offers a clear strategy: mock service clients.

Every service in the SDK exports a Go interface with all the client function signatures. Let's look at the DynamoDB interface:

```go
type DynamoDBAPI interface {
	CreateTableWithContext(aws.Context, *dynamodb.CreateTableInput, ...request.Option) (*dynamodb.CreateTableOutput, error)
	DeleteItemWithContext(aws.Context, *dynamodb.DeleteItemInput, ...request.Option) (*dynamodb.DeleteItemOutput, error)
	GetItemWithContext(aws.Context, *dynamodb.GetItemInput, ...request.Option) (*dynamodb.GetItemOutput, error)
	PutItemWithContext(aws.Context, *dynamodb.PutItemInput, ...request.Option) (*dynamodb.PutItemOutput, error)
	QueryWithContext(aws.Context, *dynamodb.QueryInput, ...request.Option) (*dynamodb.QueryOutput, error)
	UpdateItemWithContext(aws.Context, *dynamodb.UpdateItemInput, ...request.Option) (*dynamodb.UpdateItemOutput, error)
    ...
```
> From aws-sdk-go [dynamodbiface/interface.go](https://github.com/aws/aws-sdk-go/blob/master/service/dynamodb/dynamodbiface/interface.go)

There are close to 100 signatures in the interface that map to the 35+ DynamoDB API methods and utility functions for interacting with them.

The SDK satisfies the interface with methods that map to the real DynamoDB API. For example we can see how the `GetItemWithContext` takes a `dynamodb.GetItemInput` struct, crafts an HTTP POST with a `GetItem` header and a JSON body with the table and key name, and converts the HTTP response to a `dynamodb.GetItemOutput` struct.

> See aws-sdk-go [dynamodb/api](https://github.com/aws/aws-sdk-go/blob/master/service/dynamodb/api.go##L1745)

## Go Code -- Interface and Real Client

Our program interacts with a `DynamoDB` variable and methods like `PutItemWithContext` on it:

```go
func userPut(ctx context.Context, u *User) error {
	_, err := DynamoDB.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(u.ID),
			},
			"token": &dynamodb.AttributeValue{
				B: u.Token,
			},
			"username": &dynamodb.AttributeValue{
				S: aws.String(u.Username),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})

	return errors.WithStack(err)
}
```
> From [user.go](../user.go)


By default we initialize the `DynamoDB` variable as the SDK dynamodb client, with standard configuration for the AWS region and credentials. The only trick is that the variable type is the `DynamoDBAPI` interface which the SDK client satisfies.

```go
var DynamoDB = NewDynamoDB()

// DynamoDBAPI is a subset of dynamodbiface.DynamoDBAPI
type DynamoDBAPI interface {
	DeleteItemWithContext(ctx aws.Context, input *dynamodb.DeleteItemInput, opts ...request.Option) (*dynamodb.DeleteItemOutput, error)
	GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error)
	PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error)
}

// NewDynamoDB is an xray instrumented DynamoDB client
func NewDynamoDB() DynamoDBAPI {
	c := dynamodb.New(sess)
	xray.AWS(c.Client)
	return c
}
```
> From [aws.go](../aws.go)

## Go Code -- Interface and Mock Client

Note that we made our own `DynamoDBAPI` interface as a subset of the SDK `dynamodbiface.DynamoDBAPI`. This is so we don't have to write out 100s of methods for our test implementation. Now we can build a simple mock client that returns deterministic outputs:

```go
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
```
> From [aws_test.go](../aws_test.go)

Sometimes it is useful for the test implementation to process the input argument into the output variable. Here we make a mock KMS client that "encrypts" the input via Base64, a deterministic process that is easy to test.

```go
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
```
> From [aws_test.go](../aws_test.go)

Now in our test programs we can replace `DynamoDB`, `KMS`, etc. with our mock implementations. Our `UserCreate` function calls `DynamoDB.PutItemWithContext` and our mock client controls the output.

```go
func TestUserCreate(t *testing.T) {
	DynamoDB = &MockDynamoDB{
		GetItemOutput: &dynamodb.GetItemOutput{},
	}

	KMS = &MockKMS{}

	UUIDGen = func() uuid.UUID {
		return uuid.Must(uuid.FromString("26f0dc9f-4483-4b65-8724-3d1598ff6d14"))
	}

	r, err := UserCreate(context.Background(), events.APIGatewayProxyRequest{
		Body: `{"username": "test"}`,
	})
	assert.NoError(t, err)

	assert.EqualValues(t,
		events.APIGatewayProxyResponse{
			Body: "{\n  \"id\": \"26f0dc9f-4483-4b65-8724-3d1598ff6d14\",\n  \"username\": \"test\"\n}\n",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			StatusCode: 200,
		},
		r,
	)
}
```
> From [user_test.go](../user_test.go)

## Summary

The AWS SDK for Go offers a clear strategy for testing our code:

- Define an interface with the AWS SDK methods in use
- Use the AWS SDK clients as the interface implementation by default
- Build a mock client for the interface implementation for testing

We no longer have to worry about:

- Recording HTTP API request/response pairs
- Building a test API HTTP server

Go interfaces and the AWS SDK for Go make our software easy to build and test.