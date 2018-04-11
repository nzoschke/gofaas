package gofaas

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
)

// HandlerAPIGateway is an API Gateway Proxy Request handler function
type HandlerAPIGateway func(context.Context, events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// HandlerCloudWatch is a CloudWatchEvent handler function
type HandlerCloudWatch func(context.Context, events.CloudWatchEvent) error

// HandlerWorker is a Worker handler function
type HandlerWorker func(context.Context, WorkerEvent) error

// NotifyAPIGateway wraps a handler func and sends an SNS notification on error
func NotifyAPIGateway(h HandlerAPIGateway) HandlerAPIGateway {
	return func(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		r, err := h(ctx, e)
		notify(ctx, err)
		return r, err
	}
}

// NotifyCloudWatch wraps a handler func and sends an SNS notification on error
func NotifyCloudWatch(h HandlerCloudWatch) HandlerCloudWatch {
	return func(ctx context.Context, e events.CloudWatchEvent) error {
		err := h(ctx, e)
		notify(ctx, err)
		return err
	}
}

// NotifyWorker wraps a handler func and sends an SNS notification on error
func NotifyWorker(h HandlerWorker) HandlerWorker {
	return func(ctx context.Context, e WorkerEvent) error {
		err := h(ctx, e)
		notify(ctx, err)
		return err
	}
}

func notify(ctx context.Context, err error) {
	if err == nil {
		return
	}

	subj := fmt.Sprintf("ERROR %s", os.Getenv("AWS_LAMBDA_FUNCTION_NAME"))
	msg := fmt.Sprintf("%+v\n", err)
	log.Printf("%s %s\n", subj, msg)

	topic := os.Getenv("NOTIFICATION_TOPIC")
	if topic == "" {
		return
	}

	_, err = SNS.PublishWithContext(ctx, &sns.PublishInput{
		Message:  aws.String(msg),
		Subject:  aws.String(subj),
		TopicArn: aws.String(topic),
	})
	if err != nil {
		log.Printf("NotifyError SNS Publish error %+v\n", err)
	}
}
