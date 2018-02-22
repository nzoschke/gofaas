package gofaas

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// User represents a user
type User struct {
	ID         string `json:"id"`
	Token      []byte `json:"-"`
	TokenPlain string `json:"token,omitempty"`
	Username   string `json:"username"`
}

// RE is an empty response
var RE = events.APIGatewayProxyResponse{}

// UserCreate creates a user
func UserCreate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	u := &User{}
	if err := json.Unmarshal([]byte(e.Body), u); err != nil {
		return RE, errors.WithStack(err)
	}

	u.ID = uuid.NewV4().String()
	u.TokenPlain = uuid.NewV4().String()

	if err := userPut(ctx, u); err != nil {
		return RE, errors.WithStack(err)
	}

	return userResponse(u)
}

// UserRead returns a user by id
func UserRead(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	decrypt := false
	if e.QueryStringParameters["token"] == "true" {
		decrypt = true
	}

	u, err := userGet(ctx, e.PathParameters["id"], decrypt)
	if err != nil {
		return RE, errors.WithStack(err)
	}

	return userResponse(u)
}

// UserUpdate updates a user by id
func UserUpdate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	nu := &User{}
	if err := json.Unmarshal([]byte(e.Body), nu); err != nil {
		return RE, errors.WithStack(err)
	}

	u, err := userGet(ctx, e.PathParameters["id"], false)
	if err != nil {
		return RE, errors.WithStack(err)
	}

	u.Username = nu.Username

	if err := userPut(ctx, u); err != nil {
		return RE, errors.WithStack(err)
	}

	return userResponse(u)
}

func userGet(ctx context.Context, id string, decrypt bool) (*User, error) {
	out, err := DynamoDB().GetItemWithContext(ctx, &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(id),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u := User{
		ID:       *out.Item["id"].S,
		Token:    out.Item["token"].B,
		Username: *out.Item["username"].S,
	}

	// optionally decrypt the token ciphertext
	if decrypt {
		out, err := KMS().DecryptWithContext(ctx, &kms.DecryptInput{
			CiphertextBlob: u.Token,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		u.TokenPlain = string(out.Plaintext)
	}

	return &u, nil
}

func userPut(ctx context.Context, u *User) error {
	// encrypt a token plaintext if set
	if u.TokenPlain != "" {
		out, err := KMS().EncryptWithContext(ctx, &kms.EncryptInput{
			Plaintext: []byte(u.TokenPlain),
			KeyId:     aws.String(os.Getenv("KEY_ID")),
		})
		if err != nil {
			return err
		}

		u.Token = out.CiphertextBlob
		u.TokenPlain = ""
	}

	_, err := DynamoDB().PutItemWithContext(ctx, &dynamodb.PutItemInput{
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
	return err
}

func userResponse(u *User) (events.APIGatewayProxyResponse, error) {
	b, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return RE, errors.WithStack(err)
	}

	return events.APIGatewayProxyResponse{
		Body: string(b) + "\n",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		StatusCode: 200,
	}, nil
}
