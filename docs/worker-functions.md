# Go Worker Functions
### With Lambda, S3 and CloudWatch Events

Lambda isn't just for HTTP functions. Another application of a Go Lambda function is one that we will invoke manually or automatically to do some work. To accomplish this we need something to work against such as an S3 bucket, and the the Lambda Invoke API or [CloudWatch Events](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html) to trigger our worker.

## Go Code

First, we write a function that will do some work then save a report to S3. We anticipate invoking this manually with custom data like a requestor and a start time, so we define a custom event.

```go
import "github.com/aws/aws-sdk-go/service/s3"

var S3 = s3.New(session.Must(session.NewSession()))

type WorkerEvent struct {
	SourceIP  string    `json:"source_ip"`
	TimeEnd   time.Time `json:"time_end"`
	TimeStart time.Time `json:"time_start"`
}

func Worker(ctx context.Context, e WorkerEvent) error {
	// perform work here

	e.TimeEnd = time.Now()
	b, err := json.Marshal(e)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = S3.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Body:   bytes.NewReader(b),
		Bucket: aws.String(os.Getenv("BUCKET")),
		Key:    aws.String(uuid.NewV4().String()),
	})
	return errors.WithStack(err)
}
```
> From [worker.go](../worker.go)

This function simply returns an error to tell Lambda if it was successful or not.

## AWS Config

Next we add the config for our function and the S3 bucket it uses:

```yaml
Resources:
  Bucket:
    Type: AWS::S3::Bucket

  WorkerFunction:
    Properties:
      CodeUri: ./handlers/worker/main.zip
      Environment:
        Variables:
          BUCKET: !Ref Bucket
      Handler: main
      Policies:
        - S3CrudPolicy:
            BucketName: !Ref Bucket
      Runtime: go1.x
      Timeout: 15
    Type: AWS::Serverless::Function
```
> From [template.yml](../template.yml)

Note the longer timeout (15s versus default 3s) we give the worker in case it needs it. Also note the environment variable and policy for the bucket. When we deploy this, AWS will set up the bucket and permissions before creating the Lambda function.

## Go Code

Next, we write another function that cleans up the bucket. This will be called automatically by AWS so it takes a CloudWatch event.

```go
import "github.com/aws/aws-lambda-go/events"

func WorkerPeriodic(ctx context.Context, e events.CloudWatchEvent) error {
	iter := s3manager.NewDeleteListIterator(S3, &s3.ListObjectsInput{
		Bucket: aws.String(os.Getenv("BUCKET")),
	})

	err := s3manager.NewBatchDeleteWithClient(S3).Delete(ctx, iter)
	return errors.WithStack(err)
}
```
> From [worker.go](../worker.go)

## AWS Config

Next we add the config for our function and the S3 bucket it uses:

```yaml
Resources:
  WorkerPeriodicFunction:
    Properties:
      CodeUri: ./handlers/worker-periodic/main.zip
      Environment:
        Variables:
          BUCKET: !Ref Bucket
      Events:
        Request:
          Properties:
            Schedule: rate(1 day)
          Type: Schedule
      Handler: main
      Policies:
        - Statement:
            - Action:
                - s3:DeleteObject
              Effect: Allow
              Resource: !Sub "arn:aws:s3:::${Bucket}/*"
            - Action:
                - s3:ListBucket
              Effect: Allow
              Resource: !Sub "arn:aws:s3:::${Bucket}"
      Runtime: go1.x
      Timeout: 15
    Type: AWS::Serverless::Function
```
> From [template.yml](../template.yml)

Note the `rate(1 day)` ScheduleExpression. We could make this more frequent with `rate(1 minute)` or more specific with `cron(0 12 * * ? *)` (every day at 12). When we deploy this AWS will automatically invoke our function on this schedule. See the [CloudWatch Schedule Expressions guide](https://docs.aws.amazon.com/AmazonCloudWatch/latest/events/ScheduledEvents.html) for more details.

Also note the specific policy. At the time of writing, the simpler `S3CrudPolicy` doesn't actually add a delete permission, so we take matters into our own hands. We aim for the least privilege, so we give our function a single action on the bucket (list), and a single action on its contents (delete). For further reading check out the [per-function policies](docs/per-function-policies.md) doc.

## Package and Deploy

We need to make the boilerplate `worker` and `worker-periodic` Go programs that Lambda will invoke. Check out the the [dev, package, deploy](dev-package-deploy.md) doc for more details.

From here we can assume we have these programs and a single command to deploy:

```console
$ make deploy
cd ./handlers/worker && GOOS=linux go build...
cd ./handlers/worker-periodic && GOOS=linux go build...
aws cloudformation package ...
aws cloudformation deploy ...
```

Finally we can invoke our function manually:

```console
$ aws lambda invoke --function-name gofaas-WorkerFunction \
  --payload '{"time_start": "2018-02-21T15:00:43.511Z"}'  \
  --log-type Tail --output text --query 'LogResult' out.log | base64 -D

START RequestId: 0bb47628-1718-11e8-ad73-c58e72b8826c Version: $LATEST
2018/02/21 15:01:07 Worker Event: {SourceIP: TimeEnd:0001-01-01 00:00:00 +0000 UTC TimeStart:2018-02-21 15:00:43.511 +0000 UTC}
END RequestId: 0bb47628-1718-11e8-ad73-c58e72b8826c
REPORT RequestId: 0bb47628-1718-11e8-ad73-c58e72b8826c  Duration: 11.11 ms  Billed Duration: 100 ms  Memory Size: 128 MB  Max Memory Used: 41 MB
```

And we can review logs to see our periodic function called once a day:

```console
$ aws logs filter-log-events --log-group-name '/aws/lambda/gofaas-WorkerPeriodicFunction' --output text --query 'events[*].{Message:message}'

START RequestId: ae1b5451-1727-11e8-991a-85c308f12bbb Version: $LATEST
2018/02/23 03:01:43 WorkerPeriodic Event: ...
END RequestId: ae1b5451-1727-11e8-991a-85c308f12bbb
REPORT RequestId: ae1b5451-1727-11e8-991a-85c308f12bbb  Duration: 1236.68 ms  Billed Duration: 1300 ms  Memory Size: 128 MB  Max Memory Used: 45 MB

START RequestId: c96123d4-1727-11e8-b0e4-27c53f455614 Version: $LATEST
2018/02/24 03:01:41 WorkerPeriodic Event: ...
END RequestId: c96123d4-1727-11e8-b0e4-27c53f455614
REPORT RequestId: c96123d4-1727-11e8-b0e4-27c53f455614  Duration: 144.81 ms  Billed Duration: 200 ms  Memory Size: 128 MB  Max Memory Used: 46 MB
```

## Summary

When building worker functions we:

- Design custom events
- Configure scheduled events
- Add a storage service to our stack and policies for our functions
- Write Go funcs for the events that perform work

We no longer have to worry about:

- Work queues
- Worker pools
- Running a scheduler process or service

Lambda makes building workers significantly easier.