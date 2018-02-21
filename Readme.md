# gofaas - Go Functions-as-a-Service

A project that demonstrates idiomatic Go with AWS Lambda and related "serverless" services.

## Motivation

Functions-as-a-Service (FaaS) like AWS Lambda are one of the latest advances in cloud Infrastructure-as-a-Service (IaaS). Go is particularly well-suited to run in Lambda due to its speed, size and cross-compiler. Check out the [Intro to Go Functions-as-a-Service and Lambda](docs/intro-go-faas.md) doc for more explaination.

For a long time, Go in Lambda was only possible through hacks -- execution shims, 3rd party frameworks and middleware, and no dev/prod parity. But in January 2018, [AWS launched official Go support for Lambda](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/) and [Go released 1.10](https://golang.org/doc/go1.10) paving the clearest path yet for us Gophers.

This project demonstrates a simple and clean foundation for Go in Lambda. You can clone and deploy it with a few commands to get a feel for the stack. Or you can fork and rework it to turn it into your own web application.

It demonstrates:

| Component                               | Via                                     | Status |
| --------------------------------------- |-----------------------------------------|--------|
| Web functions                           | Lambda, API Gateway                     |   ✓    |
| Worker functions (one-off and periodic) | Lambda, Invoke API, CloudWatch Events   |   ✓    |
| Packaging, development and deployment   | make, go, sam, CloudFormation (SAM)     |   ✓    |
| Per-function environment and policies   | Lambda, IAM                             |        |
| Custom domains                          | CloudFront, ACM                         |   ✓    |
| Logs, Tracing                           | CloudWatch Logs, X-Ray, AWS SDKs for Go |   ✓    |
| Notifications                           | SNS                                     |   ✓    |
| Databases                               | DynamoDB                                |        |
| Encryption                              | KMS                                     |        |

What's remarkable is how little work is required to get all this. By standing on the shoulders of Go and AWS, all the undifferentiated heavy lifting is done. We just have to add our business logic functions.

We don't need a framework, Platform-as-a-Service or even any 3rd party Software-as-a-Service to accomplish this. We need Go, an AWS account and a config file [like the one demonstrated here](template.yml).

## Quick Start

This project uses [Go 1.10](https://golang.org/), [dep](https://github.com/golang/dep), [AWS CLI](https://aws.amazon.com/cli/), [AWS SAM Local](https://docs.aws.amazon.com/lambda/latest/dg/test-sam-local.html) and [Docker for Mac](https://www.docker.com/docker-mac).

```console
## install tools

$ brew install aws-cli go
$ go get -u github.com/awslabs/aws-sam-local 
$ go get -u github.com/golang/dep/cmd/dep
```

<details>
<summary>We may want to double check the installed versions...</summary>
&nbsp;

```console
## check versions

$ aws --version
aws-cli/1.14.40 Python/3.6.4 Darwin/17.4.0 botocore/1.8.44

$ aws-sam-local -v
sam version snapshot

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
go version go1.10 darwin/amd64
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

### Get the app

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

This gives us confidence in our Go environment.

### Develop the app

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

## invoke the worker

$ echo '{}' | aws-sam-local local invoke WorkerFunction
2018/02/17 20:00:58 Worker Event: {SourceIP: TimeEnd:0001-01-01 00:00:00 +0000 UTC TimeStart:0001-01-01 00:00:00 +0000 UTC}
```

<details>
<summary>We may want to review all SAM logs to better understand function invocation...</summary>
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

2018/02/17 12:00:37 Reading invoke payload from stdin (you can also pass it from file with --event)
2018/02/17 12:00:37 Invoking handler (go1.x)
2018/02/17 12:00:37 Decompressing /Users/noah/go/src/github.com/nzoschke/gofaas/handlers/worker/handler.zip
2018/02/17 12:00:37 Mounting /private/var/folders/px/fd8j3qvn13gcxw9_nw25pphw0000gn/T/aws-sam-local-1518897637127351189 as /var/task:ro inside runtime container
START RequestId: 996f94f4-2fbe-16af-f33a-5e70e0199f35 Version: $LATEST
2018/02/17 20:00:58 Worker Event: {SourceIP: TimeEnd:0001-01-01 00:00:00 +0000 UTC TimeStart:0001-01-01 00:00:00 +0000 UTC}
END RequestId: 996f94f4-2fbe-16af-f33a-5e70e0199f35
REPORT RequestId: 996f94f4-2fbe-16af-f33a-5e70e0199f35	Duration: 486.90 ms	Billed Duration: 500 ms	Memory Size: 128 MB	Max Memory Used: 13 MB	
```
</details>

This gives us confidence in our development environment.

### Deploy the app

```console
## package and deploy the app

$ make deploy

make_bucket: pkgs-572007530218-us-east-1
Uploading to 59d2ea5b6bdf38fcbcf62236f4c26f21  3018471 / 3018471.0  (100.00%)
Waiting for changeset to be created
Waiting for stack create/update to complete
Successfully created/updated stack - gofaas
ApiUrl	https://x19vpdk568.execute-api.us-east-1.amazonaws.com/Prod

## request the app

$ curl https://x19vpdk568.execute-api.us-east-1.amazonaws.com/Prod
<html><body><h1>GoFAAS Dashboard</h1></body></html>
```

This gives us confidence in our production environment.

## Docs

Check out [the gofaas docs folder](docs/) where each component is explained in more details.

## Contributing

Find a bug or see a way to improve the project? [Open an issue](https://github.com/nzoschke/gofaas/issues).

## License

Apache 2.0 © 2018 Noah Zoschke