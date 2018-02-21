package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nzoschke/gofaas"
)

func main() {
	lambda.Start(gofaas.NotifyAPIGateway(gofaas.UserCreate))
}
