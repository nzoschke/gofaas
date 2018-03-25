package gofaas

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/kms"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// User represents a user
type User struct {
	ID         string `json:"id"`
	Token      []byte `json:"-"`
	TokenPlain string `json:"token,omitempty"`
	Username   string `json:"username"`
}

// UserCreate creates a user
func UserCreate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, _, err := JWTClaims(e, &jwt.StandardClaims{})
	if err != nil {
		return r, nil
	}

	u := &User{}
	if err := json.Unmarshal([]byte(e.Body), u); err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	u.ID = UUIDGen().String()
	u.TokenPlain = UUIDGen().String()

	if err := userPut(ctx, u); err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	return userResponse(u)
}

// UserDelete deletes a user by id
func UserDelete(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, _, err := JWTClaims(e, &jwt.StandardClaims{})
	if err != nil {
		return r, nil
	}

	u, err := userGet(ctx, e.PathParameters["id"], false)
	if err != nil {
		if err, ok := err.(ResponseError); ok {
			return err.Response()
		}
		return responseEmpty, errors.WithStack(err)
	}

	if err := userDelete(ctx, u); err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	return userResponse(u)
}

// UserRead returns a user by id
func UserRead(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, _, err := JWTClaims(e, &jwt.StandardClaims{})
	if err != nil {
		return r, nil
	}

	decrypt := false
	if e.QueryStringParameters["token"] == "true" {
		decrypt = true
	}

	u, err := userGet(ctx, e.PathParameters["id"], decrypt)
	if err != nil {
		if err, ok := err.(ResponseError); ok {
			return err.Response()
		}
		return responseEmpty, errors.WithStack(err)
	}

	return userResponse(u)
}

// UserUpdate updates a user by id
func UserUpdate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, _, err := JWTClaims(e, &jwt.StandardClaims{})
	if err != nil {
		return r, nil
	}

	nu := &User{}
	if err := json.Unmarshal([]byte(e.Body), nu); err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	u, err := userGet(ctx, e.PathParameters["id"], false)
	if err != nil {
		if err, ok := err.(ResponseError); ok {
			return err.Response()
		}
		return responseEmpty, errors.WithStack(err)
	}

	u.Username = nu.Username

	if err := userPut(ctx, u); err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	return userResponse(u)
}

func userGet(ctx context.Context, id string, decrypt bool) (*User, error) {
	out, err := DynamoDB.GetItemWithContext(ctx, &dynamodb.GetItemInput{
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
	if out.Item == nil {
		return nil, ResponseError{"not found", 404}
	}

	u := User{
		ID:       *out.Item["id"].S,
		Token:    out.Item["token"].B,
		Username: *out.Item["username"].S,
	}

	// optionally decrypt the token ciphertext
	if decrypt {
		out, err := KMS.DecryptWithContext(ctx, &kms.DecryptInput{
			CiphertextBlob: u.Token,
		})
		if err != nil {
			return nil, errors.WithStack(err)
		}
		u.TokenPlain = string(out.Plaintext)
	}

	return &u, nil
}

func userDelete(ctx context.Context, u *User) error {
	_, err := DynamoDB.DeleteItemWithContext(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": &dynamodb.AttributeValue{
				S: aws.String(u.ID),
			},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})

	return errors.WithStack(err)
}

func userPut(ctx context.Context, u *User) error {
	// encrypt a token plaintext if set
	if u.TokenPlain != "" {
		out, err := KMS.EncryptWithContext(ctx, &kms.EncryptInput{
			Plaintext: []byte(u.TokenPlain),
			KeyId:     aws.String(os.Getenv("KEY_ID")),
		})
		if err != nil {
			return errors.WithStack(err)
		}

		u.Token = out.CiphertextBlob
		u.TokenPlain = ""
	}

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

func userResponse(u *User) (events.APIGatewayProxyResponse, error) {
	b, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	return events.APIGatewayProxyResponse{
		Body: string(b) + "\n",
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		StatusCode: 200,
	}, nil
}
