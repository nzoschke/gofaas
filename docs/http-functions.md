# Go HTTP Functions
### With Lambda and API Gateway

An obvious application of a Go Lambda function is to handle an HTTP request. To accomplish this, we need the "serverless" API Gateway service to receive HTTP requests, translate that into an event, invoke our Lambda function, take it's return value, and turn it into an HTTP response. There's a lot of cool tech and options behind API Gateway service, but the promise of FaaS is that we don't have to worry about it. So lets jump straight to our Go function.

## Go Code -- Request Handler

First, we write a HTTP request handler function, which couldn't be easier in Go:

```go
import "github.com/aws/aws-lambda-go/events"

func Dashboard(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body: string("<html><body><h1>gofaas dashboard</h1></body></html>\n"),
		Headers: map[string]string{
			"Content-Type": "text/html",
		},
		StatusCode: 200,
	}, nil
}
```
> From [dashboard.go](../dashboard.go)

The `APIGatewayProxyRequest` struct contains a user's HTTP request body, headers and metadata. The `APIGatewayProxyResponse` struct contains our HTTP response body, headers and status code.

This function is essentially error proof, but if it did return an error the API Gateway knows to respond to the user with a `502 Bad Gateway` HTTP response.

## AWS Config

Next, we need to write a config file that tells AWS where this function fits into the cloud service architecture. Here we create a Go Lambda function and and connect it to a HTTP `GET /` route.

We are using the AWS Serverless Application Model (SAM), which makes this configuration file nice and simple. Behind the scenes this will be transformed to a config that creates an Application Gateway, Lambda function and proper IAM permissions.

```yaml
Resources:
  DashboardFunction:
    Properties:
      CodeUri: ./handlers/dashboard/main.zip
      Events:
        Request:
          Properties:
            Method: GET
            Path: /
          Type: Api
      Handler: main
      Runtime: go1.x
    Type: AWS::Serverless::Function
```
> From [template.yml](../template.yml)

## Package and Deploy

Notice how `CodeUri` points to a zip file? We need a bit more Go and `Makefile` glue to package the function.

So we need to write a Go **program** that Lambda will **invoke** to call our function. For this we use the Lambda SDK `Start` helper.

```go
package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/nzoschke/gofaas"
)

func main() {
	lambda.Start(gofaas.Dashboard)
}

```
> From [handlers/dashboard/main.go](handlers/dashboard/main.go)

Then we build a **package**, a zip file with a Linux binary that Lambda can run.

```console
$ cd handlers/dashboard &&          \
    GOOS=linux go build -o main  \
    zip main main.zip
```
> From [Makefile](../Makefile)

Note how the Go cross-compiler makes it easy to build a Lambda package. This eliminates all cross-platform and dependency management challenges, and gives us a ~3 MB zip file we are confident we can deploy and execute quickly.

Now we can deploy it:

```console
$ aws cloudformation package --output-template-file out.yml --s3-bucket $(BUCKET) --template-file template.yml
$ aws cloudformation deploy --capabilities CAPABILITY_NAMED_IAM --template-file out.yml --stack-name gofaas
```
> From [Makefile](../Makefile)

The `package` command uploads the zip file to S3 and writes a new template with the S3 URL. The `deploy` command creates or updates our Lambda function with the new package. In less than a minute we have a Go HTTP function online.

Finally we can call our function over HTTP:

```console
$ curl https://x19vpdk568.execute-api.us-east-1.amazonaws.com/Prod
<html><body><h1>gofaas dashboard</h1></body></html>
```

## Summary

Building and deploying an HTTP function with Go is fast and easy. We just have to:

- Write a Go func for a request and response event
- Write AWS config for Lambda and API Gateway

We no longer have to worry about:

- Application or infrastructure frameworks
- HTTP servers
- Build or runtime containers, instances or clusters
- Auto scaling
- Paying for idle servers

Go tools, Lambda and API Gateway make building HTTP services significantly easier.