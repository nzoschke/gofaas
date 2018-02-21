---
Title: Go HTTP Functions
Subtitle: For Lambda and API Gateway
---

An obvious application of a Go Lambda function is to handle an HTTP request. To accomplish this, we need the "serverless" API Gateway service to receive HTTP requests, translate that into an event, invoke our Lambda function, take it's return value, and turn it into an HTTP response. There's a lot of cool tech and optionality behind API Gateway service, but the promise of FaaS is that we don't have to worry about it. So lets jump straight to our function.

## Go Code

First, we write a HTTP request handler function, which couldn't be easier in Go:

```go
func Dashboard(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("Hello %s", e.RequestContext.Identity.SourceIP),
		StatusCode: 200,
	}, nil
}
```
> From [dashboard.go](http://github.com/nzoschke/gofaas/tree/dashboard.go)

The `APIGatewayProxyRequest` struct contains a user's HTTP request body, headers and metadata. The `APIGatewayProxyResponse` struct contains our HTTP response body, headers and status code.

This function is essentially error proof, but if it did return an error the API Gateway knows to respond to the user with a 500 status code.

## AWS Config

Next, we need to write a config file that tells AWS where this function fits into the cloud service architecture. Here we create a Go Lambda function and and connect it to a HTTP `GET /` route.

We are using the AWS Serverless Application Model (SAM), which makes this configuration file nice and simple. Behind the scenes this will be transformed to a config that creates an Application Gateway, Lambda function and proper IAM permissions.

```yaml
Resources:
    DashboardFunction:
        Properties:
            handler: handlers/dashboard/handler.zip
            runtime: go1.x
```
> From [template.yml](http://github.com/nzoschke/gofaas/tree/template.yml)

## Package, Develop and Deploy

Notice how the Handler properly points to a zip file? We need a bit more Go and Makefile glue to build the package.

So we need to write a Go **program** that Lambda can **invoke** to call our function. For this we use the Lambda SDK `Start` helper.

```go
func main() {
    lambda.Start(Dashboard)
}
```
> From [handlers/dashboard/main.go](http://github.com/nzoschke/gofaas/tree/dashboard.go)

Then we build a **package**, a zip file with a Linux binary that Lambda can run.

```shell
$ cd handlers/dashboard &&          \
    GOOS=linux go build -o handler  \
    zip handler handler.zip
```
> From [Makefile](http://github.com/nzoschke/gofaas/tree/Makefile)

Note how the Go cross-compiler makes it easy to build a Lambda package. This eliminates all cross-platform and dependency management challenges, and gives us a ~3 MB zip file we are confident we can deploy and execute quickly.



## Summary