# Development, Packaging and Deployment
### With Go, Lambda and SAM

The [Twelve-Factor App](https://12factor.net/) documents best practices for building software-as-a-service, most of which apply to our Go FaaS app. Two of the factors cover how to develop and deploy an app:

* [Build, release, run](https://12factor.net/build-release-run) -- Strictly separate build and run stages
* [Dev/prod parity](https://12factor.net/dev-prod-parity) -- Keep development, staging, and production as similar as possible

We can easily implement these these factors for our Go FaaS app with help of the [AWS SAM CLI](https://github.com/awslabs/aws-sam-cli) development environment.

## Build with go cross-compiler

Go is a compiled language, so a build phase is implicit in every development workflow.

A killer feature Go brings to the table is a cross compiler. Mac, Windows and any other operating system with the Go tool chain has the capability to build Linux binaries with a single command.

A killer feature Lambda brings is that it's input is a simple `.zip` file, a format with ubiquitous support.

So our build and packaging process is two commands that run on virtually any computer:

```console
$ GOOS=linux go build -o main
$ zip main.zip main
```

Compare this to the tools and services required to build Docker images, EC2 AMIs or Heroku slugs...

## Develop with SAM Local

But a new problem arises... How do run assemble these function packages into an API on our development box? Enter [AWS SAM CLI](https://github.com/awslabs/aws-sam-cli), a tool for local development and testing of serverless applications. It leverages Docker and the [lambci/lambda](https://hub.docker.com/r/lambci/lambda/) images to run the Go Linux binary.

### invoke

The simplest example is the `sam local invoke` command:

```console
$ echo '{}' | sam local invoke WorkerFunction
2018-05-11 08:18:21 Reading invoke payload from stdin (you can also pass it from file with --event)
2018-05-11 08:18:21 Invoking main (go1.x)
2018-05-11 08:18:21 Starting new HTTP connection (1): 169.254.169.254
Fetching lambci/lambda:go1.x Docker container image......
2018-05-11 08:18:24 Mounting handlers/worker as /var/task:ro inside runtime container
START RequestId: 358be5fc-928c-1a9a-b655-c84bb2958a5e Version: $LATEST
2018/02/24 23:27:04 Worker Event: {SourceIP: TimeEnd:0001-01-01 00:00:00 +0000 UTC TimeStart:0001-01-01 00:00:00 +0000 UTC}
END RequestId: 358be5fc-928c-1a9a-b655-c84bb2958a5e
REPORT RequestId: 358be5fc-928c-1a9a-b655-c84bb2958a5e  Duration: 524.34 ms  Billed Duration: 600 ms  Memory Size: 128 MB	Max Memory Used: 14 MB
```

This offers a fairly faithful representation of the Lambda production environment.

### start-api

We can also run assemble our HTTP functions with `sam local start-api`:

```console
$ sam local start-api

Mounting handler (go1.x) at http://127.0.0.1:3000/users [POST]
Mounting handler (go1.x) at http://127.0.0.1:3000/users/{id} [PUT]
Mounting handler (go1.x) at http://127.0.0.1:3000/users/{id} [DELETE]
Mounting handler (go1.x) at http://127.0.0.1:3000/ [GET]
Mounting handler (go1.x) at http://127.0.0.1:3000/users/{id} [GET]

## run `curl http://localhost:3000`

2018/02/24 15:41:57 Mounting handlers/dashboard as /var/task:ro inside runtime container
START RequestId: b913c432-3ed8-1e30-5c6d-e4582e59cb02 Version: $LATEST
END RequestId: b913c432-3ed8-1e30-5c6d-e4582e59cb02
REPORT RequestId: b913c432-3ed8-1e30-5c6d-e4582e59cb02  Duration: 2.12 ms  Billed Duration: 100 ms  Memory Size: 128 MB  Max Memory Used: 8 MB
```

This boots a local API Gateway that takes an HTTP request, invokes a function with the request event, and returns an HTTP response from the response event.

With a couple of first-party commands we can develop our app with a strong amount of dev/prod parity.

Compare this to the difference between developing an Rails app with `rails server` and deploying it to Elastic Beanstalk.

### make and watchexec

The final trick is to rebuild handler packages on every code change.

Thanks to the fact that the local API server mounts the handler directory on every request, and Go effectively caches builds, we have a simple solution: rebuild all handlers in parallel on every change.

We can use`make`, the ubiquitous tool for generating binaries from source files with [`watchexec`](https://github.com/mattgreen/watchexec), a program that watches for file changes and re-runs a command.

```console
$ watchexec -f '*.go' 'make -j handlers'
```
> From [Makefile](../Makefile)

Again we find a simple solution with great dev/prod parity.

## CloudFormation package and deploy

Our `template.yml` AWS config file is a CloudFormation template of the [SAM dialect](https://github.com/awslabs/serverless-application-model/blob/master/versions/2016-10-31.md). So we can lean on the `aws` CLI to release and run our app.

The release step is accomplished with the `sam package` command. This zips the handler directory and uploads the package to S3 and writes a new CloudFormation template with the S3 URLs. The run step is executed with the `sam deploy` command. This uses the CloudFormation API to update our resources -- such as updating Lambda functions to the new release.

```console
$ aws cloudformation package --output-template-file out.yml --s3-bucket $(BUCKET) --template-file template.yml
$ aws cloudformation deploy --capabilities CAPABILITY_NAMED_IAM --template-file out.yml --stack-name gofaas
```
> From [Makefile](../Makefile)

There's a lot of functionality baked into these commands like uploading package as efficiently as possible, creating new resources in dependency order, and safely rolling back updates on a failure. But it's all managed by CloudFormation.

The end result is glorious: a single config file that declares our entire app infrastructure, and a single command to deploy our app that generally takes less than a minute.

## Summary

An app with Go and SAM offers:

- Fast cross-compiled builds
- Single command release and run steps
- A development environment with strong production parity

We don't need to worry about:

- Dockerfile or docker-compose.yml files
- Code syncing
- Complex package formats
- Build services

Go and SAM makes it significantly easier to build and release applications with strong "dev/prod" parity.
