# GoFAAS

Demo app that uses idiomatic Go and AWS.

## Motivation

I'm here to share a secret: Go in AWS Lambda is one of the best ways to write and run code.

Brandur recently wrote a great post: [Speed and Stability: Why Go is a Great Fit for Lambda](https://brandur.org/go-lambda). Having used Go in Lambda for years, I couldn't agree more. Running Go code on Lambda has resulted in systems that are the most cheap, fast, reliable, operational, and secure I have ever encountered.

Up until recently, this was only possible through hacks -- execution shims, heavy middleware, and no dev/prod parity.

That's because Go and AWS Lambda landscape is evolving very quickly. Offical Go support for Lambda was launched only a month ago. Go 1.10 is still in beta, and `dep` is under active development. The AWS Serverless Application Model (SAM) is in beta and hasn't got much attention yet.

However it's crystal clear that these techniques are the best practices.

So this project ties everything together. You can check it out and deploy it with a couple commands to get a feel for the tools. And you can fill in the blanks to turn it into your own web app.

It demonstrates:

* Go project layout
* Web handler
* Worker function
* Database
* Periodic tasks
* Tracing
* Notifications
* One-command dev environment
* One-command deployment

What's remarkable is how little work is required to get all this. By standing on the shoulders of Go and AWS, all the undifferentiated heavy lifting is done. We just have to add our business logic functions.

We don't need a framework or a Platform-as-a-Service or even any 3rd party Software-as-a-Service to accomplish this. We need Go, AWS Lambda, and other AWS infrastructure services.

## Pre-reqs

This app uses [Go 1.10 beta](https://beta.golang.org/), [dep](https://github.com/golang/dep), [AWS CLI](https://aws.amazon.com/cli/), [AWS SAM Local](https://docs.aws.amazon.com/lambda/latest/dg/test-sam-local.html) and [Docker for Mac](https://www.docker.com/docker-mac).

```console
$ brew install aws-cli
$ brew install go --devel
$ go get github.com/awslabs/aws-sam-local 
$ go get -u github.com/golang/dep/cmd/dep
```

You may want to verify versions match...

<details>

```console
## check versions
$ aws --version
aws-cli/1.14.40 Python/3.6.4 Darwin/17.4.0 botocore/1.8.44

Studio:gofaas noah$ docker version
Client:
 Version:	17.12.0-ce
 API version:	1.35
 Go version:	go1.9.2
 Git commit:	c97c6d6
 Built:	Wed Dec 27 20:03:51 2017
 OS/Arch:	darwin/amd64

Server:
 Engine:
  Version:	17.12.0-ce
  API version:	1.35 (minimum version 1.12)
  Go version:	go1.9.2
  Git commit:	c97c6d6
  Built:	Wed Dec 27 20:12:29 2017
  OS/Arch:	linux/amd64
  Experimental:	true

$ go version
go version go1.10rc2 darwin/amd64

$ aws-sam-local -v
sam version snapshot
```
</details>

## Get the project

```console
## get the project
$ PKG=github.com/nzoschke/gofaas
$ go get $PKG && cd $GOPATH/src/$PKG

## verify tests pass
$ make test
...
ok  	github.com/nzoschke/gofaas	0.014s
```
