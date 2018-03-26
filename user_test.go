package gofaas

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

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

	u := User{}
	err = json.Unmarshal([]byte(r.Body), &u)
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
