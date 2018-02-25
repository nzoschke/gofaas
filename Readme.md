# Go Functions-as-a-Service

Running a Go app on AWS Lambda is easier than ever, once you figure out how to configure Lambda, API Gateway and 10 or other "serverless" services to support the Go functions.

This is a boiilerplate app with all the AWS pieces configured correctly (see the [CloudFormation template](template.yml)) and explained in depth (see the [docs folder](docs/)). With this foundation you can skip over all the undifferentiated work and focus entirely on your Go project.

## Motivation

Functions-as-a-Service (FaaS) like AWS Lambda are one of the latest advances in cloud Infrastructure-as-a-Service (IaaS). Go is particularly well-suited to run in Lambda due to its speed, size and cross-compiler. Check out the [Intro to Go Functions-as-a-Service and Lambda](docs/intro-go-faas.md) doc for more explaination.

For a long time, Go in Lambda was only possible through hacks -- execution shims, 3rd party frameworks and middleware, and little dev/prod parity. But in January 2018, [AWS launched official Go support for Lambda](https://aws.amazon.com/blogs/compute/announcing-go-support-for-aws-lambda/) and [Go released v1.10](https://golang.org/doc/go1.10) paving the clearest path yet for us Gophers.

This project demonstrates a simple and clean foundation for Go in Lambda. You can clone and deploy it with a few commands to get a feel for the stack. Or you can fork and rework it to turn it into your own web application.

It demonstrates:

| Component                                    | Via                                     | Config, Code       |
| -------------------------------------------- |-----------------------------------------|:------------------:|
| [HTTP functions][1]                          | Lambda, API Gateway                     | [💾](dashboard.go) |
| [Worker functions (one-off and periodic)][2] | Lambda, Invoke API, CloudWatch Events   | [💾](worker.go)    |
| [Development, packaging and deployment][3]   | make, go, aws-sam-local, CloudFormation | [⚙️](Makefile)     |
| [Per-function environment and policies][4]   | Lambda, IAM                             | [⚙️](template.yml) |
| [Custom domains][5]                          | CloudFront, ACM                         | [⚙️](template.yml) |
| [Logs, Tracing][6]                           | CloudWatch Logs, X-Ray, AWS SDKs for Go | [💾](aws.go)       |
| [Notifications][7]                           | SNS                                     | [💾](notify.go)    |
| [Databases and encryption at rest][8]        | DynamoDB, KMS                           | [💾](user.go)      |

[1]: docs/http-functions.md
[2]: docs/worker-functions.md
[3]: docs/dev-package-deploy.md
[4]: docs/per-function-policies.md
[5]: docs/custom-domains.md
[6]: docs/logs-tracing.md
[7]: docs/notifications.md
[8]: docs/databases-encryption.md

What's remarkable is how little work is required to get all functionality for our app. We don't need a framework, Platform-as-a-Service, or even any 3rd party Software-as-a-Service. And yes, we don't need servers. By standing on the shoulders of Go and AWS, all the undifferentiated heavy lifting is handled.

We just need a good [CloudFormation config file](template.yml) and a simple [Makefile](Makefile), then we can focus entirely on writing our Go functions.

## Quick Start

This project uses :

- [AWS CLI](https://aws.amazon.com/cli/)
- [AWS SAM Local](https://docs.aws.amazon.com/lambda/latest/dg/test-sam-local.html)
- [Docker CE](https://www.docker.com/community-edition)
- [Go 1.10](https://golang.org/)
- [watchexec](https://github.com/mattgreen/watchexec)

Install the CLI tools and Docker CE

```console
$ brew install aws-cli go watchexec
$ go get -u github.com/awslabs/aws-sam-local
$ open https://store.docker.com/search?type=edition&offering=community
```

<details>
<summary>We may want to double check the installed versions...</summary>
&nbsp;

```console
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

$ watchexec --version
watchexec 1.8.6
```
</details>

<details>
<summary>We may also want to configure the AWS CLI with IAM keys to develop and deploy our application...</summary>
&nbsp;

Follow the [Creating an IAM User in Your AWS Account](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_users_create.html) doc to create a IAM user with programmatic access. Call the user `gofaas-admin` and attach the "Administrator Access" policy for now.

Then configure the CLI. Here we are creating a new profile that we can switch to with `export AWS_PROFILE=gofaas`. This will help us isolate our experiments from other AWS work.

Configure an AWS profile with keys and switch to the profile:

```console
$ aws configure --profile gofaas
AWS Access Key ID [None]: AKIA................
AWS Secret Access Key [None]: PQN4CWZXXbJEgnrom2fP0Z+z................
Default region name [None]: us-east-1
Default output format [None]: json

$ export AWS_PROFILE=gofaas
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

### Get the App

We start by getting and testing the `github.com/nzoschke/gofaas`.

```console
$ PKG=github.com/nzoschke/gofaas
$ go get $PKG && cd $GOPATH/src/$PKG

$ make test
...
ok  	github.com/nzoschke/gofaas	0.014s
```

This gives us confidence in our Go environment.

### Develop the App

We can then build the app and start a development server:

```console
$ make dev
cd ./handlers/dashboard && GOOS=linux go build...
2018/02/25 08:03:12 Connected to Docker 1.35
2018/02/16 07:40:32 Fetching lambci/lambda:go1.x image for go1.x runtime...

Mounting handler (go1.x) at http://127.0.0.1:3000/users/{id} [DELETE]
Mounting handler (go1.x) at http://127.0.0.1:3000/users/{id} [PUT]
Mounting handler (go1.x) at http://127.0.0.1:3000/users/{id} [GET]
Mounting handler (go1.x) at http://127.0.0.1:3000/ [GET]
Mounting handler (go1.x) at http://127.0.0.1:3000/users [POST]
```

Now we can access our HTTP functions on port 3000:

```console
$ curl http://localhost:3000
<html><body><h1>gofaas dashboard</h1></body></html>
```

We can also invoke a function directly:

```
$ echo '{}' | aws-sam-local local invoke WorkerFunction
...
START RequestId: 36d6d40e-0d4b-168c-63d5-76b25f543d21 Version: $LATEST
2018/02/25 16:05:21 Worker Event: {SourceIP: TimeEnd:0001-01-01 00:00:00 +0000 UTC TimeStart:0001-01-01 00:00:00 +0000 UTC}
END RequestId: 36d6d40e-0d4b-168c-63d5-76b25f543d21
REPORT RequestId: 36d6d40e-0d4b-168c-63d5-76b25f543d21	Duration: 681.67 ms	Billed Duration: 700 ms	Memory Size: 128 MB	Max Memory Used: 14 MB
```

Note: if you see `No AWS credentials found. Missing credentials may lead to slow startup...`, review `aws configure list` and your `AWS_PROFILE` env var.

This gives us confidence in our development environment.

### Deploy the App

Now we can package and deploy the app:

```console
$ make deploy
make_bucket: pkgs-572007530218-us-east-1
Uploading to 59d2ea5b6bdf38fcbcf62236f4c26f21  3018471 / 3018471.0  (100.00%)
Waiting for changeset to be created
Waiting for stack create/update to complete
Successfully created/updated stack - gofaas

ApiUrl	https://x19vpdk568.execute-api.us-east-1.amazonaws.com/Prod
```

Now we can access our HTTP functions on AWS:


```console
$ curl https://x19vpdk568.execute-api.us-east-1.amazonaws.com/Prod
<html><body><h1>gofaas dashboard</h1></body></html>
```

We can also invoke a function directly:

```
$ aws lambda invoke --function-name gofaas-WorkerFunction --log-type Tail --output text --query 'LogResult' out.log | base64 -D
START RequestId: 0bb47628-1718-11e8-ad73-c58e72b8826c Version: $LATEST
2018/02/21 15:01:07 Worker Event: {SourceIP: TimeEnd:0001-01-01 00:00:00 +0000 UTC TimeStart:0001-01-01 00:00:00 +0000 UTC}
END RequestId: 0bb47628-1718-11e8-ad73-c58e72b8826c
REPORT RequestId: 0bb47628-1718-11e8-ad73-c58e72b8826c	Duration: 11.11 ms	Billed Duration: 100 ms 	Memory Size: 128 MB	Max Memory Used: 41 MB
```

Look at that speedy 11 ms duration! Go is faster than the minimum billing duration of 100 ms.

This gives us confidence in our production environment.

### Development Environment

If we want to work on the [worker](docs/worker-functions.md) or [database](docs/databases.md) functions locally, we need to give the functions environment variables with pointers to, DynamoDB, KMS and S3. Open up `env.json` and set `BUCKET`, etc. with the ids of the resources we just created on deploy:

```console
$ aws cloudformation describe-stack-resources --output text --stack-name gofaas \
  --query 'StackResources[*].{Name:LogicalResourceId,Id:PhysicalResourceId,Type:ResourceType}' | \
  grep 'Bucket\|Key\|UsersTable'

gofaas-bucket-aykdokk6aek8            Bucket      AWS::S3::Bucket
8eb8e209-51fb-41fa-adfe-1ec401667df4  Key         AWS::KMS::Key
gofaas-UsersTable-1CYAQH3HHHRGW       UsersTable  AWS::DynamoDB::Table
```

## Docs

Check out [the gofaas docs folder](docs/) where each component is explained in more details.

## Contributing

Find a bug or see a way to improve the project? [Open an issue](https://github.com/nzoschke/gofaas/issues).

## License

Apache 2.0 © 2018 Noah Zoschke