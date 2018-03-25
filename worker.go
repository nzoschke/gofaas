package gofaas

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// WorkerEvent is foo
type WorkerEvent struct {
	SourceIP  string    `json:"source_ip"`
	TimeEnd   time.Time `json:"time_end"`
	TimeStart time.Time `json:"time_start"`
}

// WorkCreate invokes the worker func
func WorkCreate(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	r, _, err := JWTClaims(e, &jwt.StandardClaims{})
	if err != nil {
		return r, nil
	}

	out, err := Lambda.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(os.Getenv("WORKER_FUNCTION_NAME")),
		InvocationType: aws.String("Event"), // async
	})
	if err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	b, err := json.Marshal(out)
	if err != nil {
		return responseEmpty, errors.WithStack(err)
	}

	r.Body = string(b)
	return r, nil
}

// Worker is invoked directly to perform work then upload a report to S3
func Worker(ctx context.Context, e WorkerEvent) error {
	log.Printf("Worker Event: %+v\n", e)

	e.TimeEnd = time.Now()

	b, err := json.Marshal(e)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = S3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   bytes.NewReader(b),
		Bucket: aws.String(os.Getenv("BUCKET")),
		Key:    aws.String(uuid.NewV4().String()),
	})
	return errors.WithStack(err)
}

// WorkerPeriodic runs on a schedule to clean S3
func WorkerPeriodic(ctx context.Context, e events.CloudWatchEvent) error {
	log.Printf("WorkerPeriodic Event: %+v\n", e)

	iter := s3manager.NewDeleteListIterator(
		S3,
		&s3.ListObjectsInput{
			Bucket: aws.String(os.Getenv("BUCKET")),
		},
		iterWithContext(ctx),
	)

	err := s3manager.NewBatchDeleteWithClient(S3).Delete(ctx, iter)
	return errors.WithStack(err)
}

// iteratorSetContext sets context on the list request
// see https://github.com/aws/aws-sdk-go/issues/1790
func iterWithContext(ctx context.Context) func(*s3manager.DeleteListIterator) {
	return func(i *s3manager.DeleteListIterator) {
		nr := i.Paginator.NewRequest
		i.Paginator.NewRequest = func() (*request.Request, error) {
			req, err := nr()
			req.SetContext(ctx)
			return req, err
		}
	}
}
