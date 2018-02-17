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
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
)

// WorkerEvent is foo
type WorkerEvent struct {
	SourceIP  string    `json:"source_ip"`
	TimeEnd   time.Time `json:"time_end"`
	TimeStart time.Time `json:"time_start"`
}

// Worker is invoked directly to perform work then upload a report to S3
func Worker(ctx context.Context, e WorkerEvent) error {
	log.Printf("Worker Event: %+v\n", e)

	e.TimeEnd = time.Now()

	b, err := json.Marshal(e)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = S3().PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   bytes.NewReader(b),
		Bucket: aws.String(os.Getenv("BUCKET")),
		Key:    aws.String(uuid.NewV4().String()),
	})
	return errors.WithStack(err)
}

// WorkerPeriodic runs on a schedule to clean S3
func WorkerPeriodic(ctx context.Context, e events.CloudWatchEvent) error {
	log.Printf("WorkerPeriodic Event: %+v\n", e)

	// FIXME: s3manager isn't context / xray compatible
	c := s3.New(session.Must(session.NewSession()))

	iter := s3manager.NewDeleteListIterator(c, &s3.ListObjectsInput{
		Bucket: aws.String(os.Getenv("BUCKET")),
	})

	err := s3manager.NewBatchDeleteWithClient(c).Delete(ctx, iter)
	return errors.WithStack(err)
}
