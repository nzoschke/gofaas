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
	ID       string `json:"id"`
	Token    string `json:"token,omitempty"`
	Username string `json:"username"`
}

// RE is an empty response
var RE = events.APIGatewayProxyResponse{}

// UserCreate creates a user
func UserCreate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	u := User{}
	if err := json.Unmarshal([]byte(e.Body), &u); err != nil {
		return RE, errors.WithStack(err)
	}

	u.ID = uuid.NewV4().String()

	out, err := KMS().EncryptWithContext(ctx, &kms.EncryptInput{
		Plaintext: []byte(uuid.NewV4().String()),
		KeyId:     aws.String(os.Getenv("KEY_ID")),
	})
	if err != nil {
		return RE, errors.WithStack(err)
	}

	_, err = DynamoDB().PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(u.ID),
			},
			"token": &dynamodb.AttributeValue{
				B: out.CiphertextBlob,
			},
			"username": &dynamodb.AttributeValue{
				S: aws.String(u.Username),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
	if err != nil {
		return RE, errors.WithStack(err)
	}

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

// UserRead returns a user by id
func UserRead(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	u, err := getUser(ctx, e.PathParameters["id"])
	if err != nil {
		return RE, nil
	}

	// mask token
	if e.QueryStringParameters["token"] != "true" {
		u.Token = ""
	}

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

// UserUpdate updates a user by id
func UserUpdate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	u, err := getUser(ctx, e.PathParameters["id"])
	if err != nil {
		return RE, nil
	}

	_, err = DynamoDB().PutItemWithContext(ctx, &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(u.ID),
			},
			"token": &dynamodb.AttributeValue{
				B: []byte(u.Token), // FIXME: pass through from getUser
			},
			"username": &dynamodb.AttributeValue{
				S: aws.String(u.Username),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
	if err != nil {
		return RE, errors.WithStack(err)
	}

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

func getUser(ctx context.Context, id string) (*User, error) {
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

	dout, err := KMS().DecryptWithContext(ctx, &kms.DecryptInput{
		CiphertextBlob: out.Item["token"].B,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	u := User{
		ID:       *out.Item["id"].S,
		Token:    string(dout.Plaintext),
		Username: *out.Item["username"].S,
	}

	return &u, nil
}
