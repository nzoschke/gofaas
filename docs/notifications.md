# Notifications
### With Go, Lambda and SNS

Traces and logs help us dig into requests, responses and errors, but they don't alert us when an error occurs. For that, we need notifications.

Email and SMS notifications are easy with the Simple Notification Service (SNS).

## Go Code

First we introduce a notification middleware -- a function that takes and returns a function of the same definition. This one takes a handler function, calls it, sends an SNS notification if it returned an error then returns its return values.

```go
type HandlerAPIGateway func(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

func NotifyAPIGateway(h HandlerAPIGateway) HandlerAPIGateway {
	return func(ctx context.Context, e events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		r, err := h(ctx, e)
		notify(ctx, err)
		return r, err
	}
}

func notify(ctx context.Context, err error) {
	topic := os.Getenv("NOTIFICATION_TOPIC")
	if err == nil || topic == "" {
		return
	}

	subj := fmt.Sprintf("ERROR %s", os.Getenv("AWS_LAMBDA_FUNCTION_NAME"))
	msg := fmt.Sprintf("%+v\n", err)
	log.Printf("%s %s\n", subj, msg)

	_, err = SNS().PublishWithContext(ctx, &sns.PublishInput{
		Message:  aws.String(msg),
		Subject:  aws.String(subj),
		TopicArn: aws.String(topic),
	})
}
```
> From [notify.go](../notify.go)

Then we update our handler programs to use the middleware:

```go
func main() {
	lambda.Start(gofaas.NotifyAPIGateway(gofaas.Dashboard))
}
```
> From [handlers/dashboard/main.go](../handlers/dashboard/main.go)

There are a couple Go-isms to note here.

The middleware technique means we don't have to change our function code at all. Handlers can always `return err` and leave it up to the middleware to publish to SNS.

Also notice the use of the `errors` package and `errors.WithStack()` here and throughout our functions. This gives us high quality back traces in our logs and notifications.

The `notify` function would also be a good place to send errors directly to Rollbar, etc.

## AWS Config

Next, we need to add a SNS resources to our config. We create the SNS topic and pass it to every function environment, and we conditionally create SNS subscriptions for an email or phone number parameter. We also give every function the `SNSPublishMessagePolicy` so it can publish to the topic.

```yaml
Conditions:
  NotificationEmailSpecified: !Not [!Equals [!Ref NotificationEmail, ""]]
  NotificationNumberSpecified: !Not [!Equals [!Ref NotificationNumber, ""]]

Globals:
  Function:
    Environment:
      Variables:
        NOTIFICATION_TOPIC: !Ref NotificationTopic

Parameters:
  NotificationEmail:
    Default: ""
    Type: String

  NotificationNumber:
    Default: ""
    Type: String

Resources:
  DashboardFunction:
    Properties:
      Policies:
        - SNSPublishMessagePolicy:
            TopicName: !GetAtt NotificationTopic.TopicName
    Type: AWS::Serverless::Function

  NotificationTopic:
    Properties:
      DisplayName: Notification
      Subscription:
        - !If
          - NotificationEmailSpecified
          - Endpoint: !Ref NotificationEmail
            Protocol: email
          - !Ref AWS::NoValue
        - !If
          - NotificationNumberSpecified
          - Endpoint: !Ref NotificationNumber
            Protocol: sms
          - !Ref AWS::NoValue
      TopicName: Notification
    Type: AWS::SNS::Topic
```
> From [template.yml](../template.yml)

## Summary

Sending notifications with Go and SNS is straightforward:

- Implement a notification middleware for our handlers
- Configure SNS with an email and/or SMS number

We no longer have to worry about:

- Email or SMS gateways
- 3rd party SaaS

AWS services make monitoring our application simple and cost effective.
