package gofaas

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

var (
	responseEmpty = events.APIGatewayProxyResponse{}
)

// ResponseError is an error type that indicates a non-200 response
type ResponseError struct {
	Body       string
	StatusCode int
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("%s (%d)", e.Body, e.StatusCode)
}

// Response returns an API Gateway Response event
func (e ResponseError) Response() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("{%q: %q}\n", "error", e.Body),
		StatusCode: e.StatusCode,
	}, nil
}
