package gofaas

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

// WorkerEvent is foo
type WorkerEvent struct {
	SourceIP  string    `json:"source_ip"`
	TimeEnd   time.Time `json:"time_end"`
	TimeStart time.Time `json:"time_start"`
}

// Worker performs a task then uploads a report to S3
func Worker(ctx context.Context, e WorkerEvent) error {
	e.TimeEnd = time.Now()

	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = S3().PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   bytes.NewReader(b),
		Bucket: aws.String(os.Getenv("BUCKET")),
		Key:    aws.String(uuid.NewV4().String()),
	})
	return err
}
