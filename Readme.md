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

* Web handler
* Worker function
* Function-specific env and capabilities
* Database
* Periodic tasks
* Logs
* Tracing
* Notifications
* Go project layout
* One-command dev environment
* One-command deployment
* Deployment parameters
* Custom domain

What's remarkable is how little work is required to get all this. By standing on the shoulders of Go and AWS, all the undifferentiated heavy lifting is done. We just have to add our business logic functions.

We don't need a framework or a Platform-as-a-Service or even any 3rd party Software-as-a-Service to accomplish this. We need Go, AWS Lambda, and other AWS infrastructure services.

## Pre-reqs

This app uses [Go 1.10 beta](https://beta.golang.org/), [dep](https://github.com/golang/dep), [AWS CLI](https://aws.amazon.com/cli/), [AWS SAM Local](https://docs.aws.amazon.com/lambda/latest/dg/test-sam-local.html) and [Docker for Mac](https://www.docker.com/docker-mac).

```console
## install tools

$ brew install aws-cli
$ brew install go --devel
$ go get github.com/awslabs/aws-sam-local 
$ go get -u github.com/golang/dep/cmd/dep
```

<details>
<summary>We may want to double check the installed versions...</summary>
&nbsp;

```console
## check versions

$ aws --version
aws-cli/1.14.40 Python/3.6.4 Darwin/17.4.0 botocore/1.8.44

$ docker version
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

<details>
<summary>We may also want to configure the AWS CLI with IAM keys to develop and deploy our application...</summary>
&nbsp;

Follow the [Creating an IAM User in Your AWS Account](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html) doc to create a IAM user with programmatic access. Call the user `gofaas-admin` and attach the "Administrator Access" policy for now.

Then configure the CLI. Here we are creating a new profile that we can switch to with `export AWS_PROFILE=gofaas`. This will help us isolate our experiments from other AWS work.

```console
## configure the AWS CLI with keys

$ aws configure --profile gofaas
AWS Access Key ID [None]: AKIA................
AWS Secret Access Key [None]: PQN4CWZXXbJEgnrom2fP0Z+z................
Default region name [None]: us-east-1
Default output format [None]: json

## configure this session to use the profile

$ export AWS_PROFILE=gofaas

## verify the profile

$ aws iam get-user
{
    "User": {
        "Path": "/",
        "UserName": "gofaas-admin",
        "UserId": "AIDAJA44LJEOECDPZ3S5U",
        "Arn": "arn:aws:iam::572007530218:user/gofaas-admin",
        "CreateDate": "2018-02-16T16:17:24Z"
    }
}
```
</details>

## Get the app

We start by getting and testing the `github.com/nzoschke/gofaas`.

```console
## get the project

$ PKG=github.com/nzoschke/gofaas
$ go get $PKG && cd $GOPATH/src/$PKG

## verify tests pass

$ make test
...
ok  	github.com/nzoschke/gofaas	0.014s
```

This gives us confidence in our environment.

## Develop the app

```console
## build and start development server

$ make dev
cd ./handlers/dashboard && GOOS=linux go build...
2018/02/16 07:40:32 Fetching lambci/lambda:go1.x image for go1.x runtime...
Mounting handler (go1.x) at http://127.0.0.1:3000/ [GET]
```

```console
## request the app

$ curl http://localhost:3000
<html><body><h1>GoFAAS Dashboard</h1></body></html>
```

We may want to review all the SAM logs to better understand how our function is invoked...

<details>
<summary>Review all SAM logs to better understand function invocation...</summary>
&nbsp;

```console
$ make dev

aws-sam-local local start-api -n env.json
2018/02/16 08:24:33 Connected to Docker 1.35
2018/02/16 08:24:33 Fetching lambci/lambda:go1.x image for go1.x runtime...
go1.x: Pulling from lambci/lambda
Digest: sha256:d77adf847c45dcb5fae3cd93283447fad3f3d51ead024aed0c866a407a206e7c
Status: Image is up to date for lambci/lambda:go1.x

Mounting handler (go1.x) at http://127.0.0.1:3000/ [GET]

You can now browse to the above endpoints to invoke your functions.
You do not need to restart/reload SAM CLI while working on your functions,
changes will be reflected instantly/automatically. You only need to restart
SAM CLI if you update your AWS SAM template.

2018/02/16 08:24:37 Invoking handler (go1.x)
2018/02/16 08:24:37 Decompressing /Users/noah/go/src/github.com/nzoschke/gofaas/handlers/dashboard/handler.zip
2018/02/16 08:24:37 Mounting /private/var/folders/px/fd8j3qvn13gcxw9_nw25pphw0000gn/T/aws-sam-local-1518798277763101448 as /var/task:ro inside runtime container
START RequestId: 0619a836-ce3d-1819-8edc-2005395b83a6 Version: $LATEST
END RequestId: 0619a836-ce3d-1819-8edc-2005395b83a6
REPORT RequestId: 0619a836-ce3d-1819-8edc-2005395b83a6	Duration: 1.56 ms	Billed Duration: 100 ms	Memory Size: 128 MB	Max Memory Used: 5 MB
```
</details>

## Deploy the app

```console
$ make deploy
```